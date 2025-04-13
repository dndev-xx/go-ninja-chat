package testingh

import (
	"context"

	"github.com/stretchr/testify/suite"
)

type ContextSuite struct {
	suite.Suite

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
	cs.Ctx, cs.ctxCancel = context.WithCancel(cs.SuiteCtx)
}

func (cs *ContextSuite) TearDownTest() {
	cs.ctxCancel()
}
