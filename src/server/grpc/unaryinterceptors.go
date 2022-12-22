package grpc

import (
	"context"

	"github.com/fatmalabidi/MailBoxServcie/src/config"
	utl "github.com/fatmalabidi/MailBoxServcie/src/utils"
	"github.com/fatmalabidi/MailBoxServcie/src/utils/helpers"
	vld "github.com/fatmalabidi/MailBoxServcie/src/utils/validators"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/status"
)

func UnaryValidationInterceptor(
	ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (interface{}, error) {
	logger := helpers.GetLogger()
	conf, err := config.LoadConfig()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to load configuration")
	}
	switch info.Server.(type) {
	case *SvcHandler:
	case *health.Server:
		logger.Info("health checking ...")
	default:
		return nil, status.Errorf(codes.Unavailable, "the runner type is invalid")
	}
	cid := utl.ExtractCidFromCtx(ctx)
	validCtx, err := handleContext(ctx, conf.Server.Deadline, cid)
	if err != nil {
		logger.Error(err)
		return nil, status.Error(codes.DeadlineExceeded, err.Error())
	}

	newCtx, err := handleToken(validCtx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, err.Error())
	}
	newCtx = utl.EnrichContext(newCtx)

	// check req is not nil or empty
	if ok := vld.IsNilOrEmpty(req); ok {
		logger.Error("nil/empty fields")
		grpcErr := status.Error(
			codes.InvalidArgument,
			"nil/empty fields",
		)
		return nil, grpcErr
	}
	// check if ids are valid
	err = handleID(req, cid)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	return handler(newCtx, req)
}
