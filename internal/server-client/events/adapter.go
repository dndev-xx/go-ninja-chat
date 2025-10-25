package clientevents

import (
	eventstream "github.com/dndev-xx/go-ninja-chat/internal/services/event-stream"
	websocketstream "github.com/dndev-xx/go-ninja-chat/internal/websocket-stream"
)

var _ websocketstream.EventAdapter = Adapter{}

type Adapter struct{}

func (Adapter) Adapt(ev eventstream.Event) (any, error) {
	// FIXME: Реализуй меня.
	return nil, nil
}
