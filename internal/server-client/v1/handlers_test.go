package v1_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"

	"github.com/dndev-xx/go-ninja-chat/internal/middlewares"
	clientv1 "github.com/dndev-xx/go-ninja-chat/internal/server-client/v1"
	clientv1mocks "github.com/dndev-xx/go-ninja-chat/internal/server-client/v1/mocks"
	"github.com/dndev-xx/go-ninja-chat/internal/testingh"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

type HandlersSuite struct {
	testingh.ContextSuite

	ctrl              *gomock.Controller
	getHistoryUseCase *clientv1mocks.MockgetHistoryUseCase
	handlers          clientv1.Handlers

	clientID types.UserID
}

func TestHandlersSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(HandlersSuite))
}

func (s *HandlersSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.getHistoryUseCase = clientv1mocks.NewMockgetHistoryUseCase(s.ctrl)
	{
		var err error
		s.handlers, err = clientv1.NewHandlers(clientv1.NewOptions(
			s.getHistoryUseCase,
		))
		s.Require().NoError(err)
	}
	s.clientID = types.NewUserID()

	s.ContextSuite.SetupTest()
}

func (s *HandlersSuite) TearDownTest() {
	s.ctrl.Finish()

	s.ContextSuite.TearDownTest()
}

func (s *HandlersSuite) newEchoCtx(
	requestID types.RequestID,
	path string, //nolint:unparam
	body string,
) (*httptest.ResponseRecorder, echo.Context) {
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderXRequestID, requestID.String())

	resp := httptest.NewRecorder()

	ctx := echo.New().NewContext(req, resp)
	middlewares.SetToken(ctx, s.clientID)

	return resp, ctx
}
