package utils

import (
	"context"
	"time"

	"github.com/dgrijalva/jwt-go"
	auth "github.com/fatmalabidi/Service-Cache/pkg/cache/auth/token"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

type contextKey string

// const related to context metadata keys
const (
	KeyCid           = contextKey("correlation_id")
	KeyAuthorization = contextKey("authorization")
	KeyRequester     = contextKey("cognito:username")
	KeyEmail         = contextKey("email")
	KeyPhoneNumber   = contextKey("phone_number")
)

type AddValueToCtxFn func(ctx context.Context) context.Context

// BuildContext takes a context and a list of functions to be applied to this context and return back the new xontext
func BuildContext(ctx context.Context, valueCtxFns ...AddValueToCtxFn) context.Context {
	for _, f := range valueCtxFns {
		ctx = f(ctx)
	}
	return ctx
}

// WithValue   returns a function that add a value to the context with the given key
func WithValue(key, value interface{}) AddValueToCtxFn {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, key, value)
	}
}

// AddTimeoutToCtx add timeout to context.
func AddTimeoutToCtx(ctx context.Context, timeOutDuration int64) (context.Context, context.CancelFunc) {
	timeOut := time.Duration(timeOutDuration) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeOut)
	return ctx, cancel
}

func ExtractCidFromCtx(ctx context.Context) string {
	cid := uuid.New().String()
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return cid
	}
	authHeader, ok := md[string(KeyCid)]
	if !ok {
		return cid
	}
	cid = authHeader[0]
	return cid
}

// EnrichContext function extracts the keys' values in the auth and flatten them in the context values before returning it
func EnrichContext(ctx context.Context) context.Context {
	md, _ := metadata.FromIncomingContext(ctx)
	// get CID
	cidHeader, ok := md[string(KeyCid)]
	// if it does exist, set it
	if ok {
		ctx = context.WithValue(ctx, KeyCid, cidHeader[0])
		// if it doesnt, set new one
	} else {
		ctx = context.WithValue(ctx, KeyCid, uuid.New().String())
	}
	token := md[string(KeyAuthorization)]
	validToken, _ := auth.IsValidToken(token[0])
	mapClaims, ok := validToken.Claims.(jwt.MapClaims)
	if !ok {
		return ctx
	}
	keys := []contextKey{
		KeyRequester,
		KeyPhoneNumber,
		KeyEmail,
	}
	for _, key := range keys {
		value, ok := mapClaims[string(key)]
		if ok {
			ctx = context.WithValue(ctx, key, value)
		}
	}
	return ctx
}
