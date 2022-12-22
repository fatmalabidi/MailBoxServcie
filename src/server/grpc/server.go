package grpc

import (
	"fmt"
	"net"

	"github.com/fatmalabidi/MailBoxServcie/src/config"
	svc "github.com/fatmalabidi/protobuf/mailboxsvc"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	h "google.golang.org/grpc/health"
	ghv1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Runner struct {
	Config         *config.Config
	Logger         *logrus.Logger
	healthCheckSrv *h.Server
	*grpc.Server
}

// NewGrpcServer create a runner with  its dependencies
func NewGrpcServer(conf *config.Config, logger *logrus.Logger) *Runner {
	healthCheckSrv := h.NewServer()
	server := grpc.NewServer()
	return &Runner{
		Config:         conf,
		Logger:         logger,
		healthCheckSrv: healthCheckSrv,
		Server:         server,
	}
}

// Start starts the server
func (r *Runner) Start(stopCh chan bool) error {
	// create listener
	listener, err := net.Listen(
		"tcp",
		fmt.Sprintf(":%v", r.Config.Server.Port),
	)
	if err != nil {
		return err
	}
	r.registerServices()
	go func() {
		r.Logger.Debug("Starting server...")
		err = r.Server.Serve(listener)
		if err != nil {
			errorMsg := fmt.Sprintf("server can't be started, err: %v", err)
			r.Logger.Error(errorMsg)
			panic(errorMsg)
		}
	}()
	go r.stop(stopCh)
	return nil
}

func (r *Runner) stop(stopCh chan bool) {
	stop := <-stopCh
	if stop {
		r.Logger.Debug("server stopped")
		// It stops the server from accepting new connections
		r.Server.GracefulStop()
		// sets all serving status to NOT_SERVING
		r.healthCheckSrv.Shutdown()
	}
}

// registerServices registers a service and its implementation to the gRPC server
func (r *Runner) registerServices() {
	svcHandler := NewSvcHandler(r.Config, r.Logger)
	svc.RegisterMailboxServiceServer(r.Server, svcHandler)
	ghv1.RegisterHealthServer(r.Server, r.healthCheckSrv)
	// registers the server reflection service on the given gRPC server
	reflection.Register(r.Server)
}
