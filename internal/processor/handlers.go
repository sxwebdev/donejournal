package processor

import (
	"context"
	"fmt"
	"strings"

	"github.com/dromara/carbon/v2"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/sxwebdev/donejournal/internal/models"
	"github.com/sxwebdev/donejournal/internal/store/repos/repo_todos"
)

// HandleCommand handles bot commands like /start, /menu, etc.
func (s *Processor) HandleCommand(ctx context.Context, chatID int64, command string) error {
	switch command {
	case "/start", "/menu":
		return s.showMainMenu(ctx, chatID)
	default:
		return nil
	}
}

// HandleCallbackQuery handles callback queries from inline buttons.
func (s *Processor) HandleCallbackQuery(ctx context.Context, query *telego.CallbackQuery) error {
	chatID := query.Message.GetChat().ID
	messageID := query.Message.GetMessageID()
	callbackData := query.Data
	userID := query.From.ID

	// Answer the callback query first
	if err := s.botService.AnswerCallbackQuery(ctx, query.ID, ""); err != nil {
		s.logger.Errorf("failed to answer callback query: %v", err)
	}

	switch callbackData {
	case "show_todos":
		return s.showTodosMenu(ctx, chatID, messageID)
	case "show_done":
		return s.showDoneMenu(ctx, chatID, messageID)
	case "todos_today":
		return s.showTodosForDay(ctx, chatID, messageID, userID, "today")
	case "todos_tomorrow":
		return s.showTodosForDay(ctx, chatID, messageID, userID, "tomorrow")
	case "done_today":
		return s.showDoneForDay(ctx, chatID, messageID, userID, "today")
	case "done_yesterday":
		return s.showDoneForDay(ctx, chatID, messageID, userID, "yesterday")
	case "done_day_before":
		return s.showDoneForDay(ctx, chatID, messageID, userID, "day_before_yesterday")
	case "back_to_main":
		return s.showMainMenuEdit(ctx, chatID, messageID)
	default:
		return nil
	}
}

// showMainMenu shows the main menu with buttons.
func (s *Processor) showMainMenu(ctx context.Context, chatID int64) error {
	text := "Главное меню\n\nВыберите действие:"

	buttons := [][]telego.InlineKeyboardButton{
		{
			tu.InlineKeyboardButton("📝 Показать TODO").WithCallbackData("show_todos"),
		},
		{
			tu.InlineKeyboardButton("✅ Показать выполненное").WithCallbackData("show_done"),
		},
	}

	return s.botService.SendMessageWithButtons(ctx, chatID, text, buttons)
}

// showMainMenuEdit edits message to show main menu.
func (s *Processor) showMainMenuEdit(ctx context.Context, chatID int64, messageID int) error {
	text := "Главное меню\n\nВыберите действие:"

	buttons := [][]telego.InlineKeyboardButton{
		{
			tu.InlineKeyboardButton("📝 Показать TODO").WithCallbackData("show_todos"),
		},
		{
			tu.InlineKeyboardButton("✅ Показать выполненное").WithCallbackData("show_done"),
		},
	}

	return s.botService.EditMessageText(ctx, chatID, messageID, text, buttons)
}

// showTodosMenu shows menu for selecting TODO period.
func (s *Processor) showTodosMenu(ctx context.Context, chatID int64, messageID int) error {
	text := "📝 TODO\n\nВыберите период:"

	buttons := [][]telego.InlineKeyboardButton{
		{
			tu.InlineKeyboardButton("Сегодня").WithCallbackData("todos_today"),
		},
		{
			tu.InlineKeyboardButton("Завтра").WithCallbackData("todos_tomorrow"),
		},
		{
			tu.InlineKeyboardButton("« Назад").WithCallbackData("back_to_main"),
		},
	}

	return s.botService.EditMessageText(ctx, chatID, messageID, text, buttons)
}

// showDoneMenu shows menu for selecting done period.
func (s *Processor) showDoneMenu(ctx context.Context, chatID int64, messageID int) error {
	text := "✅ Выполненное\n\nВыберите период:"

	buttons := [][]telego.InlineKeyboardButton{
		{
			tu.InlineKeyboardButton("Сегодня").WithCallbackData("done_today"),
		},
		{
			tu.InlineKeyboardButton("Вчера").WithCallbackData("done_yesterday"),
		},
		{
			tu.InlineKeyboardButton("Позавчера").WithCallbackData("done_day_before"),
		},
		{
			tu.InlineKeyboardButton("« Назад").WithCallbackData("back_to_main"),
		},
	}

	return s.botService.EditMessageText(ctx, chatID, messageID, text, buttons)
}

