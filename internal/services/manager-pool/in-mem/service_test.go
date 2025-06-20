package inmemmanagerpool_test

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"

	managerpool "github.com/dndev-xx/go-ninja-chat/internal/services/manager-pool"
	inmemmanagerpool "github.com/dndev-xx/go-ninja-chat/internal/services/manager-pool/in-mem"
	"github.com/dndev-xx/go-ninja-chat/internal/testingh"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

type ServiceSuite struct {
	testingh.ContextSuite
	pool managerpool.Pool
}

func TestServiceSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ServiceSuite))
}

func (s *ServiceSuite) SetupTest() {
	s.ContextSuite.SetupTest()
	s.pool = inmemmanagerpool.New()
}

func (s *ServiceSuite) TearDownTest() {
	s.NoError(s.pool.Close())
	s.ContextSuite.TearDownTest()
}

func (s *ServiceSuite) TestEmpty() {
	s.Equal(0, s.pool.Size())

	_, err := s.pool.Get(s.Ctx)
	s.Require().ErrorIs(err, managerpool.ErrNoAvailableManagers)

	contains, err := s.pool.Contains(s.Ctx, types.NewUserID())
	s.Require().NoError(err)
	s.False(contains)
}

func (s *ServiceSuite) TestFIFOLogic() {
	const managersNum = 10
	managers := make([]types.UserID, 0, managersNum)

	for i := 0; i < managersNum; i++ {
		m := types.NewUserID()
		managers = append(managers, m)

		s.T().Logf("%d: put %s", i, m)
		err := s.pool.Put(s.Ctx, m)
		s.Require().NoError(err)

		contains, err := s.pool.Contains(s.Ctx, m)
		s.Require().NoError(err)
		s.True(contains)
	}
	s.Len(managers, managersNum)
	s.Equal(managersNum, s.pool.Size())

	for i, m := range managers {
		mm, err := s.pool.Get(s.Ctx)
		s.Require().NoError(err)

		s.T().Logf("%d: got %s", i, m)
		s.Equal(m.String(), mm.String())
		s.Equal(len(managers)-i-1, s.pool.Size())

		contains, err := s.pool.Contains(s.Ctx, m)
		s.Require().NoError(err)
		s.False(contains)
	}
}

func (s *ServiceSuite) TestPut_Idempotency() {
	m := types.NewUserID()
	for i := 0; i < 3; i++ {
		err := s.pool.Put(s.Ctx, m)
		s.Require().NoError(err)
		s.Equal(1, s.pool.Size())

		contains, err := s.pool.Contains(s.Ctx, m)
		s.Require().NoError(err)
		s.True(contains)
	}

	mm, err := s.pool.Get(s.Ctx)
	s.Require().NoError(err)
	s.Equal(m.String(), mm.String())
	s.Equal(0, s.pool.Size())

	contains, err := s.pool.Contains(s.Ctx, m)
	s.Require().NoError(err)
	s.False(contains)
}

func (s *ServiceSuite) TestConcurrency() {
	const (
		managersNum = 100
		putInterval = 25 * time.Millisecond
		putWorkers  = 10
	)

	managers := make([]types.UserID, managersNum)
	for i := 0; i < managersNum; i++ {
		managers[i] = types.NewUserID()
	}
	randManager := func() types.UserID { return managers[rand.Int()%len(managers)] } //nolint:gosec // for test math/rand is OK

	ctx, cancel := context.WithTimeout(s.Ctx, time.Second)
	defer cancel()

	wg, ctx := errgroup.WithContext(ctx)

	wg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(putInterval):
				_ = s.pool.Size()
			}
		}
	})

	wg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(putInterval):
				_, _ = s.pool.Contains(ctx, randManager())
			}
		}
	})

	wg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(2 * putInterval):
				if _, err := s.pool.Get(ctx); err != nil && !errors.Is(err, managerpool.ErrNoAvailableManagers) {
					return err
				}
			}
		}
	})

	for i := 0; i < putWorkers; i++ {
		wg.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(putInterval):
					if err := s.pool.Put(ctx, randManager()); err != nil {
						return err
					}
				}
			}
		})
	}

	s.NoError(wg.Wait())
}
