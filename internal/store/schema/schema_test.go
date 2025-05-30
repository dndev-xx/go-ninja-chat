package schema_test

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dndev-xx/go-ninja-chat/internal/store"
	storechat "github.com/dndev-xx/go-ninja-chat/internal/store/chat"
	"github.com/dndev-xx/go-ninja-chat/internal/store/enttest"
	"github.com/dndev-xx/go-ninja-chat/internal/store/message"
	"github.com/dndev-xx/go-ninja-chat/internal/store/problem"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

func TestChatServiceSchema(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := enttest.Open(t, "sqlite3",
		"file:schema_test.TestChatServiceSchema?mode=memory&cache=shared&_fk=1")
	defer func() { require.NoError(t, client.Close()) }()

	clientID := types.NewUserID()
	managerID := types.NewUserID()

	// Init.

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

	_ = client.Message.CreateBulk(
		// Dialog 1.
		client.Message.
			Create().
			SetChatID(chat.ID).
			SetProblemID(problems[0].ID).
			SetAuthorID(clientID).
			SetIsVisibleForClient(true).
			SetIsVisibleForManager(true).
			SetIsBlocked(false).
			SetIsService(false).
			SetBody("Hello, manager!"),

		client.Message.
			Create().
			SetChatID(chat.ID).
			SetProblemID(problems[0].ID).
			SetAuthorID(managerID).
			SetIsVisibleForClient(true).
			SetIsVisibleForManager(true).
			SetIsBlocked(false).
			SetIsService(false).
			SetBody("Hello, client!"),

		// Dialog 2.
		client.Message.
			Create().
			SetChatID(chat.ID).
			SetProblemID(problems[1].ID).
			SetAuthorID(clientID).
			SetIsVisibleForClient(true).
			SetIsVisibleForManager(true).
			SetIsBlocked(false).
			SetIsService(false).
			SetBody("I lost my money."),

		client.Message.
			Create().
			SetChatID(chat.ID).
			SetProblemID(problems[1].ID).
			SetAuthorID(managerID).
			SetIsVisibleForClient(true).
			SetIsVisibleForManager(true).
			SetIsBlocked(false).
			SetIsService(false).
			SetBody("No money, no honey."),
	).SaveX(ctx)

	// Querying.
	var chatProblemIDs []types.ProblemID
	client.Chat.QueryProblems(chat).Select(problem.FieldID).ScanX(ctx, &chatProblemIDs)
	assert.Equal(t, []types.ProblemID{problems[0].ID, problems[1].ID}, chatProblemIDs)

	p1messages := client.Problem.QueryMessages(problems[0]).Select(message.FieldBody).StringsX(ctx)
	assert.Equal(t, []string{"Hello, manager!", "Hello, client!"}, p1messages)

	p2messages := client.Problem.QueryMessages(problems[1]).Select(message.FieldBody).StringsX(ctx)
	assert.Equal(t, []string{"I lost my money.", "No money, no honey."}, p2messages)

	t.Run("assert edges", func(t *testing.T) {
		chat := client.Chat.Query().
			Where(storechat.ID(chat.ID)).
			WithMessages().
			WithProblems(func(query *store.ProblemQuery) {
				query.
					WithChat().
					WithMessages(func(query *store.MessageQuery) {
						query.
							WithProblem().
							WithChat()
					})
			}).
			OnlyX(ctx)

		assert.Len(t, chat.Edges.Problems, 2)
		assert.Len(t, chat.Edges.Messages, 4)

		p1, p2 := chat.Edges.Problems[0], chat.Edges.Problems[1]
		{
			require.NotNil(t, p1.Edges.Chat)
			assert.Equal(t, chat.ID, p1.Edges.Chat.ID)
			assert.Len(t, p1.Edges.Messages, 2)

			require.NotNil(t, p2.Edges.Chat)
			assert.Equal(t, chat.ID, p2.Edges.Chat.ID)
			assert.Len(t, p2.Edges.Messages, 2)
		}

		m11, m21 := p1.Edges.Messages[0], p2.Edges.Messages[1]
		{
			require.NotNil(t, m11.Edges.Chat)
			assert.Equal(t, chat.ID, m11.Edges.Chat.ID)
			require.NotNil(t, m11.Edges.Problem)
			assert.Equal(t, p1.ID, m11.Edges.Problem.ID)

			require.NotNil(t, m21.Edges.Chat)
			assert.Equal(t, chat.ID, m21.Edges.Chat.ID)
			require.NotNil(t, m21.Edges.Problem)
			assert.Equal(t, p2.ID, m21.Edges.Problem.ID)
		}
	})

	t.Run("client must has only one chat", func(t *testing.T) {
		_, err := client.Chat.Create().SetClientID(clientID).Save(ctx)
		t.Log(err)
		require.Error(t, err)
	})

	t.Run("do not accept zero ids", func(t *testing.T) {
		_, err := client.Chat.Create().
			SetClientID(types.UserID{}).
			Save(ctx)
		t.Log(err)
		require.Error(t, err)
	})
}
