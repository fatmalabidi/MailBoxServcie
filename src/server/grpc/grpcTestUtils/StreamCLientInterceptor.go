package grpcTestUtils

import (
	"context"
	"google.golang.org/grpc"
)

// This StreamInterceptor in being used only by the client test

func WithClientStreamInterceptor() grpc.DialOption {
	return grpc.WithStreamInterceptor(clientStreamInterceptor)
}

func clientStreamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	mutex.Lock()
	defer mutex.Unlock()
	newCtx, err := SetTokenToCtx(ctx)
	if err != nil {
		return nil, err
	}
	return streamer(newCtx, desc, cc, method, opts...)
}
