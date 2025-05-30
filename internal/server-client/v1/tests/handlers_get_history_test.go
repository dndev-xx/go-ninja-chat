package v1_test

// import (
// 	"net/http"

// 	internalerrors "github.com/dndev-xx/go-ninja-chat/internal/errors"
// 	clientv1 "github.com/dndev-xx/go-ninja-chat/internal/server-client/v1/pkg"
// 	"github.com/dndev-xx/go-ninja-chat/internal/types"
// )

// func (s *HandlersSuite) TestGetHistory_BindRequestError() {
// 	// Arrange.
// 	reqID := types.NewRequestID()
// 	resp, eCtx := s.newEchoCtx(reqID, "/v1/getHistory", `{"page_size":`)

// 	// Action.
// 	err := s.handlers.PostGetHistory(eCtx, clientv1.PostGetHistoryParams{XRequestID: reqID})

// 	// Assert.
// 	s.Require().Error(err)
// 	s.Equal(http.StatusBadRequest, internalerrors.GetServerErrorCode(err))
// 	s.Empty(resp.Body)
// }

// func (s *HandlersSuite) TestGetHistory_Usecase_InvalidRequest() {
// 	// Arrange.
// 	reqID := types.NewRequestID()
// 	resp, eCtx := s.newEchoCtx(reqID, "/v1/getHistory", `{"pageSize":9}`)
// 	s.getHistoryUseCase.EXPECT().Handle(eCtx.Request().Context(), gethistory.Request{
// 		ID:       reqID,
// 		ClientID: s.clientID,
// 		PageSize: 9,
// 		Cursor:   "",
// 	}).Return(gethistory.Response{}, gethistory.ErrInvalidRequest)

// 	// Action.
// 	err := s.handlers.PostGetHistory(eCtx, clientv1.PostGetHistoryParams{XRequestID: reqID})

// 	// Assert.
// 	s.Require().Error(err)
// 	s.Equal(http.StatusBadRequest, internalerrors.GetServerErrorCode(err))
// 	s.Empty(resp.Body)
// }

// func (s *HandlersSuite) TestGetHistory_Usecase_InvalidCursor() {
// 	// Arrange.
// 	reqID := types.NewRequestID()
// 	resp, eCtx := s.newEchoCtx(reqID, "/v1/getHistory", `{"cursor":"abracadabra"}`)
// 	s.getHistoryUseCase.EXPECT().Handle(eCtx.Request().Context(), gethistory.Request{
// 		ID:       reqID,
// 		ClientID: s.clientID,
// 		PageSize: 0,
// 		Cursor:   "abracadabra",
// 	}).Return(gethistory.Response{}, gethistory.ErrInvalidCursor)

// 	// Action.
// 	err := s.handlers.PostGetHistory(eCtx, clientv1.PostGetHistoryParams{XRequestID: reqID})

// 	// Assert.
// 	s.Require().Error(err)
// 	s.Equal(http.StatusBadRequest, internalerrors.GetServerErrorCode(err))
// 	s.Empty(resp.Body)
// }

// func (s *HandlersSuite) TestGetHistory_Usecase_UnknownError() {
// 	// Arrange.
// 	reqID := types.NewRequestID()
// 	resp, eCtx := s.newEchoCtx(reqID, "/v1/getHistory", `{"pageSize":10}`)
// 	s.getHistoryUseCase.EXPECT().Handle(eCtx.Request().Context(), gethistory.Request{
// 		ID:       reqID,
// 		ClientID: s.clientID,
// 		PageSize: 10,
// 	}).Return(gethistory.Response{}, errors.New("something went wrong"))

// 	// Action.
// 	err := s.handlers.PostGetHistory(eCtx, clientv1.PostGetHistoryParams{XRequestID: reqID})

// 	// Assert.
// 	s.Require().Error(err)
// 	s.Empty(resp.Body)
// }

// func (s *HandlersSuite) TestGetHistory_Usecase_Success() {
// 	// Arrange.
// 	reqID := types.NewRequestID()
// 	resp, eCtx := s.newEchoCtx(reqID, "/v1/getHistory", `{"pageSize":10}`)

// 	msgs := []gethistory.Message{
// 		{
// 			ID:         types.NewMessageID(),
// 			AuthorID:   types.NewUserID(),
// 			Body:       "hello!",
// 			CreatedAt:  time.Unix(1, 1).UTC(),
// 			IsReceived: true,
// 			IsBlocked:  false,
// 			IsService:  false,
// 		},
// 		{
// 			ID:         types.NewMessageID(),
// 			AuthorID:   types.UserIDNil,
// 			Body:       "service message",
// 			CreatedAt:  time.Unix(2, 2).UTC(),
// 			IsReceived: true,
// 			IsBlocked:  false,
// 			IsService:  true,
// 		},
// 	}
// 	s.getHistoryUseCase.EXPECT().Handle(eCtx.Request().Context(), gethistory.Request{
// 		ID:       reqID,
// 		ClientID: s.clientID,
// 		PageSize: 10,
// 	}).Return(gethistory.Response{
// 		Messages:   msgs,
// 		NextCursor: "",
// 	}, nil)

// 	// Action.
// 	err := s.handlers.PostGetHistory(eCtx, clientv1.PostGetHistoryParams{XRequestID: reqID})

// 	// Assert.
// 	s.Require().NoError(err)
// 	s.Equal(http.StatusOK, resp.Code)
// 	s.JSONEq(fmt.Sprintf(`
// {
//     "data":
//     {
//         "messages":
//         [
//             {
//                 "authorId": %q,
//                 "body": "hello!",
//                 "createdAt": "1970-01-01T00:00:01.000000001Z",
//                 "id": %q,
//                 "isBlocked": false,
//                 "isReceived": true,
//                 "isService": false
//             },
//             {
//                 "body": "service message",
//                 "createdAt": "1970-01-01T00:00:02.000000002Z",
//                 "id": %q,
//                 "isBlocked": false,
//                 "isReceived": true,
//                 "isService": true
//             }
//         ],
//         "next": ""
//     }
// }`, msgs[0].AuthorID, msgs[0].ID, msgs[1].ID), resp.Body.String())
// }
