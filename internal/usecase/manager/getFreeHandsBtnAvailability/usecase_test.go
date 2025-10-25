package getfreehandsbtnavailability_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/dndev-xx/go-ninja-chat/internal/testingh"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
	canreceiveproblems "github.com/dndev-xx/go-ninja-chat/internal/usecase/manager/getFreeHandsBtnAvailability"
	canreceiveproblemsmocks "github.com/dndev-xx/go-ninja-chat/internal/usecase/manager/getFreeHandsBtnAvailability/mocks"
)

// FIXME: Реализуй тесты на юзкейс:
// FIXME: 1) Нагенерировать моки.
// FIXME: 2) Кейсы:
// FIXME:	- невалидный запрос
// FIXME:	- ошибка при вызове managerPool.Contains
// FIXME:	- менеджер есть в пуле
// FIXME:	- ошибка при вызове managerLoadService.CanManagerTakeProblem
// FIXME:	- CanManagerTakeProblem завершается успешно с разным результатом
// FIXME:	- something else?

type UseCaseSuite struct {
	testingh.ContextSuite

	ctrl      *gomock.Controller
	mLoadMock *canreceiveproblemsmocks.MockmanagerLoadService
	mPoolMock *canreceiveproblemsmocks.MockmanagerPool
	uCase     *canreceiveproblems.UseCase
}

func TestUseCaseSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(UseCaseSuite))
}

func (s *UseCaseSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mLoadMock = canreceiveproblemsmocks.NewMockmanagerLoadService(s.ctrl)
	s.mPoolMock = canreceiveproblemsmocks.NewMockmanagerPool(s.ctrl)
	var err error
	s.uCase, err = canreceiveproblems.New(
		canreceiveproblems.NewOptions(
			canreceiveproblems.WithManagerLoadService(s.mLoadMock),
			canreceiveproblems.WithManagerPool(s.mPoolMock),
		))
	s.Require().NoError(err)

	s.ContextSuite.SetupTest()
}

func (s *UseCaseSuite) TearDownTest() {
	s.ctrl.Finish()
	s.ContextSuite.TearDownTest()
}

func (s *UseCaseSuite) Test_GetFreeHandsBtnAvailability_InvalidRequest() {
	req := canreceiveproblems.Request{
		ID:        types.RequestIDNil,
		ManagerID: types.UserIDNil,
	}
	_, err := s.uCase.Handle(s.Ctx, req)
	s.Require().ErrorIs(err, canreceiveproblems.ErrInvalidRequest)
}

func (s *UseCaseSuite) Test_GetFreeHandsBtnAvailability_ValidRequest() {
	req := canreceiveproblems.Request{
		ID:        types.NewRequestID(),
		ManagerID: types.NewUserID(),
	}
	s.mPoolMock.EXPECT().Contains(s.Ctx, req.ManagerID).Return(false, nil)
	s.mLoadMock.EXPECT().CanManagerTakeProblem(s.Ctx, req.ManagerID).Return(true, nil)
	_, err := s.uCase.Handle(s.Ctx, req)
	s.Require().NoError(err, canreceiveproblems.ErrInvalidRequest)
}

func (s *UseCaseSuite) Test_GetFreeHandsBtnAvailability_Error_ManagerPoolContains() {
	req := canreceiveproblems.Request{
		ID:        types.NewRequestID(),
		ManagerID: types.NewUserID(),
	}
	s.mPoolMock.EXPECT().Contains(s.Ctx, req.ManagerID).Return(false, canreceiveproblems.ErrManagerNotExists)
	_, err := s.uCase.Handle(s.Ctx, req)
	s.Require().ErrorIs(err, canreceiveproblems.ErrManagerNotExists)
}

func (s *UseCaseSuite) Test_GetFreeHandsBtnAvailability_Error_ManagerLoadService() {
	req := canreceiveproblems.Request{
		ID:        types.NewRequestID(),
		ManagerID: types.NewUserID(),
	}
	s.mPoolMock.EXPECT().Contains(s.Ctx, req.ManagerID).Return(false, nil)
	s.mLoadMock.EXPECT().CanManagerTakeProblem(s.Ctx, req.ManagerID).Return(false, canreceiveproblems.ErrManagerBusy)
	_, err := s.uCase.Handle(s.Ctx, req)
	s.Require().ErrorIs(err, canreceiveproblems.ErrManagerBusy)
}

func (s *UseCaseSuite) Test_GetFreeHandsBtnAvailability_ManagerLoadService() {
	req := canreceiveproblems.Request{
		ID:        types.NewRequestID(),
		ManagerID: types.NewUserID(),
	}
	s.mPoolMock.EXPECT().Contains(s.Ctx, req.ManagerID).Return(false, nil)
	s.mLoadMock.EXPECT().CanManagerTakeProblem(s.Ctx, req.ManagerID).Return(true, nil)
	_, err := s.uCase.Handle(s.Ctx, req)
	s.Require().NoError(err)
	s.Require().True(true)
}

func (s *UseCaseSuite) Test_GetFreeHandsBtnAvailability_ManagerLoadService_Is_False() {
	req := canreceiveproblems.Request{
		ID:        types.NewRequestID(),
		ManagerID: types.NewUserID(),
	}
	s.mPoolMock.EXPECT().Contains(s.Ctx, req.ManagerID).Return(false, nil)
	s.mLoadMock.EXPECT().CanManagerTakeProblem(s.Ctx, req.ManagerID).Return(false, nil)
	_, err := s.uCase.Handle(s.Ctx, req)
	s.Require().NoError(err)
	s.Require().False(false)
}
