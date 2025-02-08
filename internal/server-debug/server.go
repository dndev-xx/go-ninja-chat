package serverdebug

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/dndev-xx/go-ninja-chat/internal/logger"
	"github.com/dndev-xx/go-ninja-chat/internal/buildinfo"
	"github.com/dndev-xx/go-ninja-chat/internal/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
)

const (
	readHeaderTimeout = time.Second
	shutdownTimeout   = 3 * time.Second
)

//go:generate options-gen -out-filename=server_options.gen.go -from-struct=Options
type Options struct {
	addr string `option:"mandatory" validate:"required,hostname_port"`
}

var defaultOptions = Options{
	addr: ":8080",
}

type Option func(*Options)

type Server struct {
	lg  *zap.Logger
	srv *http.Server
}

func NewOptions(addr string, opts ...Option) Options {
	options := defaultOptions
	options.addr = addr

	for _, opt := range opts {
		opt(&options)
	}

	return options
}

func New(opts Options) (*Server, error) {
	if err := validator.Validator.Struct(opts); err != nil {
		return nil, fmt.Errorf("validate options: %w", err)
	}

	lg := zap.L().Named("server-debug")

	e := echo.New()
	e.Use(middleware.Recover())

	s := &Server{
		lg: lg,
		srv: &http.Server{
			Addr:              opts.addr,
			Handler:           e,
			ReadHeaderTimeout: readHeaderTimeout,
		},
	}
	index := newIndexPage()

	e.GET("/version", s.Version)
	index.addPage("/version", "Get build information")
	e.PUT("/log/level", s.logLevelHandler)
	index.addPage("/log/level", "Change log level (PUT)")
	e.GET("/log/level", s.getLogLevelHandler)
	index.addPage("/log/level", "Get current log level (GET)")
	s.setupPprof(e, index)

	e.GET("/", index.handler)
	return s, nil
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
		s.lg.Info("listen and serve", zap.String("addr", s.srv.Addr))

		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen and serve: %v", err)
		}
		return nil
	})

	return eg.Wait()
}

func (s *Server) Version(c echo.Context) error {
	info := buildinfo.BuildInfo

	response := map[string]interface{}{
		"go_version": info.GoVersion,
		"path":       info.Path,
		"main": map[string]string{
			"path":    info.Main.Path,
			"version": info.Main.Version,
			"sum":     info.Main.Sum,
		},
		"dependencies": []map[string]string{},
		"settings":     []map[string]string{},
	}

	for _, dep := range info.Deps {
		depInfo := map[string]string{
			"path":    dep.Path,
			"version": dep.Version,
			"sum":     dep.Sum,
		}
		if dep.Replace != nil {
			depInfo["replace"] = fmt.Sprintf("%s@%s", dep.Replace.Path, dep.Replace.Version)
		}
		response["dependencies"] = append(response["dependencies"].([]map[string]string), depInfo)
	}

	for _, setting := range info.Settings {
		response["settings"] = append(response["settings"].([]map[string]string), map[string]string{
			"key":   setting.Key,
			"value": setting.Value,
		})
	}
	return c.JSON(http.StatusOK, response)
}

func (s *Server) getLogLevelHandler(c echo.Context) error {
	currentLevel := zap.L().Level()
	return c.JSON(http.StatusOK, map[string]string{"level": currentLevel.String()})
}

func (s *Server) logLevelHandler(c echo.Context) error {
	var req struct {
		Level string `json:"level"`
	}

	if err := c.Bind(&req); err != nil || req.Level == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	if _, err := zapcore.ParseLevel(req.Level); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid level"})
	}
	opts := logger.NewOptions(req.Level)

	if err := logger.Init(opts); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to change log level"})
	}
	zap.L().Named("change-log-level").Info("log level changed", zap.String("level", req.Level))

	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) setupPprof(e *echo.Echo, index *indexPage) {
	e.GET("/debug/pprof/", echo.WrapHandler(http.HandlerFunc(pprof.Index)))
	e.GET("/debug/pprof/cmdline", echo.WrapHandler(http.HandlerFunc(pprof.Cmdline)))
	e.GET("/debug/pprof/profile", echo.WrapHandler(http.HandlerFunc(pprof.Profile)))
	e.GET("/debug/pprof/symbol", echo.WrapHandler(http.HandlerFunc(pprof.Symbol)))
	e.GET("/debug/pprof/trace", echo.WrapHandler(http.HandlerFunc(pprof.Trace)))
	e.GET("/debug/pprof/:profile", echo.WrapHandler(http.HandlerFunc(pprof.Index)))

	index.addPage("/debug/pprof/", "pprof index")
	index.addPage("/debug/pprof/cmdline", "pprof cmdline")
	index.addPage("/debug/pprof/profile", "pprof profile")
	index.addPage("/debug/pprof/symbol", "pprof symbol")
	index.addPage("/debug/pprof/trace", "pprof trace")
}
