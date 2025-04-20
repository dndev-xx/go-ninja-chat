//go:build integration

package messages_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	messagesrepo "github.com/dndev-xx/go-ninja-chat/internal/repositories/messages"
	"github.com/dndev-xx/go-ninja-chat/internal/store"
	"github.com/dndev-xx/go-ninja-chat/internal/testingh"
)

type MsgRepoHistoryAPISuite struct {
	testingh.DBSuite
	repo *messagesrepo.Repo
}

func TestMsgRepoHistoryAPISuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &MsgRepoHistoryAPISuite{DBSuite: testingh.NewDBSuite("TestMsgRepoHistoryAPISuite")})
}

func (s *MsgRepoHistoryAPISuite) SetupSuite() {
	s.DBSuite.SetupSuite()

	var err error
	s.repo, err = messagesrepo.New(messagesrepo.NewOptions(s.Database))
	s.Require().NoError(err)
}

func (s *MsgRepoHistoryAPISuite) Test_GetClientChatMessages() {
	s.Run("too small page size", func() {
		msgs, next, err := s.repo.GetClientChatMessages(s.Ctx, uuid.New(), 9, nil)
		s.Require().ErrorIs(err, messagesrepo.ErrInvalidPageSize)
		s.Nil(next)
		s.Empty(msgs)
	})

	s.Run("too big page size", func() {
		msgs, next, err := s.repo.GetClientChatMessages(s.Ctx, uuid.New(), 101, nil)
		s.Require().ErrorIs(err, messagesrepo.ErrInvalidPageSize)
		s.Nil(next)
		s.Empty(msgs)
	})

	s.Run("no last created at in cursor", func() {
		msgs, next, err := s.repo.GetClientChatMessages(s.Ctx, uuid.New(), 0, &messagesrepo.Cursor{
			LastCreatedAt: time.Time{},
			PageSize:      50,
		})
		s.Require().ErrorIs(err, messagesrepo.ErrInvalidCursor)
		s.Nil(next)
		s.Empty(msgs)
	})

	s.Run("too small page size in cursor", func() {
		msgs, next, err := s.repo.GetClientChatMessages(s.Ctx, uuid.New(), 0, &messagesrepo.Cursor{
			LastCreatedAt: time.Now(),
			PageSize:      9,
		})
		s.Require().ErrorIs(err, messagesrepo.ErrInvalidCursor)
		s.Nil(next)
		s.Empty(msgs)
	})

	s.Run("too big page size in cursor", func() {
		msgs, next, err := s.repo.GetClientChatMessages(s.Ctx, uuid.New(), 0, &messagesrepo.Cursor{
			LastCreatedAt: time.Now(),
			PageSize:      101,
		})
		s.Require().ErrorIs(err, messagesrepo.ErrInvalidCursor)
		s.Nil(next)
		s.Empty(msgs)
	})

	s.Run("client has not got any messages", func() {
		msgs, next, err := s.repo.GetClientChatMessages(s.Ctx, uuid.New(), 50, nil)
		s.Require().NoError(err)
		s.Nil(next)
		s.Empty(msgs)
	})

	s.Run("cursor logic", func() {
		const messagesCount = 30
		client1 := uuid.New()

		problem1, chat1 := s.createProblemAndChat(client1)
		preparedMsgs := s.createMessages(messagesCount, chat1, problem1, client1, true, true, false)
		s.Require().Len(preparedMsgs, messagesCount)

		// Messages from other chat must be ignored.
		client2 := uuid.New()
		problem2, chat2 := s.createProblemAndChat(client2)
		s.createMessages(3, chat2, problem2, client2, true, true, false)

		// Invisible for client messages must be ignored.
		s.createMessages(4, chat1, problem1, client1, false, true, false)

		for pageSize := 10; pageSize <= 20; pageSize++ {
			s.Run(fmt.Sprintf("page size %d", pageSize), func() {
				expected := batch[msg](pageSize, apply[*store.Message, msg](preparedMsgs, newMsgFromStoreMsg))
				actual, actualCursors := s.getClientChatMessagesWhileCursor(client1, pageSize)
				s.Run("pages", func() {
					s.Equal(expected, actual)
				})

				expectedCursors := make([]cursor, 0, len(expected))
				for i, b := range expected {
					if len(b) == pageSize && (i != len(expected)-1) {
						last := b[len(b)-1]
						expectedCursors = append(expectedCursors, cursor{
							PageSize:                 pageSize,
							LastCreatedAtAsUnixMilli: last.CreatedAtAsUnixMilli,
						})
					}
				}
				s.Run("cursors", func() {
					s.Equal(expectedCursors, actualCursors)
				})
			})
		}
	})

	s.Run("adapt logic", func() {
		client := uuid.New()
		problem, chat := s.createProblemAndChat(client)

		s.createMessages(1, chat, problem, nil, true, false, true)
		lastMsg := s.createMessages(1, chat, problem, client, true, false, false)[0]

		msgs, _, err := s.repo.GetClientChatMessages(s.Ctx, client, 11, nil)
		s.Require().NoError(err)
		s.Require().Len(msgs, 2)

		msg := msgs[0]
		s.Equal(lastMsg.ID, msg.ID)
		s.Equal(chat, msg.ChatID)
		s.Equal(client, msg.AuthorID)
		s.Equal("message #0", msg.Body)
		s.True(msg.CreatedAt.Equal(lastMsg.CreatedAt))
		s.True(msg.IsVisibleForClient)
		s.False(msg.IsVisibleForManager)
		s.False(msg.IsBlocked)
		s.False(msg.IsService)

		s.Run("service message", func() {
			svcMsg := msgs[1]
			s.True(svcMsg.AuthorID.IsZero())
			s.True(svcMsg.IsVisibleForClient)
			s.False(svcMsg.IsVisibleForManager)
			s.False(svcMsg.IsBlocked)
			s.True(svcMsg.IsService)
		})
	})
}

