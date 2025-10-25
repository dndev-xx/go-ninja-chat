package websocketstream_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	gorillaws "github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/dndev-xx/go-ninja-chat/internal/logger"
	"github.com/dndev-xx/go-ninja-chat/internal/middlewares"
	eventstream "github.com/dndev-xx/go-ninja-chat/internal/services/event-stream"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
	websocketstream "github.com/dndev-xx/go-ninja-chat/internal/websocket-stream"
)

func init() {
	logger.MustInit(logger.NewOptions("debug"))
}

func TestHTTPHandler(t *testing.T) {
	const (
		eventsNum     = 3
		eventInterval = 100 * time.Millisecond
		pingInterval  = 100 * time.Millisecond
		origin        = "http://localhost"

		headerSecWsProtocol = "Sec-WebSocket-Protocol"
		secWsProtocol       = "chat-service-protocol.test"
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uid := types.NewUserID()
	eventsCh := make(chan eventstream.Event)
	shutdownCh := make(chan struct{})

	log := zap.L().Named("TestHTTPHandler")

	eventStreamMock := EventStreamMock{uid: uid, ch: eventsCh}

	h, err := websocketstream.NewHTTPHandler(websocketstream.NewOptions(
		zap.L(),
		websocketstream.NewUpgrader([]string{origin}, secWsProtocol),
		shutdownCh,
		websocketstream.WithPingPeriod(pingInterval),
		websocketstream.WithEventStream(eventStreamMock),
		websocketstream.WithEventAdapter(EventAdapter{}),
		websocketstream.WithEventWriter(websocketstream.JSONEventWriter{}),
	))
	require.NoError(t, err)

	e := echo.New()
	e.GET("/ws", middlewares.AuthWith(uid)(h.Serve))
	s := httptest.NewServer(e)
	defer s.Close()

	u := url.URL{Scheme: "ws", Host: s.Listener.Addr().String(), Path: "/ws"}

	header := http.Header{}
	header.Add(echo.HeaderOrigin, origin)
	header.Add(headerSecWsProtocol, secWsProtocol)

	c, resp, err := gorillaws.DefaultDialer.DialContext(ctx, u.String(), header)
	require.NoError(t, err)
	assert.Equal(t, secWsProtocol, resp.Header.Get(headerSecWsProtocol))
	defer func() {
		c.WriteControl(gorillaws.CloseMessage, nil, time.Now().Add(time.Second))
		require.NoError(t, c.Close())
		require.NoError(t, resp.Body.Close())
	}()

	var pings int
	var pingsMutex sync.Mutex

	c.SetPingHandler(func(appData string) error {
		pingsMutex.Lock()
		pings++
		pingsMutex.Unlock()
		log.Debug("new ping received, send pong")
		return c.WriteControl(gorillaws.PongMessage, []byte(appData), time.Now().Add(time.Second))
	})

	c.SetPongHandler(func(appData string) error {
		log.Debug("Received pong")
		return nil
	})

	events := make([]eventstream.Event, 0, eventsNum)
	for i := 0; i < eventsNum; i++ {
		events = append(events, newMessageEvent(fmt.Sprintf("message-%d", i), &uid))
	}

	go func() {
		defer func() {
			log.Debug("finished sending events")
		}()

		for i, event := range events {
			select {
			case eventsCh <- event:
				log.Debug("sent event", zap.Int("index", i))
				time.Sleep(eventInterval)
			case <-shutdownCh:
				log.Debug("shutdown during event sending")
				return
			case <-ctx.Done():
				return
			}
		}
		log.Debug("all events sent")
	}()

	receivedEvents := make([]*eventstream.MessageSentEvent, 0, len(events))
	var eventsMutex sync.Mutex

	readDone := make(chan struct{})
	go func() {
		defer close(readDone)

		for {
			select {
			case <-shutdownCh:
				return
			case <-ctx.Done():
				return
			default:
				var event eventstream.MessageSentEvent
				err := c.ReadJSON(&event)
				if err != nil {
					if gorillaws.IsCloseError(err, gorillaws.CloseNormalClosure, gorillaws.CloseGoingAway) {
						log.Debug("websocket closed normally")
						return
					}
					log.Debug("read error", zap.Error(err))
					return
				}

				eventsMutex.Lock()
				receivedEvents = append(receivedEvents, &event)
				currentCount := len(receivedEvents)
				eventsMutex.Unlock()

				log.Debug("new event received", zap.Int("count", currentCount))

				if currentCount >= eventsNum {
					log.Debug("received all events, initiating shutdown")
					close(shutdownCh)
					return
				}
			}
		}
	}()

	select {
	case <-readDone:
		log.Debug("read completed")
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for events")
	}

	time.Sleep(100 * time.Millisecond)

	t.Run("event stream is working properly", func(t *testing.T) {
		eventsMutex.Lock()
		defer eventsMutex.Unlock()
		require.Len(t, receivedEvents, eventsNum, "should receive all events")
		for i, e := range receivedEvents {
			assert.Equal(t, events[i].(*eventstream.MessageSentEvent).Body, e.Body,
				"event content mismatch at index %d", i)
		}
	})

	t.Run("ping-pong mechanism is working properly", func(t *testing.T) {
		pingsMutex.Lock()
		defer pingsMutex.Unlock()
		t.Logf("pings: %d", pings)
		assert.Greater(t, pings, 0, "should have at least some pings")
	})

	t.Run("shutdown is working properly", func(t *testing.T) {
		err := c.WriteMessage(gorillaws.TextMessage, []byte("test"))
		assert.NoError(t, err, "should not be able to write to closed connection")
	})
}

type EventStreamMock struct {
	uid types.UserID
	ch  chan eventstream.Event
}

func (e EventStreamMock) Subscribe(ctx context.Context, userID types.UserID) (<-chan eventstream.Event, error) {
	if e.uid != userID {
		return nil, fmt.Errorf("unexpected user: %v != %v", e.uid, userID)
	}
	return e.ch, nil
}

func (e EventStreamMock) Unsubscribe(ctx context.Context, userID types.UserID) error {
	return nil
}

type EventAdapter struct{}

func (EventAdapter) Adapt(event eventstream.Event) (any, error) {
	return event, nil
}

func newMessageEvent(body string, userID *types.UserID) *eventstream.MessageSentEvent {
	createdAt := time.Now()
	return &eventstream.MessageSentEvent{
		MessageID: types.NewMessageID(),
		ChatID:    types.NewChatID(),
		AuthorID:  userID,
		Body:      body,
		CreatedAt: &createdAt,
		EventType: "event",
		EventID:   types.NewEventID(),
		RequestID: types.NewRequestID(),
		IsService: false,
	}
}
