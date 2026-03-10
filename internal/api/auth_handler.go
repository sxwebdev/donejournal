package api

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"connectrpc.com/connect"
	authv1 "github.com/sxwebdev/donejournal/api/gen/go/donejournal/auth/v1"
	"github.com/sxwebdev/donejournal/internal/config"
	"github.com/sxwebdev/tokenmanager"
	"github.com/tkcrm/mx/logger"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthHandler struct {
	logger   logger.Logger
	config   *config.Config
	tokenMgr *tokenmanager.Manager[TokenData]
}

func NewAuthHandler(l logger.Logger, conf *config.Config, tokenMgr *tokenmanager.Manager[TokenData]) *AuthHandler {
	return &AuthHandler{
		logger:   l,
		config:   conf,
		tokenMgr: tokenMgr,
	}
}

func (h *AuthHandler) LoginWithTelegram(ctx context.Context, req *connect.Request[authv1.LoginWithTelegramRequest]) (*connect.Response[authv1.LoginWithTelegramResponse], error) {
	td := req.Msg.GetTelegramData()
	if td == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("telegram_data is required"))
	}

	// Validate the hash
	if err := validateTelegramAuth(td, h.config.Telegram.BotToken); err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	// Check auth_date is not too old (5 minutes)
	authTime := time.Unix(td.AuthDate, 0)
	if time.Since(authTime) > 5*time.Minute {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("auth data expired"))
	}

	userID := strconv.FormatInt(td.Id, 10)
	additionalData := TokenData{
		TelegramID: td.Id,
		FirstName:  td.FirstName,
		LastName:   td.LastName,
		Username:   td.Username,
		PhotoURL:   td.PhotoUrl,
	}

	token, tokenData, err := h.tokenMgr.CreateToken(ctx, userID, additionalData, tokenmanager.AccessTokenType)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create token: %w", err))
	}

	return connect.NewResponse(&authv1.LoginWithTelegramResponse{
		AccessToken: token,
		ExpiresAt:   timestamppb.New(tokenData.Expiry),
		User:        tokenDataToUserProto(additionalData),
	}), nil
}

func (h *AuthHandler) GetCurrentUser(ctx context.Context, _ *connect.Request[authv1.GetCurrentUserRequest]) (*connect.Response[authv1.GetCurrentUserResponse], error) {
	data, err := tokenDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&authv1.GetCurrentUserResponse{
		User: tokenDataToUserProto(data.AdditionalData),
	}), nil
}

func (h *AuthHandler) Logout(ctx context.Context, _ *connect.Request[authv1.LogoutRequest]) (*connect.Response[authv1.LogoutResponse], error) {
	token := tokenFromContext(ctx)
	if token != "" {
		if err := h.tokenMgr.RevokeToken(ctx, token); err != nil {
			h.logger.Warnf("failed to revoke token: %v", err)
		}
	}
	return connect.NewResponse(&authv1.LogoutResponse{}), nil
}

// validateTelegramAuth validates the Telegram Login Widget data hash.
// See: https://core.telegram.org/widgets/login#checking-authorization
func validateTelegramAuth(td *authv1.TelegramLoginData, botToken string) error {
	// Build data-check-string
	pairs := []string{
		fmt.Sprintf("auth_date=%d", td.AuthDate),
		fmt.Sprintf("first_name=%s", td.FirstName),
		fmt.Sprintf("id=%d", td.Id),
	}

	if td.LastName != "" {
		pairs = append(pairs, fmt.Sprintf("last_name=%s", td.LastName))
	}
	if td.PhotoUrl != "" {
		pairs = append(pairs, fmt.Sprintf("photo_url=%s", td.PhotoUrl))
	}
	if td.Username != "" {
		pairs = append(pairs, fmt.Sprintf("username=%s", td.Username))
	}

	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	// Secret key = SHA256(bot_token)
	secretKey := sha256.Sum256([]byte(botToken))

	// HMAC-SHA-256(data-check-string, secret_key)
	mac := hmac.New(sha256.New, secretKey[:])
	mac.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	if expectedHash != td.Hash {
		return fmt.Errorf("invalid telegram auth hash")
	}

	return nil
}
