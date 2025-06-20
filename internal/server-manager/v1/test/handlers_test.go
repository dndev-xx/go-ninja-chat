package test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"

	"github.com/dndev-xx/go-ninja-chat/internal/middlewares"
	managerv1 "github.com/dndev-xx/go-ninja-chat/internal/server-manager/v1"
	managerv1mocks "github.com/dndev-xx/go-ninja-chat/internal/server-manager/v1/mocks"
	"github.com/dndev-xx/go-ninja-chat/internal/testingh"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

type HandlersSuite struct {
	testingh.ContextSuite

	ctrl                               *gomock.Controller
	getFreeHandsBtnAvailabilityUseCase *managerv1mocks.MockgetFreeHandsBtnAvailabilityUseCase
	getFreeHands                       *managerv1mocks.MockgetFreeHands
	handlers                           *managerv1.Handlers

	managerID types.UserID
}

func TestHandlersSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(HandlersSuite))
}

func (s *HandlersSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.getFreeHandsBtnAvailabilityUseCase = managerv1mocks.NewMockgetFreeHandsBtnAvailabilityUseCase(s.ctrl)
	s.getFreeHands = managerv1mocks.NewMockgetFreeHands(s.ctrl)
	{
		var err error
		s.handlers, err = managerv1.NewHandlers(managerv1.NewOptions(s.getFreeHandsBtnAvailabilityUseCase, s.getFreeHands))
		s.Require().NoError(err)
	}
	s.managerID = types.NewUserID()

	s.ContextSuite.SetupTest()
}

func (s *HandlersSuite) TearDownTest() {
	s.ctrl.Finish()

	s.ContextSuite.TearDownTest()
}

func (s *HandlersSuite) newEchoCtx(
	requestID types.RequestID,
	path string,
	body string,
) (*httptest.ResponseRecorder, echo.Context) {
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderXRequestID, requestID.String())

	// Добавляем токен в заголовок Authorization
	req.Header.Set(echo.HeaderAuthorization, "Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICI3cHpiVUdSYWtRWkRrVnlGbzZCOVZHbmxOUWJUM2ozT202NE9VV1V1UDlBIn0.eyJleHAiOjE3NDU3NzIxNzAsImlhdCI6MTc0NTc3MTg3MCwiYXV0aF90aW1lIjoxNzQ1NzcxNjEwLCJqdGkiOiI4OGIyYTY0OC0yMWM5LTQ0NGQtYTg5YS1mZDVkOTU4ODkwZDAiLCJpc3MiOiJodHRwOi8vbG9jYWxob3N0OjMwMTAvcmVhbG1zL0JhbmsiLCJhdWQiOiJhY2NvdW50Iiwic3ViIjoiYTk5NDY4ZDYtNzM4Yy00NTc2LTg2YjctOGQ1NGI2M2M5MWJjIiwidHlwIjoiQmVhcmVyIiwiYXpwIjoiY2hhdC11aS1jbGllbnQiLCJub25jZSI6IjhmZjE3NjJhLTIxNjAtNDdjNi05NWUwLWUyMTBhNDg1YjU4OCIsInNlc3Npb25fc3RhdGUiOiIxZmE3Nzk2Ni05MTFhLTQxNTQtYjY3NS02OTE5OTBmODE0MzkiLCJhY3IiOiIwIiwiYWxsb3dlZC1vcmlnaW5zIjpbIioiXSwicmVhbG1fYWNjZXNzIjp7InJvbGVzIjpbIm9mZmxpbmVfYWNjZXNzIiwiZGVmYXVsdC1yb2xlcy1iYW5rIiwidW1hX2F1dGhvcml6YXRpb24iXX0sInJlc291cmNlX2FjY2VzcyI6eyJjaGF0LXVpLWNsaWVudCI6eyJyb2xlcyI6WyJzdXBwb3J0LWNoYXQtY2xpZW50Il19LCJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19LCJzY29wZSI6Im9wZW5pZCBwcm9maWxlIGVtYWlsIiwic2lkIjoiMWZhNzc5NjYtOTExYS00MTU0LWI2NzUtNjkxOTkwZjgxNDM5IiwiZW1haWxfdmVyaWZpZWQiOmZhbHNlLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJib25kMDA3IiwiZ2l2ZW5fbmFtZSI6IiIsImZhbWlseV9uYW1lIjoiIiwiZW1haWwiOiJib25kMDA3QG1haWwucnUifQ.e5kQs7J6X43SaqJVzz6N3zh2ZYxv5-YrVf77jWPF4p3XrkoUpa6X8tHe8QqakzYDjceLuUeUTlbUomrpGzeESodJk8uFVAjov5P0ivNeKZ8zJytZ1vkYR7r764QyuFCDhhCqqW6EKukm_L1gCvtMiA9cjmQbjQ2Fl11Z-FWDZC_Rtf3HQqmOvJALWukXOFEvm7Xv50xP-JMGkP_h3u1Jt-QdDUqQR-jsSmDiYN6zja_FA26KPNhBh1arry_WF_bGrNwnIpsQIfoKBduwbOGHC21cy7w4ePfR3jsH1tXKSwFNhmV_S9weeOKkvcoCWkE6l3M_E3gbBCvoEhvgui4E6A")

	resp := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, resp)

	token := &jwt.Token{
		Claims: &middlewares.Claims{
			StandardClaims: &jwt.StandardClaims{
				Subject: s.managerID.String(),
			},
			ResourceAccess: map[string]struct {
				Roles []string `json:"roles"`
			}{
				"chat-ui-manager": {
					Roles: []string{"support-chat-manager"},
				},
			},
		},
	}
	ctx.Set("user-token", token)

	return resp, ctx
}
