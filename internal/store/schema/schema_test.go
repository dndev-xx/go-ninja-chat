package schema_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dndev-xx/go-ninja-chat/internal/store"
	storechat "github.com/dndev-xx/go-ninja-chat/internal/store/chat"
	"github.com/dndev-xx/go-ninja-chat/internal/store/enttest"
	"github.com/dndev-xx/go-ninja-chat/internal/store/message"
	"github.com/dndev-xx/go-ninja-chat/internal/store/problem"
)

func TestChatServiceSchema(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := enttest.Open(t, "sqlite3",
		"file:schema_test.TestChatServiceSchema?mode=memory&cache=shared&_fk=1")
	defer func() { require.NoError(t, client.Close()) }()

	clientID := uuid.New()
	managerID := uuid.New()

	// Init chat and problems
	chat := client.Chat.
		Create().
		SetClientID(clientID).
		SaveX(ctx)

	problems := client.Problem.
		CreateBulk(
			client.Problem.
				Create().
				SetChatID(chat.ID).
				SetManagerID(managerID),
			client.Problem.
				Create().
				SetChatID(chat.ID).
				SetManagerID(managerID),
		).SaveX(ctx)

	// Create messages
	messages := client.Message.CreateBulk(
		// First problem messages
		client.Message.
			Create().
			SetChatID(chat.ID).
			SetProblemID(problems[0].ID).
			SetAuthorID(clientID).
			SetIsVisibleForClient(true).
			SetIsVisibleForManager(true).
			SetBody("Client message 1"),
		client.Message.
			Create().
			SetChatID(chat.ID).
			SetProblemID(problems[0].ID).
			SetAuthorID(managerID).
			SetIsVisibleForClient(true).
			SetIsVisibleForManager(true).
			SetBody("Manager reply 1"),

		// Second problem messages
		client.Message.
			Create().
			SetChatID(chat.ID).
			SetProblemID(problems[1].ID).
			SetAuthorID(clientID).
			SetIsVisibleForClient(true).
			SetIsVisibleForManager(true).
			SetBody("Client message 2"),
		client.Message.
			Create().
			SetChatID(chat.ID).
			SetProblemID(problems[1].ID).
			SetAuthorID(managerID).
			SetIsVisibleForClient(true).
			SetIsVisibleForManager(true).
			SetBody("Manager reply 2"),
	).SaveX(ctx)

	t.Run("check chat relationships", func(t *testing.T) {
		// Query chat with related entities
		queriedChat := client.Chat.Query().
			Where(storechat.ID(chat.ID)).
			WithProblems(func(q *store.ProblemQuery) {
				q.WithMessages()
			}).
			WithMessages().
			OnlyX(ctx)

		// Verify chat fields
		assert.Equal(t, clientID, queriedChat.ClientID)
		assert.NotZero(t, queriedChat.CreatedAt)

		// Verify problems
		require.Len(t, queriedChat.Edges.Problems, 2)
		assert.Equal(t, problems[0].ID, queriedChat.Edges.Problems[0].ID)
		assert.Equal(t, problems[1].ID, queriedChat.Edges.Problems[1].ID)

		// Verify messages
		require.Len(t, queriedChat.Edges.Messages, 4)
		for _, msg := range queriedChat.Edges.Messages {
			assert.Equal(t, chat.ID, msg.ChatID)
		}

		// Verify problem messages
		for _, prob := range queriedChat.Edges.Problems {
			assert.Len(t, prob.Edges.Messages, 2)
			for _, msg := range prob.Edges.Messages {
				assert.Equal(t, prob.ID, msg.ProblemID)
			}
		}
	})

	t.Run("check problem relationships", func(t *testing.T) {
		for _, prob := range problems {
			queriedProblem := client.Problem.Query().
				Where(problem.ID(prob.ID)).
				WithChat().
				WithMessages().
				OnlyX(ctx)

			// Verify problem fields
			assert.Equal(t, chat.ID, queriedProblem.ChatID)
			assert.Equal(t, managerID, queriedProblem.ManagerID)
			assert.NotZero(t, queriedProblem.CreatedAt)

			// Verify chat relationship
			require.NotNil(t, queriedProblem.Edges.Chat)
			assert.Equal(t, chat.ID, queriedProblem.Edges.Chat.ID)

			// Verify messages
			require.Len(t, queriedProblem.Edges.Messages, 2)
			for _, msg := range queriedProblem.Edges.Messages {
				assert.Equal(t, prob.ID, msg.ProblemID)
			}
		}
	})

	t.Run("check message relationships", func(t *testing.T) {
		for _, msg := range messages {
			queriedMessage := client.Message.Query().
				Where(message.ID(msg.ID)).
				WithChat().
				WithProblem().
				OnlyX(ctx)

			// Verify message fields
			assert.Equal(t, chat.ID, queriedMessage.ChatID)
			assert.True(t, queriedMessage.IsVisibleForClient)
			assert.True(t, queriedMessage.IsVisibleForManager)
			assert.NotZero(t, queriedMessage.CreatedAt)

			// Verify chat relationship
			require.NotNil(t, queriedMessage.Edges.Chat)
			assert.Equal(t, chat.ID, queriedMessage.Edges.Chat.ID)

			// Verify problem relationship
			require.NotNil(t, queriedMessage.Edges.Problem)
			assert.Contains(t, []uuid.UUID{problems[0].ID, problems[1].ID}, queriedMessage.Edges.Problem.ID)
		}
	})

	t.Run("unique client constraint", func(t *testing.T) {
		_, err := client.Chat.
			Create().
			SetClientID(clientID).
			Save(ctx)
		require.Error(t, err)
	})
	
	t.Run("non-zero client ID required", func(t *testing.T) {
		_, err := client.Chat.
			Create().
			SetClientID(uuid.Nil).
			Save(ctx)
		require.NoError(t, err)
	})
}