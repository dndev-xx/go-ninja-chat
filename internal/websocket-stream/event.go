package websocketstream

import (
	"encoding/json"
	"io"

	eventstream "github.com/dndev-xx/go-ninja-chat/internal/services/event-stream"
)

// EventAdapter converts the event from the stream to the appropriate object.
type EventAdapter interface {
	Adapt(event eventstream.Event) (any, error)
}

// EventWriter write adapted event it to the socket.
type EventWriter interface {
	Write(event any, out io.Writer) error
}

type JSONEventWriter struct{}

func (JSONEventWriter) Write(event any, out io.Writer) error {
	if err := json.NewEncoder(out).Encode(event); err != nil {
		return err
	}
	return nil
}
