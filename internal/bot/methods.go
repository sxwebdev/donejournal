package bot

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

// DownloadFile downloads a file from Telegram by file ID and returns its bytes.
func (b *Bot) DownloadFile(ctx context.Context, fileID string) ([]byte, error) {
	if fileID == "" {
		return nil, fmt.Errorf("fileID is required")
	}

	file, err := b.bot.GetFile(ctx, &telego.GetFileParams{FileID: fileID})
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	fileURL := b.bot.FileDownloadURL(file.FilePath)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code when downloading file: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file body: %w", err)
	}

	return data, nil
}

// SendMessage sends a message to a chat.
// It tries Markdown parse mode first; on parse error it falls back to plain text.
func (b *Bot) SendMessage(ctx context.Context, chatID int64, text string) error {
	if chatID == 0 {
		return fmt.Errorf("chatID is required")
	}

	if text == "" {
		return fmt.Errorf("text is required")
	}

	_, err := b.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID:    telego.ChatID{ID: chatID},
		Text:      text,
		ParseMode: "markdown",
	})
	if err != nil {
		// If Telegram can't parse the markdown, retry without parse mode
		if strings.Contains(err.Error(), "can't parse entities") {
			_, err = b.bot.SendMessage(ctx, &telego.SendMessageParams{
				ChatID: telego.ChatID{ID: chatID},
				Text:   text,
			})
			if err != nil {
				return fmt.Errorf("failed to send message: %w", err)
			}
			return nil
		}
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// SendMessageWithButtons sends a message with inline keyboard buttons.
func (b *Bot) SendMessageWithButtons(ctx context.Context, chatID int64, text string, buttons [][]telego.InlineKeyboardButton) error {
	if chatID == 0 {
		return fmt.Errorf("chatID is required")
	}

	if text == "" {
		return fmt.Errorf("text is required")
	}

	keyboard := tu.InlineKeyboard(buttons...)

	_, err := b.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID:      telego.ChatID{ID: chatID},
		Text:        text,
		ParseMode:   "markdown",
		ReplyMarkup: keyboard,
	})
	if err != nil {
		return fmt.Errorf("failed to send message with buttons: %w", err)
	}

	return nil
}

// AnswerCallbackQuery answers a callback query.
func (b *Bot) AnswerCallbackQuery(ctx context.Context, callbackQueryID string, text string) error {
	if callbackQueryID == "" {
		return fmt.Errorf("callbackQueryID is required")
	}

	err := b.bot.AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
		CallbackQueryID: callbackQueryID,
		Text:            text,
	})
	if err != nil {
		return fmt.Errorf("failed to answer callback query: %w", err)
	}

	return nil
}

// SendChatAction sends a chat action (e.g. "typing") to indicate the bot is processing.
func (b *Bot) SendChatAction(ctx context.Context, chatID int64, action string) error {
	return b.bot.SendChatAction(ctx, &telego.SendChatActionParams{
		ChatID: telego.ChatID{ID: chatID},
		Action: action,
	})
}

// EditMessageText edits the text of a message.
func (b *Bot) EditMessageText(ctx context.Context, chatID int64, messageID int, text string, buttons [][]telego.InlineKeyboardButton) error {
	if chatID == 0 {
		return fmt.Errorf("chatID is required")
	}

	if messageID == 0 {
		return fmt.Errorf("messageID is required")
	}

	if text == "" {
		return fmt.Errorf("text is required")
	}

	var keyboard *telego.InlineKeyboardMarkup
	if len(buttons) > 0 {
		keyboard = tu.InlineKeyboard(buttons...)
	}

	_, err := b.bot.EditMessageText(ctx, &telego.EditMessageTextParams{
		ChatID:      telego.ChatID{ID: chatID},
		MessageID:   messageID,
		Text:        text,
		ParseMode:   "markdown",
		ReplyMarkup: keyboard,
	})
	if err != nil {
		return fmt.Errorf("failed to edit message text: %w", err)
	}

	return nil
}
