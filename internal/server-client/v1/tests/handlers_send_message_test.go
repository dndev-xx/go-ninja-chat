package v1_test

import (
	"fmt"
	"net/http"
	"strings"

	internalerrors "github.com/dndev-xx/go-ninja-chat/internal/errors"
	clientv1 "github.com/dndev-xx/go-ninja-chat/internal/server-client/v1/pkg"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

func (s *HandlersSuite) TestSendMessage_BindRequestError() {
	// Arrange.
	reqID := types.NewRequestID()
	resp, eCtx := s.newEchoCtx(reqID, "/v1/sendMessage", `{"messageBody": "Hel"`)
	// Action.
	err := s.handlers.PostSendMessage(eCtx, clientv1.PostSendMessageParams{XRequestID: reqID})

	// Assert.
	s.T().Log(err.Error())
	s.Require().Error(err)
	s.Equal(http.StatusBadRequest, internalerrors.GetServerErrorCode(err))
	s.Empty(resp.Body)
}

func (s *HandlersSuite) TestSendMessage_BindRequestEmptyMsgError() {
	// Arrange.
	reqID := types.NewRequestID()
	resp, eCtx := s.newEchoCtx(reqID, "/v1/sendMessage", `{"messageBody": ""}`)

	// Action.
	err := s.handlers.PostSendMessage(eCtx, clientv1.PostSendMessageParams{XRequestID: reqID})

	// Assert.
	s.T().Log(err.Error())
	s.Require().Error(err)
	s.Equal(http.StatusBadRequest, internalerrors.GetServerErrorCode(err))
	s.Empty(resp.Body)
}

func (s *HandlersSuite) TestSendMessage_BindRequestMoreMsgError() {
	// Arrange.
	reqID := types.NewRequestID()
	bigMsg := strings.Repeat("!", 4001)
	resp, eCtx := s.newEchoCtx(reqID, "/v1/sendMessage", fmt.Sprintf(`{"messageBody": "%s"}`, bigMsg))

	// Action.
	err := s.handlers.PostSendMessage(eCtx, clientv1.PostSendMessageParams{XRequestID: reqID})

	// Assert.
	s.T().Log(err.Error())
	s.Require().Error(err)
	s.Equal(http.StatusBadRequest, internalerrors.GetServerErrorCode(err))
	s.Empty(resp.Body)
}
