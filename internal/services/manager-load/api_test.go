package managerload_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	managerload "github.com/dndev-xx/go-ninja-chat/internal/services/manager-load"
	managerloadmocks "github.com/dndev-xx/go-ninja-chat/internal/services/manager-load/mocks"
	"github.com/dndev-xx/go-ninja-chat/internal/testingh"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

// FIXME: Реализуй тесты на метод CanManagerTakeProblem:
// FIXME: 1) Нагенерировать моки.
// FIXME: 2) Покрыть граничные случаи между opts.maxProblemsAtTime и problemsRepo.GetManagerOpenProblemsCount.
// FIXME: 3) Какой результат метода, если репа возвращает ошибку?
// FIXME: 4) Something else?

type ServiceSuite struct {
	testingh.ContextSuite

	ctrl *gomock.Controller

	problemsRepo *managerloadmocks.MockproblemsRepository
	managerLoad  *managerload.Service
}

func TestServiceSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ServiceSuite))
}

func (s *ServiceSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.problemsRepo = managerloadmocks.NewMockproblemsRepository(s.ctrl)

	var err error
	s.managerLoad, err = managerload.New(managerload.NewOptions(
		10,
		s.problemsRepo,
	))
	s.Require().NoError(err)
	s.Require().NotNil(s.managerLoad)
	s.Require().NotNil(s.problemsRepo)

	s.ContextSuite.SetupTest()
}

func (s *ServiceSuite) TearDownTest() {
	s.ctrl.Finish()

	s.ContextSuite.TearDownTest()
}

func (s *ServiceSuite) TestCanManagerTakeProblem_WhenBelowLimit() {
	managerID := types.NewUserID()
	ctx := context.Background()

	s.problemsRepo.EXPECT().
		GetManagerOpenProblemsCount(ctx, managerID).
		Return(9, nil)

	canTake, err := s.managerLoad.CanManagerTakeProblem(ctx, managerID)
	s.Require().NoError(err)
	s.True(canTake)
}

func (s *ServiceSuite) TestCanManagerTakeProblem_WhenAtLimit() {
	managerID := types.NewUserID()
	ctx := context.Background()

	s.problemsRepo.EXPECT().
		GetManagerOpenProblemsCount(ctx, managerID).
		Return(10, nil)

	canTake, err := s.managerLoad.CanManagerTakeProblem(ctx, managerID)
	s.Require().NoError(err)
	s.False(canTake)
}

func (s *ServiceSuite) TestCanManagerTakeProblem_WhenAboveLimit() {
	managerID := types.NewUserID()
	ctx := context.Background()

	s.problemsRepo.EXPECT().
		GetManagerOpenProblemsCount(ctx, managerID).
		Return(11, nil)

	canTake, err := s.managerLoad.CanManagerTakeProblem(ctx, managerID)
	s.Require().NoError(err)
	s.False(canTake)
}

func (s *ServiceSuite) TestCanManagerTakeProblem_WhenRepoReturnsError() {
	managerID := types.NewUserID()
	ctx := context.Background()
	expectedErr := errors.New("database error")

	s.problemsRepo.EXPECT().
		GetManagerOpenProblemsCount(ctx, managerID).
		Return(0, expectedErr)

	canTake, err := s.managerLoad.CanManagerTakeProblem(ctx, managerID)
	s.Require().Error(err)
	s.Require().EqualError(err, "get manager open problems count: database error")
	s.False(canTake)
}

func (s *ServiceSuite) TestCanManagerTakeProblem_WhenZeroMaxProblems() {
	service, err := managerload.New(managerload.NewOptions(
		1,
		s.problemsRepo,
	))
	s.Require().NoError(err)

	managerID := types.NewUserID()
	ctx := context.Background()

	s.problemsRepo.EXPECT().
		GetManagerOpenProblemsCount(ctx, managerID).
		Return(1, nil)

	canTake, err := service.CanManagerTakeProblem(ctx, managerID)
	s.Require().NoError(err)
	s.False(canTake)
}
