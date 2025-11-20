package bot

import (
	"context"
	"fmt"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

// SendMessage sends a message to a chat.
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
