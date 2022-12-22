package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/fatmalabidi/MailBoxServcie/lambda/dynamo/mailbox/duplicator/mailboxHandler"
	"github.com/sirupsen/logrus"
)

func main() {
	awsSession := session.Must(session.NewSessionWithOptions(
		session.Options{Config: aws.Config{
			MaxRetries: aws.Int(5),
		},
			SharedConfigState: session.SharedConfigEnable,
		}))
	db := dynamodb.New(awsSession)
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	Handler := mailboxHandler.Handler{
		Db:     db,
		Logger: logger,
	}
	lambda.Start(Handler.Handle)
}
