package main

import (
    "context"
    "fmt"
    "log"

    //"github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    //"github.com/aws/aws-sdk-go-v2/service/ssm"
)




func main() {
    // Oh shit is it this? Bad C lib? https://github.com/aws/aws-sdk-go/issues/4835#issuecomment-1544383910

    fmt.Println("Starting up")
    myContext := context.TODO()
    //_ = context.TODO()
    //cfg, err := config.LoadDefaultConfig(myContext, config.WithRegion("us-east-2") )
    _, err := config.LoadDefaultConfig(myContext, config.WithRegion("us-east-2") )

    if err != nil {
        log.Fatalf("Unable to load SDK config: %v", err)
    }

    /*
    // Get the SSM client
    //ssmClient := ssm.NewFromConfig(cfg)
    //ssmClient := ssm.NewFromConfig(cfg)
    _ = ssm.NewFromConfig(cfg)

    // Docs for this stuff: https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/ssm#Client.GetParameters

    // Get PGSQL params from the client
    parameterNames := []string{"/letsvalidate/db/aws/us-east-2/dbhost"}

    // Intentionally null pointer -- GetParameters ignores this param for
    // string types, but it has to get passed to compile
    var withDecryption *bool

    getParamsInput := &ssm.GetParametersInput{ 
        Names           : parameterNames, 
        WithDecryption  : withDecryption,
    }
        
    _, err = ssmClient.GetParameters(context.TODO(), getParamsInput)

    if err != nil {
        log.Fatalf("Could not get params")
    }

    fmt.Println("Whoah got a param?!?!?")

    // Connect to SSM to get PGSQL params
    //ssmClient := createSSMClient()

    // connect to PGSQL

    // Pull list of URL's that need to be checked

    // create channel to pass URL's 

    // Create workers waitgroup

    // Fire off worker goroutines, incrementing waitgroup each time

    // Wait on workgroup

    */
    // At this point, we're last process running, so we can exit cleanly
    fmt.Println("Exiting cleanly")
}
