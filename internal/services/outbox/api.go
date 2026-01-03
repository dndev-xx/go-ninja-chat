package outbox

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/dndev-xx/go-ninja-chat/internal/repositories/messages"
	eventstream "github.com/dndev-xx/go-ninja-chat/internal/services/event-stream"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

func (s *Outbox) Put(ctx context.Context, name, payload string, availableAt time.Time) (types.JobID, error) {
	s.mu.RLock()
	_, exists := s.registry[name]
	s.mu.RUnlock()

	if !exists {
		return types.JobIDNil, fmt.Errorf("job %q is not registered", name)
	}

	if availableAt.Before(time.Now()) {
		availableAt = time.Now()
	}

	jobID, err := s.repo.CreateJob(ctx, name, payload, availableAt)
	if err != nil {
		return types.JobIDNil, fmt.Errorf("failed to create job: %w", err)
	}

	s.logger.Debug("job created",
		zap.String("job_name", name),
		zap.String("job_id", jobID.String()),
		zap.Time("available_at", availableAt),
	)

	return jobID, nil
}

func (s *Outbox) PublishEvent(ctx context.Context, reqID types.RequestID, msg messages.Message) error {
	event := eventstream.NewMessageSentEventWithDetails(types.EventID(reqID), reqID, msg.ChatID, msg.ID, msg.AuthorID, msg.CreatedAt, msg.Body, msg.IsService)
	s.eventStream.Publish(ctx, msg.AuthorID, event)
	return nil
}
