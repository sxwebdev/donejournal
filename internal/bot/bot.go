package bot

import (
	"context"
	"fmt"

	"github.com/mymmrac/telego"
	"github.com/tkcrm/mx/logger"
)

const serviceName = "telegram-bot"

type Bot struct {
	logger logger.Logger
	bot    *telego.Bot

	updates chan telego.Update
}

func New(l logger.Logger, botToken string) (*Bot, error) {
	bot, err := telego.NewBot(botToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot instance: %w", err)
	}

	return &Bot{
		logger:  logger.With(l, "service", serviceName),
		bot:     bot,
		updates: make(chan telego.Update, 500),
	}, nil
}

// Name returns the service name.
func (b *Bot) Name() string {
	return serviceName
}

// Start starts the bot.
func (b *Bot) Start(ctx context.Context) error {
	res, err := b.bot.GetMe(ctx)
	if err != nil {
		return fmt.Errorf("failed to get bot info: %w", err)
	}

	b.logger.Infof("bot started: @%s (ID: %d)", res.Username, res.ID)

	updates, err := b.bot.UpdatesViaLongPolling(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start receiving updates: %w", err)
	}

	// Loop through all updates when they came
	go func() {
		for update := range updates {
			// Log different types of updates
			if update.Message != nil {
				b.logger.Debugw(
					"received message",
					"from_id", update.Message.From.ID,
					"from_username", update.Message.From.Username,
					"text", update.Message.Text,
				)
			} else if update.CallbackQuery != nil {
				b.logger.Debugw(
					"received callback query",
					"from_id", update.CallbackQuery.From.ID,
					"from_username", update.CallbackQuery.From.Username,
					"data", update.CallbackQuery.Data,
				)
			}

			b.updates <- update
		}
	}()

	return nil
}

// Stop stops the bot.
func (b *Bot) Stop(ctx context.Context) error {
	return nil
}

// Api returns the underlying telego Bot instance.
func (b *Bot) API() *telego.Bot { return b.bot }

// OnUpdate registers a handler for incoming updates.
func (b *Bot) OnUpdate() <-chan telego.Update {
	return b.updates
}
