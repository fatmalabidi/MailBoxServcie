package grpcTestUtils

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/fatmalabidi/MailBoxServcie/src/config"
	"github.com/fatmalabidi/MailBoxServcie/src/utils"
	"github.com/fatmalabidi/MailBoxServcie/src/utils/helpers"
	cacheConf "github.com/fatmalabidi/Service-Cache/config"
	"github.com/fatmalabidi/Service-Gateway/src/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var mutex sync.RWMutex

const (
	COGNITO_USERNAME          = "USERNAME"
	COGNITO_PASSWORD          = "PASSWORD"
	USER_PASSWORD_AUTH string = "USER_PASSWORD_AUTH"
)

func SetTokenToCtx(ctx context.Context) (context.Context, error) {
	userID, password := helpers.GetCredFromEnv()
	return SetTokenToCtxByAuth(ctx, userID, password)
}

func getTokenByAuth(cache cacheConf.Config, userID string, password string) (string, error) {
	authParam := make(map[string]*string)
	authParam[COGNITO_USERNAME] = aws.String(userID)
	authParam[COGNITO_PASSWORD] = aws.String(password)
	input := cognitoidentityprovider.InitiateAuthInput{
		AnalyticsMetadata: nil,
		AuthFlow:          aws.String(USER_PASSWORD_AUTH),
		AuthParameters:    authParam,
		ClientId:          aws.String(cache.Cognito.ClientID),
	}
	output, err := cache.Cognito.Client.InitiateAuth(&input)
	if err != nil {
		return "", err
	}
	token := *output.AuthenticationResult.IdToken
	return token, nil
}

func SetTokenToCtxByAuth(ctx context.Context, userID string, password string) (context.Context, error) {
	conf, _ := config.LoadConfig()
	confCognito := client.Cognito{
		UserPoolID: conf.Cognito.UserPoolID,
		ClientID:   conf.Cognito.ClientID,
		Region:     conf.Cognito.Region,
	}
	cognitoClient := client.GetCognitoClient(confCognito)

	cacheClient := cacheConf.Config{
		Cognito:    *cognitoClient,
		MaxTimeOut: conf.Server.Deadline,
	}

	token, err := getTokenByAuth(cacheClient, userID, password)
	if err != nil {
		return ctx, status.Errorf(codes.Unauthenticated, err.Error())
	}
	md := metadata.Pairs(string(utils.KeyAuthorization), token)
	ctx = metadata.NewOutgoingContext(ctx, md)
	return ctx, nil
}
