package context

import (
	"flag"
	"fmt"

	"github.com/dndev-xx/go-ninja-chat/internal/config"
	"github.com/dndev-xx/go-ninja-chat/internal/logger"
	swag "github.com/getkin/kin-openapi/openapi3"
	serverdebug "github.com/dndev-xx/go-ninja-chat/internal/server-debug"
	serverclient "github.com/dndev-xx/go-ninja-chat/internal/server-client"
	h "github.com/dndev-xx/go-ninja-chat/internal/server-client/v1"
	"go.uber.org/zap"
)

var configPath = flag.String("config", "configs/config.toml", "Path to config file")

type AppContext struct {
	Config      	*config.Config
	Logger      	*zap.Logger
	DebugServer 	*serverdebug.Server
	Swagger 		*swag.T
	ClientServer 	*serverclient.Server
}

type AppBuilder struct {
	App *AppContext
	err error
}

func NewAppBuilder() *AppBuilder {
	return &AppBuilder{
		App: &AppContext{},
	}
}

func (b *AppBuilder) WithConfig() Builder {
	flag.Parse()
	cfg, err := config.ParseAndValidate(*configPath)
	if err != nil {
		b.err = fmt.Errorf("parse and validate config %q: %v", *configPath, err)
		return b
	}
	b.App.Config = cfg
	return b
}

func (b *AppBuilder) WithLogger() Builder {
	if err := logger.Init(logger.NewOptions(
		b.App.Config.Log.Level,
		logger.WithProductionMode(b.App.Config.Global.IsProduction()),
	)); err != nil {
		b.err = err
		return b
	}
	b.App.Logger = logger.GetLogger()
	b.App.Logger.Info("configuration logger was successful with level", zap.String("level", b.App.Config.Log.Level))
	return b
}

func (b *AppBuilder) WithDebugHTTPSrv() Builder {
	srvDebug, err := serverdebug.New(serverdebug.NewOptions(b.App.Config.Servers.Debug.Addr))
	if err != nil {
		b.err = fmt.Errorf("init debug server: %v", err)
		return b
	}
	b.App.DebugServer = srvDebug
	b.App.Logger.Info("configuration debug server was successful")
	return b
}

func (b *AppBuilder) WithSwagger() Builder {
	swagger, err := swag.NewLoader().LoadFromFile("api/client.v1.swagger.yaml")
	if err != nil {
		b.err = fmt.Errorf("load swagger spec: %v", err)
		return b
	}
	b.App.Swagger = swagger
	return b
}

func (b *AppBuilder) WithClientHTTPSrv() Builder {
	handlers, err := h.NewHandlers(h.Options{})
	if err != nil {
		b.err = fmt.Errorf("create v1 handlers %v", err)
		return b
	}
	server, err := serverclient.New(serverclient.NewOptions(
		b.App.Logger,
		b.App.Config.Servers.Client.Addr,
		b.App.Config.Servers.Client.AllowOrigins,
		b.App.Swagger,
		handlers,
	))
	if err != nil {
		b.err = fmt.Errorf("create server %v", err)
		return b
	}
	b.App.ClientServer = server
	return b
}

func (b *AppBuilder) GetContext() (*AppContext, error) {
	return b.App, b.err
}