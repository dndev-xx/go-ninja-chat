package serverdebug

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/pprof"
	"sync"
	"time"

	"github.com/dndev-xx/go-ninja-chat/internal/buildinfo"
	"github.com/dndev-xx/go-ninja-chat/internal/logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mssola/useragent"
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

type Server struct {
	lg  *zap.Logger
	srv *http.Server
	logBuffer *logBuffer
}

type logBuffer struct {
	buf    *bytes.Buffer
	maxLen int
	mu     sync.Mutex
}

func newLogBuffer(maxLen int) *logBuffer {
	return &logBuffer{
		buf:    bytes.NewBuffer(nil),
		maxLen: maxLen,
	}
}

func (lb *logBuffer) Write(p []byte) (int, error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if lb.buf.Len()+len(p) > lb.maxLen {
		lb.buf.Reset()
	}

	return lb.buf.Write(p)
}

func (lb *logBuffer) String() string {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	return lb.buf.String()
}

func newLogBufferCore(lb *logBuffer, encoder zapcore.Encoder, enab zapcore.LevelEnabler) zapcore.Core {
	return zapcore.NewCore(encoder, zapcore.AddSync(lb), enab)
}

func New(opts Options) (*Server, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validate options: %v", err)
	}

	logBuf := newLogBuffer(1 * 1024 * 1024)

		// Получаем глобальный core
	globalCore := zap.L().Core()

		// Создаем encoder для буфера
	encoderCfg := zap.NewDevelopmentEncoderConfig()
	bufferEncoder := zapcore.NewConsoleEncoder(encoderCfg)

		// Создаем core для записи в буфер
	bufferCore := newLogBufferCore(logBuf, bufferEncoder, zapcore.DebugLevel)

	// Объединяем core глобального логгера и буфера
	dual := &dualCore{
		core1: globalCore,
		core2: bufferCore,
	}
	lg := zap.New(dual).Named("server-debug")

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(LoggerMiddleware(lg))

	s := &Server{
		lg:        lg,
		logBuffer: logBuf,
		srv: &http.Server{
			Addr:              opts.addr,
			Handler:           e,
			ReadHeaderTimeout: readHeaderTimeout,
		},
	}
	index := newIndexPage()

	e.GET("/version", s.Version)
	index.addPage("/version", "Get build information")

	e.PUT("/log/level", echo.WrapHandler(logger.Level))
	e.GET("/log/level", echo.WrapHandler(logger.Level))
	e.GET("/schema/client", s.getOpenAPISpec)
	index.addPage("/schema/client", "Get specification")
	e.GET("/logs", s.AllLogs)
	index.addPage("/logs", "Get all application logs")
	{
		pprofMux := http.NewServeMux()
		pprofMux.HandleFunc("/debug/pprof/", pprof.Index)
		pprofMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		pprofMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		pprofMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		pprofMux.HandleFunc("/debug/pprof/trace", pprof.Trace)

		e.GET("/debug/pprof/*", echo.WrapHandler(pprofMux))
		index.addPage("/debug/pprof/", "Go std profiler")
		index.addPage("/debug/pprof/profile?seconds=30", "Take half-min profile")
	}

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
		s.lg.Info("debug-server", zap.String("addr", s.srv.Addr))

		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen and serve: %v", err)
		}
		return nil
	})

	return eg.Wait()
}

func (s *Server) Version(eCtx echo.Context) error {
	return eCtx.JSON(http.StatusOK, buildinfo.BuildInfo)
}

func LoggerMiddleware(logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.Request().RemoteAddr
			method := c.Request().Method
			path := c.Request().URL.Path
			userAgent := c.Request().UserAgent()
			ua := useragent.New(userAgent)
			browserName, _ := ua.Browser()
			logger.Info("Incoming request debug server",
				zap.String("ip", ip),
				zap.String("method", method),
				zap.String("path", path),
				zap.String("os", ua.OS()),
				zap.String("browser", browserName),
			)
			return next(c)
		}
	}
}
