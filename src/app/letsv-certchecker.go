package main

import (
    "context"
    "fmt"
    "log"
    "os"

    //"github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/ssm"

    // PGSQL client lib
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgtype"
)


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

/*
func encodeUUID(src [16]byte) string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", src[0:4], src[4:6], src[6:8], src[8:10], src[10:16])
}
*/

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



func main() {
    dbConnectionParams  := getDbConnectionParams()
    dbHandle            := getDbHandle( dbConnectionParams )

    // Make sure handle gets closed when we leave the current function scope
    // (meaning main exits)
    defer dbHandle.Close(context.Background())

    _ = getCertsToRetrieve( dbHandle )


    // Connect to SSM to get PGSQL params
    //ssmClient := createSSMClient()

    // connect to PGSQL

    // Pull list of URL's that need to be checked

    // create channel to pass URL's 

    // Create workers waitgroup

    // Fire off worker goroutines, incrementing waitgroup each time

    // Wait on workgroup

    // At this point, we're last process running, so we can exit cleanly
    fmt.Println("Exiting cleanly")
}
