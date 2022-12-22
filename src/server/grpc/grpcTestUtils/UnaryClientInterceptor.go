package grpcTestUtils

import (
	"context"
	"google.golang.org/grpc"
)

// This UnaryInterceptor in being used only by the client test

func WithClientUnaryInterceptor() grpc.DialOption {
	return grpc.WithUnaryInterceptor(clientInterceptor)
}

func clientInterceptor(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	mutex.Lock()
	defer mutex.Unlock()
	newCtx, err := SetTokenToCtx(ctx)
	if err != nil {
		return err
	}
	return invoker(newCtx, method, req, reply, cc, opts...)
}
