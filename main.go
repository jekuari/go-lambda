package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

type MyEvent struct {
	Name string `json:"name"`
}

func HandleRequest(ctx context.Context, event *MyEvent) (*string, error) {
	credentials := credentials.NewEnvCredentials()
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials,
		Region:      aws.String("us-east-1"),
	})

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(sess)

	return nil, nil
}

func main() {
	lambda.Start(HandleRequest)
}