// showTodosForDay shows TODOs for a specific day.
func (s *Processor) showTodosForDay(ctx context.Context, chatID int64, messageID int, userID int64, period string) error {
	var dayText string
	var dateOffset int

	switch period {
	case "today":
		dayText = "сегодня"
		dateOffset = 0
	case "tomorrow":
		dayText = "завтра"
		dateOffset = 1
	default:
		dayText = period
		dateOffset = 0
	}

	// Get todos for the specified day
	date := carbon.Now().AddDays(dateOffset)
	dateFrom := date.StartOfDay().StdTime()
	dateTo := date.EndOfDay().StdTime()

	isCompleted := false
	params := repo_todos.FindParams{
		UserID:      userID,
		IsCompleted: &isCompleted,
		DateFrom:    &dateFrom,
		DateTo:      &dateTo,
		OrderBy:     "created_at",
	}

	result, err := s.baseService.Todos().Find(ctx, params)
	if err != nil {
		return fmt.Errorf("find todos: %w", err)
	}

	var text strings.Builder
	text.WriteString(fmt.Sprintf("📝 TODO на %s\n\n", dayText))

	if result.Count == 0 {
		text.WriteString("Задач не найдено")
	} else {
		for i, todo := range result.Items {
			text.WriteString(fmt.Sprintf("%d. ", i+1))
			if todo.Status == models.TodoStatusInProgress {
				text.WriteString("⏳ ")
			}
			text.WriteString(todo.Title)
			if todo.Description != "" {
				text.WriteString(fmt.Sprintf("\n   %s", todo.Description))
			}
			text.WriteString("\n\n")
		}
		text.WriteString(fmt.Sprintf("Всего: %d", result.Count))
	}

	buttons := [][]telego.InlineKeyboardButton{
		{
			tu.InlineKeyboardButton("« Назад").WithCallbackData("show_todos"),
		},
	}

	return s.botService.EditMessageText(ctx, chatID, messageID, text.String(), buttons)
}

// showDoneForDay shows done tasks for a specific day.
func (s *Processor) showDoneForDay(ctx context.Context, chatID int64, messageID int, userID int64, period string) error {
	var dayText string
	var dateOffset int

	switch period {
	case "today":
		dayText = "сегодня"
		dateOffset = 0
	case "yesterday":
		dayText = "вчера"
		dateOffset = -1
	case "day_before_yesterday":
		dayText = "позавчера"
		dateOffset = -2
	default:
		dayText = period
		dateOffset = 0
	}

	// Get done tasks for the specified day
	date := carbon.Now().AddDays(dateOffset)
	dateFrom := date.StartOfDay().StdTime()
	dateTo := date.EndOfDay().StdTime()

	isCompleted := true
	params := repo_todos.FindParams{
		UserID:      userID,
		IsCompleted: &isCompleted,
		DateFrom:    &dateFrom,
		DateTo:      &dateTo,
		OrderBy:     "completed_at",
	}

	result, err := s.baseService.Todos().Find(ctx, params)
	if err != nil {
		return fmt.Errorf("find done: %w", err)
	}

	var text strings.Builder
	text.WriteString(fmt.Sprintf("✅ Выполнено %s\n\n", dayText))

	if result.Count == 0 {
		text.WriteString("Задач не найдено")
	} else {
		for i, todo := range result.Items {
			text.WriteString(fmt.Sprintf("%d. ✓ %s", i+1, todo.Title))
			if todo.Description != "" {
				text.WriteString(fmt.Sprintf("\n   %s", todo.Description))
			}
			if todo.CompletedAt != nil && !todo.CompletedAt.IsZero() {
				text.WriteString(fmt.Sprintf("\n   ⏰ %s", todo.CompletedAt.Format("15:04")))
			}
			text.WriteString("\n\n")
		}
		text.WriteString(fmt.Sprintf("Всего: %d", result.Count))
	}

	buttons := [][]telego.InlineKeyboardButton{
		{
			tu.InlineKeyboardButton("« Назад").WithCallbackData("show_done"),
		},
	}

	return s.botService.EditMessageText(ctx, chatID, messageID, text.String(), buttons)
}
