package api

import (
	"context"
	"fmt"
	"strconv"

	"connectrpc.com/connect"
	"github.com/sxwebdev/tokenmanager"
)

type contextKey string

const (
	contextKeyTokenData contextKey = "token_data"
	contextKeyToken     contextKey = "token"
)

// TokenData stores additional data in the token.
type TokenData struct {
	TelegramID int64  `json:"telegram_id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Username   string `json:"username"`
	PhotoURL   string `json:"photo_url"`
}

func contextWithTokenData(ctx context.Context, data *tokenmanager.Data[TokenData]) context.Context {
	ctx = context.WithValue(ctx, contextKeyTokenData, data)
	return ctx
}

func tokenDataFromContext(ctx context.Context) (*tokenmanager.Data[TokenData], error) {
	data, ok := ctx.Value(contextKeyTokenData).(*tokenmanager.Data[TokenData])
	if !ok || data == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("unauthorized"))
	}
	return data, nil
}

func userIDFromContext(ctx context.Context) (int64, error) {
	data, err := tokenDataFromContext(ctx)
	if err != nil {
		return 0, err
	}
	userID, err := strconv.ParseInt(data.UserID, 10, 64)
	if err != nil {
		return 0, connect.NewError(connect.CodeInternal, fmt.Errorf("invalid user ID in token"))
	}
	return userID, nil
}

func contextWithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, contextKeyToken, token)
}

func tokenFromContext(ctx context.Context) string {
	token, _ := ctx.Value(contextKeyToken).(string)
	return token
}
