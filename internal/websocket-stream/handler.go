package websocketstream

import (
	"context"
	"fmt"
	"sync"
	"time"

	gorillaws "github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/dndev-xx/go-ninja-chat/internal/middlewares"
	eventstream "github.com/dndev-xx/go-ninja-chat/internal/services/event-stream"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

const (
	writeTimeout = time.Second
	pongTimeout  = 10 * time.Second
	pingPeriod   = (pongTimeout * 9) / 10

	// WebSocket message types
	TextMessage   = 1
	BinaryMessage = 2
	CloseMessage  = 8
	PingMessage   = 9
	PongMessage   = 10
)

type eventStream interface {
	Subscribe(ctx context.Context, userID types.UserID) (<-chan eventstream.Event, error)
}

//go:generate options-gen -out-filename=handler_options.gen.go -from-struct=Options
type Options struct {
	pingPeriod  time.Duration `default:"3s" validate:"omitempty,min=100ms,max=30s"`
	logger      *zap.Logger   `option:"mandatory" validate:"required"`
	eventStream eventStream

	eventAdapter EventAdapter
	eventWriter  EventWriter
	upgrader     Upgrader        `option:"mandatory" validate:"required"`
	shutdownCh   <-chan struct{} `option:"mandatory" validate:"required"`
}

type HTTPHandler struct {
	Options
}

func NewHTTPHandler(opts Options) (*HTTPHandler, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	return &HTTPHandler{
		Options: opts,
	}, nil
}

func (h *HTTPHandler) Serve(eCtx echo.Context) error {
	ws, err := h.upgrader.Upgrade(eCtx.Response(), eCtx.Request(), eCtx.Response().Header())
	if err != nil {
		h.logger.Error("Failed to upgrade connection", zap.Error(err))
		return err
	}
	defer ws.Close()
	userID := middlewares.GetAuthUserID(eCtx) // middlewares.GetAuthUserID(eCtx)
	ctx, cancel := context.WithCancel(eCtx.Request().Context())
	defer cancel()

	events, err := h.eventStream.Subscribe(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to subscribe to events", zap.Error(err))
		return err
	}
	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	// Read loop (handle PONGs and close messages)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()

		pongWait := h.pingPeriod * 2
		ws.SetReadDeadline(time.Now().Add(pongWait))
		ws.SetPongHandler(func(string) error {
			ws.SetReadDeadline(time.Now().Add(pongWait))
			h.logger.Debug("Received pong")
			return nil
		})

		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				if gorillaws.IsCloseError(err, gorillaws.CloseNormalClosure, gorillaws.CloseGoingAway) {
					h.logger.Debug("Connection closed normally")
					return
				}
				errCh <- fmt.Errorf("read error: %w", err)
				return
			}
		}
	}()

	// Write loop (send PINGs and events)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()

		pingTicker := time.NewTicker(h.pingPeriod)
		defer pingTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				h.sendCloseMessage(ws)
				return

			case <-pingTicker.C:
				h.logger.Debug("Sending ping")
				if err := ws.WriteControl(gorillaws.PingMessage, []byte{}, time.Now().Add(writeTimeout)); err != nil {
					errCh <- fmt.Errorf("ping error: %w", err)
					return
				}

			case event, ok := <-events:
				if !ok {
					h.logger.Info("Event channel closed")
					h.sendCloseMessage(ws)
					return
				}

				if err := h.writeEvent(ws, event); err != nil {
					errCh <- fmt.Errorf("write event error: %w", err)
					return
				}
			}
		}
	}()

	// Monitor shutdown signal
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-h.shutdownCh:
			h.logger.Info("Shutdown signal received")
			cancel()
		case <-ctx.Done():
		}
	}()

	wg.Wait()
	close(errCh)

	for err := range errCh {
		return err
	}

	return nil
}

func (h *HTTPHandler) sendCloseMessage(ws Websocket) {
	wsCloser := newWsCloser(h.logger, ws)
	wsCloser.Close(gorillaws.CloseNormalClosure)
	// msg := gorillaws.FormatCloseMessage(gorillaws.CloseNormalClosure, "")
	// _ = ws.WriteControl(gorillaws.CloseMessage, msg, time.Now().Add(writeTimeout))
}

func (h *HTTPHandler) writeEvent(ws Websocket, event eventstream.Event) error {
	adaptedEvent, err := h.eventAdapter.Adapt(event)
	if err != nil {
		return fmt.Errorf("adapt event: %w", err)
	}

	writer, err := ws.NextWriter(gorillaws.TextMessage)
	if err != nil {
		return fmt.Errorf("get writer: %w", err)
	}
	defer writer.Close()

	if err := ws.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		return err
	}

	if err := h.eventWriter.Write(adaptedEvent, writer); err != nil {
		return fmt.Errorf("write event: %w", err)
	}

	h.logger.Debug("Event written successfully", zap.String("event_type", fmt.Sprintf("%T", event)))
	return nil
}
