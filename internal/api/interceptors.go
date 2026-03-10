package api

import (
	"context"
	"strings"

	"connectrpc.com/connect"
	"github.com/sxwebdev/donejournal/api/gen/go/donejournal/auth/v1/authv1connect"
	"github.com/sxwebdev/tokenmanager"
)

func newAuthInterceptor(tokenMgr *tokenmanager.Manager[TokenData]) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			// Skip auth for login
			if req.Spec().Procedure == authv1connect.AuthServiceLoginWithTelegramProcedure {
				return next(ctx, req)
			}

			authHeader := req.Header().Get("Authorization")
			if authHeader == "" {
				return nil, connect.NewError(connect.CodeUnauthenticated, nil)
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == authHeader {
				return nil, connect.NewError(connect.CodeUnauthenticated, nil)
			}

			data, ok := tokenMgr.ValidateToken(ctx, token, tokenmanager.AccessTokenType)
			if !ok {
				return nil, connect.NewError(connect.CodeUnauthenticated, nil)
			}

			ctx = contextWithTokenData(ctx, data)
			ctx = contextWithToken(ctx, token)

			return next(ctx, req)
		}
	}
}
