package test

import (
	"errors"
	"net/http"

	managerv1 "github.com/dndev-xx/go-ninja-chat/internal/server-manager/v1/pkg"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
	fhands "github.com/dndev-xx/go-ninja-chat/internal/usecase/manager/get-free-hands"
	canreceiveproblems "github.com/dndev-xx/go-ninja-chat/internal/usecase/manager/getFreeHandsBtnAvailability"
)

func (s *HandlersSuite) TestGetFreeHandsBtnAvailability_Usecase_Error() {
	// Arrange.
	reqID := types.NewRequestID()
	resp, eCtx := s.newEchoCtx(reqID, "/v1/getFreeHandsBtnAvailability", "")
	s.getFreeHandsBtnAvailabilityUseCase.EXPECT().Handle(eCtx.Request().Context(), canreceiveproblems.Request{
		ID:        reqID,
		ManagerID: s.managerID,
	}).Return(canreceiveproblems.Response{}, errors.New("something went wrong"))

	// Action.
	err := s.handlers.PostGetFreeHandsBtnAvailability(eCtx, managerv1.PostGetFreeHandsBtnAvailabilityParams{XRequestID: reqID})

	// Assert.
	s.Require().Error(err)
	s.Empty(resp.Body)
}

func (s *HandlersSuite) TestGetFreeHandsBtnAvailability_Usecase_Success() {
	// Arrange.
	reqID := types.NewRequestID()
	resp, eCtx := s.newEchoCtx(reqID, "/v1/getFreeHandsBtnAvailability", "")
	s.getFreeHandsBtnAvailabilityUseCase.EXPECT().Handle(eCtx.Request().Context(), canreceiveproblems.Request{
		ID:        reqID,
		ManagerID: s.managerID,
	}).Return(canreceiveproblems.Response{Result: true}, nil)

	// Action.
	err := s.handlers.PostGetFreeHandsBtnAvailability(eCtx, managerv1.PostGetFreeHandsBtnAvailabilityParams{XRequestID: reqID})

	// Assert.
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.Code)
	s.JSONEq(`
{
    "data":
    {
        "available": true
    }
}`, resp.Body.String())
}

func (s *HandlersSuite) TestGetFreeHands_Usecase_Success() {
	// Arrange.
	reqID := types.NewRequestID()
	resp, eCtx := s.newEchoCtx(reqID, "/v1/freeHands", "")
	s.getFreeHands.EXPECT().Handle(eCtx.Request().Context(), fhands.Request{
		ManagerID: s.managerID,
	}).Return(nil)

	// Action.
	err := s.handlers.PostFreeHands(eCtx, managerv1.PostFreeHandsParams{XRequestID: reqID})

	// Assert.
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.Code)
	s.JSONEq(`
{
    "data":null
}`, resp.Body.String())
}
