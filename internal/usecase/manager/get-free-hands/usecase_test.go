package getfreehands_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/dndev-xx/go-ninja-chat/internal/testingh"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
	u "github.com/dndev-xx/go-ninja-chat/internal/usecase/manager/get-free-hands"
	canreceiveproblemsmocks "github.com/dndev-xx/go-ninja-chat/internal/usecase/manager/get-free-hands/mocks"
)

type UseCaseSuite struct {
	testingh.ContextSuite
	context     context.Context
	ctrl        *gomock.Controller
	managerPool *canreceiveproblemsmocks.MockmanagerPool

	uCase *u.UseCase
}

func TestUseCaseSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(UseCaseSuite))
}

func (s *UseCaseSuite) SetupTest() {
	s.context = context.Background()
	s.ctrl = gomock.NewController(s.T())
	s.managerPool = canreceiveproblemsmocks.NewMockmanagerPool(s.ctrl)
	var err error
	s.uCase, err = u.New(
		u.NewOptions(
			u.WithManagerPool(s.managerPool),
		))
	s.Require().NoError(err)

	s.ContextSuite.SetupTest()
}

func (s *UseCaseSuite) TearDownTest() {
	s.ctrl.Finish()
	s.ContextSuite.TearDownTest()
}

func (s *UseCaseSuite) Test_Success_Put_Manager_To_Pool() {
	req := u.Request{
		ManagerID: types.NewUserID(),
	}
	s.managerPool.EXPECT().Put(s.context, req.ManagerID).Return(nil)
	err := s.uCase.Handle(s.context, req)
	s.Require().NoError(err)
}
