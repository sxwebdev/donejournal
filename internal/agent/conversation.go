package agent

import (
	"context"
	"fmt"

	"github.com/sxwebdev/donejournal/internal/agent/provider"
	"github.com/sxwebdev/donejournal/internal/store/badgerdb"
	"github.com/sxwebdev/donejournal/internal/store/storecmn"
)

const maxConversationMessages = 20

// ConversationStore manages per-user conversation history in BadgerDB.
type ConversationStore struct {
	db *badgerdb.DB
}

// NewConversationStore creates a new ConversationStore.
func NewConversationStore(db *badgerdb.DB) *ConversationStore {
	return &ConversationStore{db: db}
}

func conversationKey(userID int64) []byte {
	return fmt.Appendf(nil, "conv:%d", userID)
}

// Load retrieves the conversation history for a user.
// Returns empty slice if no history exists.
func (s *ConversationStore) Load(ctx context.Context, userID int64) ([]provider.ChatMessage, error) {
	var messages []provider.ChatMessage
	err := s.db.GetFromJSON(ctx, conversationKey(userID), &messages)
	if err != nil {
		if err == storecmn.ErrNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("load conversation: %w", err)
	}
	return messages, nil
}

// Save stores conversation history, keeping only the last N messages.
func (s *ConversationStore) Save(ctx context.Context, userID int64, messages []provider.ChatMessage) error {
	// Prune to keep last maxConversationMessages
	if len(messages) > maxConversationMessages {
		messages = messages[len(messages)-maxConversationMessages:]
	}

	if err := s.db.SetJSON(ctx, conversationKey(userID), messages, 0); err != nil {
		return fmt.Errorf("save conversation: %w", err)
	}
	return nil
}

// Clear removes conversation history for a user.
func (s *ConversationStore) Clear(ctx context.Context, userID int64) error {
	return s.db.Delete(ctx, conversationKey(userID))
}
