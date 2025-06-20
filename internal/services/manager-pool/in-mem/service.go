package inmemmanagerpool

import (
	"container/list"
	"context"
	"errors"
	"sync"

	mem "github.com/dndev-xx/go-ninja-chat/internal/services/manager-pool"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

var (
	ErrNoAvailableManagers  = errors.New("no available managers")
	ErrManagerAlreadyExists = errors.New("manager already exists")
	ErrManagerNotFound      = errors.New("manager not found")
	ErrManagerExists        = errors.New("manager pool is full")
)

const (
	serviceName = "manager-pool"
	managersMax = 1000
)

type Service struct {
	queue   *list.List
	lookup  map[types.UserID]struct{}
	mu      sync.RWMutex
	name    string
	maxSize int
}

func New() *Service {
	return &Service{
		queue:   list.New(),
		lookup:  make(map[types.UserID]struct{}),
		name:    serviceName,
		maxSize: managersMax,
	}
}

func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.queue.Init()
	s.lookup = make(map[types.UserID]struct{})
	return nil
}

func (s *Service) Get(ctx context.Context) (types.UserID, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.queue.Len() == 0 {
		return types.UserIDNil, mem.ErrNoAvailableManagers
	}

	first := s.queue.Front()
	managerID := first.Value.(types.UserID)

	s.queue.Remove(first)
	delete(s.lookup, managerID)

	return managerID, nil
}

func (s *Service) Put(ctx context.Context, managerID types.UserID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.lookup[managerID]; exists {
		return nil
	}

	if s.queue.Len() >= s.maxSize {
		return ErrManagerExists
	}

	s.queue.PushBack(managerID)
	s.lookup[managerID] = struct{}{}
	return nil
}

func (s *Service) Contains(ctx context.Context, managerID types.UserID) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.lookup[managerID]
	return exists, nil
}

func (s *Service) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.queue.Len()
}
