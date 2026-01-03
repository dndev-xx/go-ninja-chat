package v1_test

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	internalerrors "github.com/dndev-xx/go-ninja-chat/internal/errors"
	clientv1 "github.com/dndev-xx/go-ninja-chat/internal/server-client/v1/pkg"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
	sendmessage "github.com/dndev-xx/go-ninja-chat/internal/usecase/client/send-message"
)

func (s *HandlersSuite) TestSendMessage_BindRequestError() {
	// Arrange.
	reqID := types.NewRequestID()
	resp, eCtx := s.newEchoCtx(reqID, "/v1/sendMessage", `{"messageBody": "Hel"`)
	// Action.
	err := s.handlers.PostV1SendMessage(eCtx, clientv1.PostV1SendMessageParams{XRequestID: reqID})

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
	err := s.handlers.PostV1SendMessage(eCtx, clientv1.PostV1SendMessageParams{XRequestID: reqID})

	// Assert.
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
	err := s.handlers.PostV1SendMessage(eCtx, clientv1.PostV1SendMessageParams{XRequestID: reqID})

	// Assert.
	s.T().Log(err.Error())
	s.Require().Error(err)
	s.Equal(http.StatusBadRequest, internalerrors.GetServerErrorCode(err))
	s.Empty(resp.Body)
}

func (s *HandlersSuite) TestSendMessage_Usecase_ChatNotCreatedError() {
	// Arrange.
	reqID := types.NewRequestID()
	resp, eCtx := s.newEchoCtx(reqID, "/v1/sendMessage", `{"messageBody": "Hello!"}`)
	s.sendMsgUseCase.EXPECT().Handle(eCtx.Request().Context(), sendmessage.Request{
		ID:          reqID,
		ClientID:    s.clientID,
		MessageBody: "Hello!",
	}).Return(sendmessage.Response{}, sendmessage.ErrChatNotCreated)

	// Action.
	err := s.handlers.PostV1SendMessage(eCtx, clientv1.PostV1SendMessageParams{XRequestID: reqID})

	// Assert.
	s.Require().Error(err)
	s.EqualValues(clientv1.ErrorCodeInternalServerError, internalerrors.GetServerErrorCode(err))
	s.Empty(resp.Body)
}

func (s *HandlersSuite) TestSendMessage_Usecase_ProblemNotCreatedError() {
	// Arrange.
	reqID := types.NewRequestID()
	resp, eCtx := s.newEchoCtx(reqID, "/v1/sendMessage", `{"messageBody": "Hello!"}`)
	s.sendMsgUseCase.EXPECT().Handle(eCtx.Request().Context(), sendmessage.Request{
		ID:          reqID,
		ClientID:    s.clientID,
		MessageBody: "Hello!",
	}).Return(sendmessage.Response{}, sendmessage.ErrProblemNotCreated)

	// Action.
	err := s.handlers.PostV1SendMessage(eCtx, clientv1.PostV1SendMessageParams{XRequestID: reqID})

	// Assert.
	s.Require().Error(err)
	s.EqualValues(clientv1.ErrorCodeInternalServerError, internalerrors.GetServerErrorCode(err))
	s.Empty(resp.Body)
}

func (s *HandlersSuite) TestSendMessage_Usecase_Success() {
	// Arrange.
	reqID := types.NewRequestID()
	msgID := types.NewMessageID()

	resp, eCtx := s.newEchoCtx(reqID, "/v1/sendMessage", `{"messageBody": "Hello!"}`)
	s.sendMsgUseCase.EXPECT().Handle(eCtx.Request().Context(), sendmessage.Request{
		ID:          reqID,
		ClientID:    s.clientID,
		MessageBody: "Hello!",
	}).Return(sendmessage.Response{
		AuthorID:  s.clientID,
		MessageID: msgID,
		CreatedAt: time.Unix(1, 1).UTC(),
	}, nil)

	// Action.
	err := s.handlers.PostV1SendMessage(eCtx, clientv1.PostV1SendMessageParams{XRequestID: reqID})

	// Assert.
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.Code)
	s.JSONEq(fmt.Sprintf(`
{
    "data":
    {
        "authorID": "%s",
        "createdAt": "1970-01-01T00:00:01.000000001Z",
        "messageID": "%s"
    }
}`, s.clientID, msgID), resp.Body.String())
}
