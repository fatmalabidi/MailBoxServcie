package server

import (
	"errors"

	"github.com/fatmalabidi/MailBoxServcie/src/config"
	"github.com/fatmalabidi/MailBoxServcie/src/server/grpc"
	"github.com/sirupsen/logrus"
)

type RunnerHandler interface {
	Start(stop chan bool) error
}

// Create creates a Runner based on the type in the config
func Create(conf *config.Config, logger *logrus.Logger) (RunnerHandler, error) {
	switch conf.Server.Type {
	case "grpc":
		return grpc.NewGrpcServer(conf, logger), nil
	default:
		return nil, errors.New("invalid server type")
	}
}
