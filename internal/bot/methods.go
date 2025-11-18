package bot

import (
	"context"
	"fmt"

	"github.com/mymmrac/telego"
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
