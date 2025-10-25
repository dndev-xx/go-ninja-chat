package inmemeventstream

import (
	"context"
	"sync"

	eventstream "github.com/dndev-xx/go-ninja-chat/internal/services/event-stream"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

const (
	serviceName = "event-stream"
	chSubLen    = 100
)

type Service struct {
	Name string
	mu   sync.RWMutex
	// Map of userID to list of channels for that user
	subscribers map[types.UserID][]chan eventstream.Event
	// Map to track if channel is closed
	closedChannels map[chan eventstream.Event]bool
	closed         bool
}

func New() *Service {
	return &Service{
		Name:           serviceName,
		subscribers:    make(map[types.UserID][]chan eventstream.Event),
		closedChannels: make(map[chan eventstream.Event]bool),
	}
}

func (s *Service) Subscribe(ctx context.Context, userID types.UserID) (<-chan eventstream.Event, error) {
	s.mu.Lock()
	ch := make(chan eventstream.Event, chSubLen)
	s.subscribers[userID] = append(s.subscribers[userID], ch)
	s.closedChannels[ch] = false
	s.mu.Unlock()

	go func() {
		<-ctx.Done()
		s.unsubscribe(userID, ch)
	}()

	return ch, nil
}

func (s *Service) unsubscribe(userID types.UserID, ch chan eventstream.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	subscribers, exists := s.subscribers[userID]
	if !exists {
		return
	}
	loop:
		for i, subscriber := range subscribers {
			if subscriber == ch {
				s.closedChannels[ch] = true
				close(ch)

				s.subscribers[userID] = append(subscribers[:i], subscribers[i+1:]...)

				if len(s.subscribers[userID]) == 0 {
					delete(s.subscribers, userID)
				}

				delete(s.closedChannels, ch)
				break loop
			}
		}
}

func (s *Service) Publish(ctx context.Context, userID types.UserID, event eventstream.Event) error {
	if err := event.Validate(); err != nil {
		return err
	}

	s.mu.RLock()
	subs, exists := s.subscribers[userID]
	if !exists || len(subs) == 0 {
		s.mu.RUnlock()
		return nil
	}

	subsCopy := make([]chan eventstream.Event, 0, len(subs))
	for _, ch := range subs {
		if !s.closedChannels[ch] {
			subsCopy = append(subsCopy, ch)
		}
	}
	s.mu.RUnlock()

	for _, ch := range subsCopy {
		select {
		case <-ctx.Done():
			return nil
		case ch <- event:
			// Successfully sent to this subscriber
		default:
			// Channel is full, skip this subscriber to prevent blocking
		}
	}

	return nil
}

func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}
	s.closed = true

	for userID, subs := range s.subscribers {
		for _, ch := range subs {
			if !s.closedChannels[ch] {
				close(ch)
			}
		}
		delete(s.subscribers, userID)
	}

	s.closedChannels = make(map[chan eventstream.Event]bool)

	return nil
}
