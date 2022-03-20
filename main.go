package main

import (
	"context"
	"log"
	"os"

	// "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	// "github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func main() {
	awsCfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Couldn't configure aws: %v", err)
	}
	log.Printf("Got awsCfg: %#v", awsCfg)
	log.Printf("Got environment: %#v", os.Environ())
}
