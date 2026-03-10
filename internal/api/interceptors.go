package api

import (
	"context"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/sxwebdev/donejournal/api/gen/go/donejournal/auth/v1/authv1connect"
	"github.com/sxwebdev/tokenmanager"
)

type authInterceptorImpl struct {
	tokenMgr *tokenmanager.Manager[TokenData]
}

func newAuthInterceptor(tokenMgr *tokenmanager.Manager[TokenData]) connect.Interceptor {
	return &authInterceptorImpl{tokenMgr: tokenMgr}
}

func (i *authInterceptorImpl) authenticate(ctx context.Context, header http.Header, procedure string) (context.Context, error) {
	// Skip auth for login
	if procedure == authv1connect.AuthServiceLoginWithTelegramProcedure {
		return ctx, nil
	}

	authHeader := header.Get("Authorization")
	if authHeader == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	data, ok := i.tokenMgr.ValidateToken(ctx, token, tokenmanager.AccessTokenType)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	ctx = contextWithTokenData(ctx, data)
	ctx = contextWithToken(ctx, token)
	return ctx, nil
}

func (i *authInterceptorImpl) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		ctx, err := i.authenticate(ctx, req.Header(), req.Spec().Procedure)
		if err != nil {
			return nil, err
		}
		return next(ctx, req)
	}
}

func (i *authInterceptorImpl) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (i *authInterceptorImpl) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		ctx, err := i.authenticate(ctx, conn.RequestHeader(), conn.Spec().Procedure)
		if err != nil {
			return err
		}
		return next(ctx, conn)
	}
}
