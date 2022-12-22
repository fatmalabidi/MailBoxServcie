package grpc

import (
	"context"

	"github.com/fatmalabidi/MailBoxServcie/src/config"
	utl "github.com/fatmalabidi/MailBoxServcie/src/utils"
	"github.com/fatmalabidi/MailBoxServcie/src/utils/helpers"
	vld "github.com/fatmalabidi/MailBoxServcie/src/utils/validators"
	"github.com/fatmalabidi/protobuf/common"
	svc "github.com/fatmalabidi/protobuf/mailboxsvc"
	"github.com/sirupsen/logrus"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/grpc/status"
)

type serverStreamWrapper struct {
	ss   grpc.ServerStream
	ctx  context.Context
	conf *config.Config
	srv  *SvcHandler
	cid  string
	*logrus.Logger
}

func (w *serverStreamWrapper) RecvMsg(m interface{}) error {
	if err := w.ss.RecvMsg(m); err != nil {
		return err
	}

	switch r := m.(type) {
	case *svc.GetMessagesByParentReq:
		if vld.IsNilOrEmpty(r) {
			return status.Error(codes.InvalidArgument, "mandatory fields cannot be empty/nil")
		}
		return nil
	case *common.MailBoxID:
		if vld.IsNilOrEmpty(r) {
			return status.Error(codes.InvalidArgument, "mandatory fields cannot be empty/nil")
		}
		return nil
	default:
		return status.Errorf(codes.Unimplemented, "unknown request type: %T", r)
	}

}

func (w *serverStreamWrapper) Context() context.Context        { return w.ctx }
func (w *serverStreamWrapper) SendHeader(md metadata.MD) error { return w.ss.SendHeader(md) }
func (w *serverStreamWrapper) SetHeader(md metadata.MD) error  { return w.ss.SetHeader(md) }
func (w *serverStreamWrapper) SendMsg(msg interface{}) error   { return w.ss.SendMsg(msg) }
func (w *serverStreamWrapper) SetTrailer(md metadata.MD)       { w.ss.SetTrailer(md) }

// StreamValidationInterceptor is the server stream interceptor
// Validates metadata (first chunk) , fileType (second chunk), file size in the chunks left
func StreamValidationInterceptor(
	srv interface{}, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	logger := helpers.GetLogger()
	switch s := srv.(type) {
	case *SvcHandler:
		resp := &serverStreamWrapper{
			ss:     ss,
			ctx:    ss.Context(),
			conf:   s.Config,
			Logger: logger,
			srv:    s,
		}

		correlationID := utl.ExtractCidFromCtx(resp.ctx)
		validCtx, err := handleContext(resp.ctx, s.Config.Server.Deadline, correlationID)
		//set the new ctx and the cid to the response
		resp.ctx = validCtx
		resp.cid = correlationID
		if err != nil {
			logger.Error(err)
			return status.Error(codes.DeadlineExceeded, err.Error())
		}
		newCtx, err := handleToken(validCtx)
		if err != nil {
			return err
		}
		newCtx = utl.EnrichContext(newCtx)
		resp.ctx = newCtx

		return handler(srv, resp)
	case grpc_reflection_v1alpha.ServerReflectionServer:
		return handler(srv, ss)
	default:
		err := status.Error(codes.Code(code.Code_UNAVAILABLE), "runner not defined")
		logger.WithError(err).Error()
		return err
	}
}
