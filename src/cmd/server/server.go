package main

import (
	"github.com/fatmalabidi/MailBoxServcie/src/config"
	"github.com/fatmalabidi/MailBoxServcie/src/server"
	hlp "github.com/fatmalabidi/MailBoxServcie/src/utils/helpers"
	"github.com/sirupsen/logrus"
)

var (
	// stop channel is used to stop the server
	stop = make(chan bool)
)

// realMain starts the server
func realMain(conf *config.Config, logger *logrus.Logger) {
	logger.Debug("Starting mailbox service...")
	// create a server
	srv, err := server.Create(conf, logger)
	if err != nil {
		logger.WithError(err).Panic("failed to create the server")
	}
	// start server
	err = srv.Start(stop)
	if err != nil {
		logger.WithError(err).Panic("failed to start the service")
	}
	<-stop
}

func main() {
	logger := hlp.GetLogger()
	// get the configuration
	conf, err := config.LoadConfig()
	if err != nil {
		logger.WithError(err).Fatal("failed to load the configuration")
	}
	logger.Debug("Configuration parsed successfully")
	realMain(conf, logger)
}