func (s *MsgRepoHistoryAPISuite) createProblemAndChat(clientID uuid.UUID) (uuid.UUID, uuid.UUID) {
	s.T().Helper()

	chat, err := s.Database.Chat(s.Ctx).Create().SetClientID(clientID).Save(s.Ctx)
	s.Require().NoError(err)

	problem, err := s.Database.Problem(s.Ctx).Create().SetChatID(chat.ID).Save(s.Ctx)
	s.Require().NoError(err)

	return problem.ID, chat.ID
}

// createMessages creates messages and returns it in order from newest to oldest.
func (s *MsgRepoHistoryAPISuite) createMessages(
	count int,
	chatID uuid.UUID,
	problemID uuid.UUID,
	clientID uuid.UUID,
	visibleForClient bool,
	visibleForManager bool,
	isService bool,
) []*store.Message {
	s.T().Helper()

	messages := make([]*store.Message, count)
	for i := 0; i < count; i++ {
		op := s.Database.Message(s.Ctx).Create().
			SetChatID(chatID).
			SetProblemID(problemID).
			SetIsVisibleForClient(visibleForClient).
			SetIsVisibleForManager(visibleForManager).
			SetIsService(isService).
			SetBody(fmt.Sprintf("message #%d", i))
		if !clientID.IsZero() {
			op.SetAuthorID(clientID)
		}
		msg, err := op.Save(s.Ctx)

		s.Require().NoError(err)
		s.Require().NotNil(msg)
		messages[i] = msg

		time.Sleep(time.Millisecond)
	}

	for i := 0; i < len(messages)/2; i++ {
		messages[i], messages[len(messages)-i-1] = messages[len(messages)-i-1], messages[i]
	}
	return messages
}

type msg struct {
	ID                   types.MessageID
	CreatedAtAsUnixMilli int64
}

func newMsgFromRepoMsg(m messagesrepo.Message) msg {
	return msg{
		ID:                   m.ID,
		CreatedAtAsUnixMilli: m.CreatedAt.UnixMilli(),
	}
}

func newMsgFromStoreMsg(m *store.Message) msg {
	return msg{
		ID:                   m.ID,
		CreatedAtAsUnixMilli: m.CreatedAt.UnixMilli(),
	}
}

type cursor struct {
	PageSize                 int
	LastCreatedAtAsUnixMilli int64
}

func (s *MsgRepoHistoryAPISuite) getClientChatMessagesWhileCursor(clientID types.UserID, pageSize int) ([][]msg, []cursor) {
	s.T().Helper()

	var result [][]msg
	var cursors []cursor

	var (
		msgs []messagesrepo.Message
		err  error
		next *messagesrepo.Cursor
	)
	for {
		msgs, next, err = s.repo.GetClientChatMessages(s.Ctx, clientID, pageSize, next)
		s.Require().NoError(err)
		result = append(result, apply[messagesrepo.Message, msg](msgs, newMsgFromRepoMsg))

		if next == nil {
			break
		}
		cursors = append(cursors, cursor{
			PageSize:                 next.PageSize,
			LastCreatedAtAsUnixMilli: next.LastCreatedAt.UnixMilli(),
		})
	}

	return result, cursors
}

func apply[In any, Out any](in []In, f func(v In) Out) []Out {
	result := make([]Out, 0, len(in))
	for _, v := range in {
		result = append(result, f(v))
	}
	return result
}

func batch[T any](bulkSize int, msgs []T) [][]T {
	result := make([][]T, 0, len(msgs)/bulkSize)
	for i := 0; i < len(msgs); i += bulkSize {
		left := i

		right := i + bulkSize
		if right > len(msgs) {
			right = len(msgs)
		}

		result = append(result, msgs[left:right])
	}
	return result
}
