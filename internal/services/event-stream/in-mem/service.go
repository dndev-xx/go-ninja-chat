package inmemeventstream

import (
	"context"
	"errors"
	"sync"

	eventstream "github.com/dndev-xx/go-ninja-chat/internal/services/event-stream"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

const (
	serviceName = "event-stream"
)

type subscriber struct {
	ch      chan eventstream.Event
	closing chan struct{}
}

type Service struct {
	Name        string
	mu          sync.RWMutex
	subscribers map[types.UserID][]*subscriber
	closed      bool
}

func New() *Service {
	return &Service{
		Name:        serviceName,
		subscribers: make(map[types.UserID][]*subscriber),
	}
}

func (s *Service) Subscribe(ctx context.Context, userID types.UserID) (<-chan eventstream.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, errors.New("service is closed")
	}
	// буфферизированный канал является ботлнеком, заранее не сможем угадать сколько сообщений попадет в канал, если > 100 возможно недетермированное поведение.
	ch := make(chan eventstream.Event, 100)
	closing := make(chan struct{})
	sub := &subscriber{ch: ch, closing: closing}
	s.subscribers[userID] = append(s.subscribers[userID], sub)

	go func() {
		select {
		case <-ctx.Done():
			s.unsubscribe(userID, sub)
		case <-closing:
			// Уже закрыто другим способом
		}
	}()

	return ch, nil
}

func (s *Service) unsubscribe(userID types.UserID, sub *subscriber) {
	s.mu.Lock()
	defer s.mu.Unlock()

	close(sub.closing)

	subscribers, exists := s.subscribers[userID]
	if !exists {
		return
	}

	for i, subscriber := range subscribers {
		if subscriber == sub {
			s.subscribers[userID] = append(subscribers[:i], subscribers[i+1:]...)

			close(sub.ch)

			if len(s.subscribers[userID]) == 0 {
				delete(s.subscribers, userID)
			}
			return
		}
	}
}

func (s *Service) Publish(ctx context.Context, userID types.UserID, event eventstream.Event) error {
	if err := event.Validate(); err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return errors.New("service is closed")
	}

	subscribers, exists := s.subscribers[userID]
	if !exists {
		return nil
	}

	// Отправляем события только активным подписчикам
	for _, sub := range subscribers {
		select {
		case <-sub.closing:
			// Пропускаем закрытые подписчики
			continue
		default:
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case sub.ch <- event:
			// Успешно отправлено
		default:
			// Пропускаем если канал заполнен
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

	for userID, subscribers := range s.subscribers {
		for _, sub := range subscribers {
			close(sub.closing)
			close(sub.ch)
		}
		delete(s.subscribers, userID)
	}

	return nil
}
