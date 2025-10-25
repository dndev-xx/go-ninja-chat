package inmemeventstream_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"

	eventstream "github.com/dndev-xx/go-ninja-chat/internal/services/event-stream"
	inmemeventstream "github.com/dndev-xx/go-ninja-chat/internal/services/event-stream/in-mem"
	"github.com/dndev-xx/go-ninja-chat/internal/testingh"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

type ServiceSuite struct {
	testingh.ContextSuite
	stream eventstream.EventStream
}

func TestServiceSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ServiceSuite))
}

func (s *ServiceSuite) SetupTest() {
	s.ContextSuite.SetupTest()
	s.stream = inmemeventstream.New()
}

func (s *ServiceSuite) TearDownTest() {
	// s.NoError(s.stream.Close())
	s.ContextSuite.TearDownTest()
}

func (s *ServiceSuite) TestSimpleSubscription() {
	// Arrange.
	uid := types.NewUserID()

	ctx, cancel := context.WithCancel(s.Ctx)
	defer cancel()

	events, err := s.stream.Subscribe(ctx, uid)
	s.Require().NoError(err)

	bodies := []string{"Hello", "World", "!"}
	result := readNewMessageEvents(events, len(bodies))

	// Action.
	for _, b := range bodies {
		s.Require().NoError(s.stream.Publish(ctx, uid, newMessageEvent(b)))
	}

	// Assert.
	s.Equal([]string{"Hello", "World", "!"}, <-result)
}

func (s *ServiceSuite) TestEventIsMultiplexedToStreams() {
	// Arrange.
	uid := types.NewUserID()
	tab1, err := s.stream.Subscribe(s.Ctx, uid)
	s.Require().NoError(err)

	tab2, err := s.stream.Subscribe(s.Ctx, uid)
	s.Require().NoError(err)

	tab3, err := s.stream.Subscribe(s.Ctx, uid)
	s.Require().NoError(err)

	const (
		tabsCount        = 3
		messagesCount    = 5
		allMessagesCount = tabsCount * messagesCount
	)

	// Action.
	expectedCh := make(chan []string)
	go func() {
		expected := make([]string, 0, allMessagesCount)
		for i := 0; i < messagesCount; i++ {
			v := strconv.Itoa(i)
			err := s.stream.Publish(s.Ctx, uid, newMessageEvent(v))
			s.NoError(err)

			for i := 0; i < tabsCount; i++ {
				expected = append(expected, v)
			}
		}
		expectedCh <- expected
	}()

	// Assert.
	msgs := make([]string, 0, allMessagesCount)
	for i := 0; i < allMessagesCount; i++ {
		var event eventstream.Event
		select {
		case event = <-tab1:
		case event = <-tab2:
		case event = <-tab3:
		case <-time.After(time.Second):
			s.FailNow("lost events")
		}
		msgs = append(msgs, event.(*eventstream.MessageSentEvent).Body)
	}
	s.ElementsMatch(<-expectedCh, msgs)
}

func (s *ServiceSuite) TestPublishInvalidEvent() {
	uid := types.NewUserID()

	events, err := s.stream.Subscribe(s.Ctx, uid)
	s.Require().NoError(err)

	// Not filled event.
	err = s.stream.Publish(s.Ctx, uid, &eventstream.MessageSentEvent{})
	s.Require().Error(err)

	select {
	case ev := <-events:
		s.FailNow("unexpected event", ev)
	case <-time.After(100 * time.Millisecond):
	}
}

