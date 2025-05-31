package sendclientmessagejob

import (
	"encoding/json"
	"fmt"

	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

type payload struct {
	MessageID types.MessageID `json:"message_id"`
}

func MarshalPayload(messageID types.MessageID) (string, error) {
	if messageID.IsZero() {
		return "", fmt.Errorf("invalid message ID")
	}
	p := payload{
		MessageID: messageID,
	}

	jsonData, err := json.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	return string(jsonData), nil
}

func UnmarshalPayload(payloadStr string) (types.MessageID, error) {
	var p payload
	if err := json.Unmarshal([]byte(payloadStr), &p); err != nil {
		return types.MessageIDNil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if p.MessageID.IsZero() {
		return types.MessageIDNil, fmt.Errorf("invalid message ID in payload")
	}

	return p.MessageID, nil
}
