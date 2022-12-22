package grpcTestUtils

import (
	"context"
	"errors"
	"log"
	"net"

	"github.com/fatmalabidi/MailBoxServcie/src/config"
	"github.com/fatmalabidi/MailBoxServcie/src/server/authenticator"
	srv "github.com/fatmalabidi/MailBoxServcie/src/server/grpc"
	"github.com/fatmalabidi/MailBoxServcie/src/storage"
	"github.com/fatmalabidi/MailBoxServcie/src/utils/helpers"
	"github.com/fatmalabidi/MailBoxServcie/src/utils/validators"
	svc "github.com/fatmalabidi/protobuf/mailboxsvc"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

type TestCommon struct {
	listener   *bufconn.Listener
	conf       *config.Config
	srv        *grpc.Server
	svcHandler *srv.SvcHandler
	logger     *logrus.Logger
}

func StartGRPCServer(logger *logrus.Logger) *TestCommon {
	listener := bufconn.Listen(bufSize)
	conf, _ := config.LoadConfig()
	err := LoadCred(conf)
	if err != nil {
		log.Fatalf("cant load credentials,%v", err)
	}
	runner := srv.NewSvcHandler(conf, logger)
	var opts []grpc.ServerOption
	opts = append(
		opts,
		grpc.UnaryInterceptor(srv.UnaryValidationInterceptor),
		grpc.StreamInterceptor(srv.StreamValidationInterceptor),
	)
	grpcSrv := grpc.NewServer(opts...)
	svc.RegisterMailboxServiceServer(grpcSrv, runner)
	go func(listener *bufconn.Listener) {
		if err := grpcSrv.Serve(listener); err != nil {
			logger.Fatalf("Server exited with error: %v", err)
		}
	}(listener)
	return &TestCommon{
		listener:   listener,
		conf:       conf,
		srv:        grpcSrv,
		svcHandler: runner,
		logger:     logger,
	}
}

// LoadCred function looks for credentials using one of the following paths:
// * directly from environment
// * from storage
func LoadCred(conf *config.Config) error {
	if isCredInEnv() {
		return nil
	}
	return getCredFromStorage(conf)
}

// isCredInEnv function checks if the there is valid credentials in environment
func isCredInEnv() bool {
	username, password := helpers.GetCredFromEnv()
	if password == "" || !validators.IsValidID(username) {
		return false
	}
	return true
}

// loadCredFromStorage function initialise an auth and load credentials from storage
func getCredFromStorage(conf *config.Config) error {
	// create auth
	logger := helpers.GetTestLogger()
	storageHandler, err := storage.Create(&conf.Storage, logger)
	if err != nil {
		return errors.New("Fail to create storage handler: " + err.Error())
	}
	authenticator := authenticator.NewAuthenticator(storageHandler, &conf.Storage)
	// load credentials
	err = authenticator.LoadCredentials(context.Background())
	if err != nil {
		return errors.New("Fail to load credentials: " + err.Error())
	}
	return nil
}

// startClientConnection creates a client and clientConnection to the server
func (common *TestCommon) StartClientConnection() (svc.MailboxServiceClient, *grpc.ClientConn) {

	conn, err := grpc.DialContext(
		context.Background(), "", grpc.WithContextDialer(common.getBufDialer()), grpc.WithInsecure(),
		WithClientUnaryInterceptor(), WithClientStreamInterceptor())

	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}
	newClient := svc.NewMailboxServiceClient(conn)
	return newClient, conn
}

// getBufDialer creates an in-memory full-duplex network connection
func (common *TestCommon) getBufDialer() func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, url string) (net.Conn, error) {
		return common.listener.Dial()
	}
}
