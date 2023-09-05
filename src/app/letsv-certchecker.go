package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "runtime"
    "sync"
    "time"

    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/ssm"

    // PGSQL client lib
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgtype"
)

type channelUrlInfo struct {
    urlId   string
    url     string
}


func getDbConnectionParams() map[string]string {
    cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-2") )

    if err != nil {
        log.Fatalf("Unable to load SDK config: %v", err)
        os.Exit(1)
    }

    // Get the SSM client
    ssmClient := ssm.NewFromConfig(cfg)

    // Docs for this stuff: https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/ssm#Client.GetParameters

    // Get PGSQL params from the client
    parameterNames := []string{
        "/letsvalidate/db/aws/us-east-2/dbhost",
        "/letsvalidate/db/aws/us-east-2/dbname",
        "/letsvalidate/db/aws/us-east-2/dbpassword",
        "/letsvalidate/db/aws/us-east-2/dbuser",
    }

    // Intentionally null pointer -- GetParameters ignores this param for
    // string types, but it has to get passed to compile
    var withDecryption *bool

    getParamsInput := &ssm.GetParametersInput{
        Names           : parameterNames,
        WithDecryption  : withDecryption,
    }

    paramOutput, err := ssmClient.GetParameters(context.TODO(), getParamsInput)

    if err != nil {
        log.Fatalf("Could not get params")
        os.Exit(1)
    }

    dbParams := make(map[string]string)
    dbParamsKeys := []string{ "host", "dbname", "password", "user" }
    for idx, currParam := range paramOutput.Parameters {
        dbParams[dbParamsKeys[idx]] = *currParam.Value
    }

    return dbParams
}

func getDbHandle( dbConnectionParams map[string]string ) *pgx.Conn {
    // https://pkg.go.dev/github.com/jackc/pgx/v5#Connect
    connectionString := fmt.Sprintf("host=%s user=%s password=%s dbname=%s", 
        dbConnectionParams["host"], 
        dbConnectionParams["user"],
        dbConnectionParams["password"],
        dbConnectionParams["dbname"] )
    conn, err := pgx.Connect( context.Background(), connectionString )

    if err != nil {
        log.Fatalf("Bombed out in DB connection: %v", err )
        os.Exit(1)
    }

    return conn
}


func getCertsToRetrieve( dbHandle *pgx.Conn ) map[string]string {
    collectedUrlInfo := make(map[string]string)
    returnedRows, err := dbHandle.Query(context.Background(), 
        "SELECT url_id_pk, url FROM urls WHERE cert_retrieved < current_date ORDER BY cert_retrieved;" )

    if err != nil {
        log.Fatalf("Error hit when pulling URL rows: %v", err )
        os.Exit( 1 )
    }

    // https://pkg.go.dev/github.com/jackc/pgx/v5#Rows

    // Next tells us if there's more data to ready 
    for returnedRows.Next() == true {
        rowValues, err := returnedRows.Values()

        if err != nil {
            log.Fatalf("Error reading values but next returned true")
            os.Exit(1)
        }

        // Use type assertions to force the values in the returned array from
        // "any" to actual strings
        currUrlIdBytes  := rowValues[0].([16]byte)
        currUrl         := rowValues[1].(string)

        // Wow was that a hard type to work with.
        //      https://github.com/jackc/pgx/blob/v5.4.3/pgtype/uuid.go
        //
        //      Turns out we can't call pgtype.encodeUUID directly because it
        //      starts with a lowercase letter. That's tripped me up twice.
        //      
        //      Forced to bounce through pgtype.UUID and do all of the Value()
        //      and type assertions BS to get out the string representation of
        //      the GUID.
        urlIdUuid := pgtype.UUID{ Bytes: currUrlIdBytes, Valid: true }

        // Use a type assertion to get it out of a driver.Value into a string
        currUrlIdDriverValue, err := urlIdUuid.Value()

        if err != nil {
            log.Fatalf("Could not get value out of UUID bytes: %v\n", err )
            os.Exit(1)
        }

        // Need to use type assertion to get back to string
        currUrlIdString := currUrlIdDriverValue.(string)

        collectedUrlInfo[currUrlIdString] = currUrl

        fmt.Printf("Id = %s, url = %s\n", currUrlIdString, currUrl )
    }
    
    // Have to close the rows object to make the connection usable again
    returnedRows.Close()

    return collectedUrlInfo
}


func urlWorkerEntryPoint( dbConnectionParams map[string]string, urlChannel chan channelUrlInfo, wg *sync.WaitGroup ) {
    // Make sure we note we're done on the way out
    defer wg.Done()

    // Create our goroutine's individual DB connection
    //dbHandle := getDbHandle( dbConnectionParams )
    _ = getDbHandle( dbConnectionParams )

    timedOut := false
    
    // Do a read with timeout on the channel for new URL info
    for timedOut == false {
        select {
        case urlToCheck := <- urlChannel:
            fmt.Printf("Worker got URL to test with ID %s and URL %s\n", urlToCheck.urlId, urlToCheck.url )



            // Do a very short sleep but it's a point to hand off CPU resources
            // to other goroutines that are ready to do some processing
            time.Sleep( 25 * time.Millisecond )

        case <- time.After(1 * time.Second):
            // We timed out, note that we want to bail from the loop
            timedOut = true
        }
    }

    // The defer will make sure we call done on the waitgroup on the way out
}

func pullCertsAndWriteToPgsqlWorkersKv( dbConnectionParams map[string]string, certsToCheck map[string]string ) {
    // create channel to pass URL's
    uriDataChannel := make(chan channelUrlInfo)

    // Create workers waitgroup
    wg := &sync.WaitGroup{}

    numberOfWorkerProcesses := runtime.NumCPU() * 8


    // Fire off worker goroutines that read form the channel, incrementing waitgroup each time
    for i := 0; i < numberOfWorkerProcesses; i++ {
        go urlWorkerEntryPoint( dbConnectionParams, uriDataChannel, wg )
        
        // Add to the waitgroup so we know how many goroutines need to finish
        // before Wait() will return
        wg.Add(1)
    }

    // Write all the URL info into the channel for the workers to Do Their
    // Thing(tm)
    for k, v := range certsToCheck {
        currUrlInfo := channelUrlInfo{
            urlId   : k,
            url     : v,
        }

        uriDataChannel <- currUrlInfo
    }


    // Wait on workgroup
    wg.Wait()

    fmt.Println("All child worker goroutines have returned cleanly" )
}



func main() {
    fmt.Println("letsv-cert-checker starting")

    dbConnectionParams  := getDbConnectionParams()
    dbHandle            := getDbHandle( dbConnectionParams )

    // Make sure handle gets closed when we leave the current function scope
    // (meaning main exits)
    defer dbHandle.Close(context.Background())

    certsToCheck := getCertsToRetrieve( dbHandle )

    pullCertsAndWriteToPgsqlWorkersKv( dbConnectionParams, certsToCheck )

    fmt.Println("letsv-cert-checked exiting cleanly")
}