func (s *ServiceSuite) TestPublishWithoutSubscribers() {
	s.Run("no subscriptions at all", func() {
		err := s.stream.Publish(s.Ctx, types.NewUserID(), newMessageEvent("Hello"))
		s.Require().NoError(err)
	})

	s.Run("publish to offline client", func() {
		uid1, uid2 := types.NewUserID(), types.NewUserID()

		// uid1 is online.
		_, err := s.stream.Subscribe(s.Ctx, uid1)
		s.Require().NoError(err)

		// uid2 is offline.
		err = s.stream.Publish(s.Ctx, uid2, newMessageEvent("No panic"))
		s.Require().NoError(err)
	})

	s.Run("client was online and became offline", func() {
		// Arrange.
		uid := types.NewUserID()

		subscribe := func(n int) (<-chan []string, context.CancelFunc) {
			ctx, cancel := context.WithCancel(s.Ctx)
			// No cancel().

			tab, err := s.stream.Subscribe(ctx, uid)
			s.Require().NoError(err)

			return readNewMessageEvents(tab, n), func() {
				time.Sleep(10 * time.Millisecond)
				cancel()
			}
		}

		publish := func(v string) {
			err := s.stream.Publish(s.Ctx, uid, newMessageEvent(v))
			s.Require().NoError(err)
		}

		// Action.
		tab1, cancel1 := subscribe(-1)
		publish("1")

		tab2, cancel2 := subscribe(-1)
		publish("2")

		tab3, cancel3 := subscribe(-1)
		publish("3")

		cancel3()
		publish("4")

		cancel2()
		publish("5")

		cancel1()
		publish("6")

		// Assert.
		s.Equal([]string{"1", "2", "3", "4", "5"}, <-tab1)
		s.Equal([]string{"2", "3", "4"}, <-tab2)
		s.Equal([]string{"3"}, <-tab3)
	})
}

func (s *ServiceSuite) TestPublishInDifferentUserStreams() {
	// Arrange.
	const users = 3
	const messagesPerUser = 10

	uids := make([]types.UserID, 0, users)
	msgChannels := make([]<-chan []string, 0, users)

	for i := 0; i < users; i++ {
		uid := types.NewUserID()

		events, err := s.stream.Subscribe(s.Ctx, uid)
		s.Require().NoError(err)

		uids = append(uids, uid)
		msgChannels = append(msgChannels, readNewMessageEvents(events, messagesPerUser))
	}

	// Action.
	expectedMsgs := make([][]string, users)
	for i := 0; i < users; i++ {
		expectedMsgs[i] = make([]string, 0, messagesPerUser)
	}

	for i := 0; i < messagesPerUser; i++ {
		for j := 0; j < users; j++ {
			uid := uids[j]
			v := strconv.Itoa(i*users + j)

			err := s.stream.Publish(s.Ctx, uid, newMessageEvent(v))
			s.Require().NoError(err)

			expectedMsgs[j] = append(expectedMsgs[j], v)
		}
	}

	// Assert.
	receivedMsgs := make([][]string, 0, users)
	for _, ch := range msgChannels {
		receivedMsgs = append(receivedMsgs, <-ch)
	}

	s.T().Log("received events", receivedMsgs)
	s.Equal(expectedMsgs, receivedMsgs)
}

// readNewMessageEvents reads n events from the stream.
// If n is negative, then the function reads the stream until it is closed.
func readNewMessageEvents(stream <-chan eventstream.Event, n int) <-chan []string {
	result := make(chan []string, len(stream))
	var msgs []string // No preallocation, n can be negative.
	go func() {
		for ev := range stream {
			msg := ev.(*eventstream.MessageSentEvent).Body
			msgs = append(msgs, msg)
			if n != -1 && len(msgs) == n {
				break
			}
		}
		result <- msgs
	}()
	return result
}

func newMessageEvent(body string) eventstream.Event {
	createaAt := time.Now()
	usrId := types.NewUserID()
	return &eventstream.MessageSentEvent{
		MessageID: types.NewMessageID(),
		ChatID:    types.NewChatID(),
		AuthorID:  &usrId,
		Body:      body,
		CreatedAt: &createaAt,
		EventType: "event",
		EventID:   types.NewEventID(),
		RequestID: types.NewRequestID(),
		IsService: false,
	}
}
