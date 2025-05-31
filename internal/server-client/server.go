package serverclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	oapimdlwr "github.com/oapi-codegen/echo-middleware"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	mw "github.com/dndev-xx/go-ninja-chat/internal/middlewares"
	clientv1 "github.com/dndev-xx/go-ninja-chat/internal/server-client/v1/pkg"
)

const (
	readHeaderTimeout = time.Second
	shutdownTimeout   = 3 * time.Second
)

//go:generate options-gen -out-filename=server_options.gen.go -from-struct=Options
type Options struct {
	logger         *zap.Logger              `option:"mandatory"`
	addr           string                   `option:"mandatory" validate:"hostname_port"`
	allowOrigins   []string                 `option:"mandatory"`
	v1Swagger      *openapi3.T              `option:"mandatory"`
	v1Handlers     clientv1.ServerInterface `option:"mandatory"`
	keycloakClient mw.Introspector          `option:"mandatory"`
	resource       string                   `option:"mandatory"`
	role           string                   `option:"mandatory"`
	errorHandler   echo.HTTPErrorHandler    `option:"mandatory"`
}

type Server struct {
	lg  *zap.Logger
	srv *http.Server
	e   *echo.Echo
}

func New(opts Options) (*Server, error) {
	e := echo.New()
	lg := opts.logger
	e.HideBanner = true
	e.HidePort = true
	e.HTTPErrorHandler = opts.errorHandler

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: opts.allowOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			"X-Request-ID",
			"X-CSRF-Token",
		},
		ExposeHeaders:    []string{"X-Custom-Header"},
		AllowCredentials: true,
		MaxAge:           3600,
	}))
	authMiddleware := mw.NewKeycloakTokenAuth(opts.keycloakClient, opts.resource, opts.role)
	loggerMiddleware := mw.NewRequestLogger(lg)
	recoverLog := mw.NewRecovery(lg)

	v1 := e.Group("/v1",
		// mw.JSONResponseMiddleware(),
		loggerMiddleware,
		recoverLog,
		authMiddleware,
		oapimdlwr.OapiRequestValidatorWithOptions(opts.v1Swagger, &oapimdlwr.Options{
			Options: openapi3filter.Options{
				ExcludeRequestBody:  false,
				ExcludeResponseBody: true,
				AuthenticationFunc:  openapi3filter.NoopAuthenticationFunc,
			},
			SilenceServersWarning: true,
		}))

	clientv1.RegisterHandlers(v1, opts.v1Handlers)

	srv := &http.Server{
		Addr:              opts.addr,
		Handler:           e,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	return &Server{
		lg:  lg,
		srv: srv,
		e:   e,
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		return s.srv.Shutdown(ctx) //nolint:contextcheck // graceful shutdown with new context
	})

	eg.Go(func() error {
		s.lg.Info("client-server", zap.String("addr", s.srv.Addr))

		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen and serve: %v", err)
		}
		return nil
	})

	return eg.Wait()
}
