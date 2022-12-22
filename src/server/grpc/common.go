package grpc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fatmalabidi/MailBoxServcie/src/config"
	utl "github.com/fatmalabidi/MailBoxServcie/src/utils"
	vld "github.com/fatmalabidi/MailBoxServcie/src/utils/validators"
	cacheConf "github.com/fatmalabidi/Service-Cache/config"
	"github.com/fatmalabidi/Service-Cache/model"
	"github.com/fatmalabidi/Service-Cache/pkg/cache"
	"github.com/fatmalabidi/Service-Gateway/src/client"
	"github.com/fatmalabidi/Service-Gateway/src/gateway"
	"github.com/fatmalabidi/Service-Gateway/src/gateway/cachegw/extractUserRunnable"
	"github.com/fatmalabidi/Service-Gateway/src/gateway/cachegw/validateTokenRunnable"
	"github.com/fatmalabidi/protobuf/common"
	svc "github.com/fatmalabidi/protobuf/mailboxsvc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ValidationHandler = func(ctx context.Context, conf *config.Config) error

// handleContext checks context correlation Id and timeout and sets default timeout if not specified
func handleContext(
	ctx context.Context, maxTimeout time.Duration, cid string,
) (context.Context, error) {
	ctx = utl.BuildContext(ctx,
		utl.WithValue(utl.KeyCid, cid))
	// check timeout and set default timeout if not specified
	if _, ok := ctx.Deadline(); !ok {
		defaultCtx, _ := utl.AddTimeoutToCtx(ctx, int64(maxTimeout))
		return defaultCtx, nil
	}
	if deadline, _ := ctx.Deadline(); deadline.Sub(time.Now()) > maxTimeout*time.Second {
		return nil, errors.New("context deadline exceeded ")
	}
	return ctx, nil
}

// handleID checks all IDs in the request and returns an error if one of them is invalid
func handleID(req interface{}, cid string) error {
	var ok bool
	switch req := req.(type) {
	case *svc.SendMessageReq:
		ok = vld.IsValidID(req.MailboxID.Value) && vld.IsValidID(req.MessageInfo.SenderID.Value) && vld.IsValidID(req.MessageInfo.RecipientID.Value)
	case *svc.GetMessageReq:
		ok = vld.IsValidID(req.MailboxID.Value) && vld.IsValidID(req.MessageID.Value)
	case *common.MailBoxID:
		ok = vld.IsValidID(req.Value)
	case *svc.DeleteMessagesReq:
		ok = vld.IsValidID(req.MailboxID.Value) && vld.IsValidListIDs(req.MessageIDs)
	case *svc.DeleteMessageReq:
		ok = vld.IsValidID(req.MailboxID.Value) && vld.IsValidID(req.MessageID.Value)
	case *svc.GetMessagesByParentReq:
		ok = vld.IsValidID(req.MailboxID.Value) && vld.IsValidID(req.ParentMessageID.Value)
	case *grpc_health_v1.HealthCheckRequest:
		ok = true
	default:
		return status.Error(
			codes.Unimplemented,
			fmt.Sprintf("unknown request type %T", req),
		)
	}
	if !ok {
		return status.Error(
			codes.InvalidArgument,
			fmt.Sprintf("invalid ID %s", cid),
		)
	}

	return nil
}

// handleToken extracts the token from the context and validate it,
// if the token is valid so it adds the new data from the token to the context
func handleToken(ctx context.Context) (context.Context, error) {
	// 1. extract the token from the context
	token, err := getToken(ctx)
	if err != nil {
		return nil, err
	}
	// 2. validate the token with cognito using the gateway as a proxy
	conf, _ := config.LoadConfig()
	err = validateToken(ctx, conf, token)
	if err != nil {
		return nil, err
	}
	return ctx, nil
}

// validateToken checks if the token is valid then extracts the userInfo from it, otherwise it returns an error
func validateToken(ctx context.Context, conf *config.Config, token string) error {
	// 2.1 prepare the cognito client
	cognitoConf := client.Cognito{
		UserPoolID: conf.Cognito.UserPoolID,
		ClientID:   conf.Cognito.ClientID,
		Region:     conf.Cognito.Region,
	}
	cognitoClient := client.GetCognitoClient(cognitoConf)
	confCache := cacheConf.Config{
		Cognito:    *cognitoClient,
		MaxTimeOut: conf.Server.Deadline,
	}
	// 2.2 create the cashHandler that will save the data in redis
	cacheHandler := cache.Create(&confCache)

	// validateToken is of type [ValidateToken], it wraps the validation function that will be executed by the gateway
	validateToken := &validateTokenRunnable.ValidateToken{
		Input: token,
		Exec:  cacheHandler.ValidateToken,
	}
	// gw that will execute the validation function
	gw := &gateway.Gateway{}
	_, err := gw.Execute(ctx, validateToken)
	if err != nil {
		return err
	}

	err = extractUserInfo(ctx, token, cacheHandler, gw)
	if err != nil {
		return err
	}
	return nil
}

func extractUserInfo(ctx context.Context, token string, cacheHandler cache.Handler, gw *gateway.Gateway) error {
	extractUserInfo := extractUserRunnable.ExtractUserInfo{
		Input: token,
		Exec:  cacheHandler.ExtractUserInfo,
	}
	userInfo, err := gw.Execute(ctx, &extractUserInfo)
	if err != nil {
		return err
	}
	_, valid := userInfo.(model.User)
	if !valid {
		return errors.New("invalid token")
	}
	return nil
}

func getToken(validCtx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(validCtx)
	if !ok {
		return "", errors.New("missing authorization token")
	}
	authHeader, ok := md[string(utl.KeyAuthorization)]
	if !ok {
		return "", errors.New("authorization token is not supplied")
	}
	token := authHeader[0]
	return token, nil
}
