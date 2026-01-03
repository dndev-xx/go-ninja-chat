package testingh

import (
	"context"
	"sync"

	"github.com/stretchr/testify/suite"
)

type ContextSuite struct {
	suite.Suite

	mu        sync.RWMutex
	Ctx       context.Context
	ctxCancel context.CancelFunc

	SuiteCtx       context.Context
	suiteCtxCancel context.CancelFunc
}

func (cs *ContextSuite) SetupSuite() {
	cs.SuiteCtx, cs.suiteCtxCancel = context.WithCancel(context.Background())
}

func (cs *ContextSuite) TearDownSuite() {
	cs.suiteCtxCancel()
}

func (cs *ContextSuite) SetupTest() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.Ctx, cs.ctxCancel = context.WithCancel(cs.SuiteCtx)
}

func (cs *ContextSuite) TearDownTest() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if cs.ctxCancel != nil {
		cs.ctxCancel()
	}
}

func (cs *ContextSuite) Context() context.Context {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.Ctx
}
