package grpc_test

import (
	"github.com/fatmalabidi/MailBoxServcie/src/config"
	"github.com/fatmalabidi/MailBoxServcie/src/server/grpc"
	"github.com/fatmalabidi/MailBoxServcie/src/utils/helpers"
	"testing"
)

func TestSvcRunner_Start(t *testing.T) {
	conf, _ := config.LoadConfig()
	var logger = helpers.GetTestLogger()
	runner := grpc.NewGrpcServer(conf, logger)
	stopCh := make(chan bool)
	errCh := make(chan error)
	defer close(stopCh)
	go func(errCh chan error) {
		err := runner.Start(stopCh)
		errCh <- err
	}(errCh)
	err := <-errCh
	if err != nil {
		t.Fatalf("expected server to run successfully, got error: %v", err)
	}
}
