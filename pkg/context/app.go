package context

import (
	"context"
	"flag"
	"fmt"

	keycloakclient "github.com/dndev-xx/go-ninja-chat/internal/clients/keycloak"
	"github.com/dndev-xx/go-ninja-chat/internal/config"
	"github.com/dndev-xx/go-ninja-chat/internal/logger"

	repoChats "github.com/dndev-xx/go-ninja-chat/internal/repositories/chats"
	repo "github.com/dndev-xx/go-ninja-chat/internal/repositories/messages"
	repoProblems "github.com/dndev-xx/go-ninja-chat/internal/repositories/problems"
	serverclient "github.com/dndev-xx/go-ninja-chat/internal/server-client"
	h "github.com/dndev-xx/go-ninja-chat/internal/server-client/v1"
	sw "github.com/dndev-xx/go-ninja-chat/internal/server-client/v1/pkg"
	serverdebug "github.com/dndev-xx/go-ninja-chat/internal/server-debug"
	"github.com/dndev-xx/go-ninja-chat/internal/store"
	db "github.com/dndev-xx/go-ninja-chat/internal/store"

	usecase "github.com/dndev-xx/go-ninja-chat/internal/usecase/client/get-history"
	usecaseMsg "github.com/dndev-xx/go-ninja-chat/internal/usecase/client/send-message"
	swag "github.com/getkin/kin-openapi/openapi3"
	"go.uber.org/zap"
)

var configPath = flag.String("config", "configs/config.toml", "Path to config file")

type AppContext struct {
	context			context.Context
	Config      	*config.Config
	Logger      	*zap.Logger
	DebugServer 	*serverdebug.Server
	Swagger 		*swag.T
	ClientServer 	*serverclient.Server
	Stores			*store.Client
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

func (b *AppBuilder) WithContext(ctx context.Context) Builder {
	b.App.context = ctx
	return b
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
	swagger, err := sw.GetSwagger()
	if err != nil {
		b.err = fmt.Errorf("load swagger spec: %v", err)
		return b
	}
	b.App.Swagger = swagger
	return b
}

func (b *AppBuilder) WithStoresDB() Builder {
	client, err := db.NewPSQLClient(db.NewPSQLOptions(
		b.App.context,
		b.App.Config.Stores.PSQL.Addr,
		b.App.Config.Stores.PSQL.Username,
		b.App.Config.Stores.PSQL.Password,
		b.App.Config.Stores.PSQL.Database,
		b.App.Config.Stores.PSQL.Debug,
	))
	if err != nil {
		b.err = fmt.Errorf("failed connect to db: %v", err)
		return b
	}
	b.App.Stores = client
	return b
}

func (b *AppBuilder) WithClientHTTPSrv() Builder {
	db := store.NewDatabase(b.App.Stores)
	msgRepo, err := repo.New(repo.NewOptions(
    db,
	))
	chatRepo, err := repoChats.New(repoChats.NewOptions(db))
	repoProblems, err := repoProblems.New(repoProblems.NewOptions(db))
	if err != nil {
		b.err = fmt.Errorf("create v1 repository %v", err)
		return b
	}
	usecaseHist, err := usecase.New(usecase.NewOptions(msgRepo))
	usecaseMsg, err := usecaseMsg.New(usecaseMsg.NewOptions(msgRepo, chatRepo, repoProblems))
	if err != nil {
		b.err = fmt.Errorf("create v1 usecase %v", err)
		return b
	}
	handlers, err := h.NewHandlers(h.NewOptions(usecaseHist, usecaseMsg))
	if err != nil {
		b.err = fmt.Errorf("create v1 handlers %v", err)
		return b
	}
	kc, err := keycloakclient.New(keycloakclient.NewOptions(
		keycloakclient.WithBasePath(b.App.Config.Clients.Keycloak.BasePath),
		keycloakclient.WithRealm(b.App.Config.Clients.Keycloak.Realm),
		keycloakclient.WithClientID(b.App.Config.Clients.Keycloak.ClientID),
		keycloakclient.WithClientSecret(b.App.Config.Clients.Keycloak.ClientSecret),
		keycloakclient.WithDebugMode(b.App.Config.Clients.Keycloak.DebugMode),
	))
	if err != nil {
		b.err = fmt.Errorf("create keycloak error %v", err)
		return b
	}
	server, err := serverclient.New(serverclient.NewOptions(
		b.App.Logger,
		b.App.Config.Servers.Client.Addr,
		b.App.Config.Servers.Client.AllowOrigins,
		b.App.Swagger,
		handlers,
		kc,
		b.App.Config.Servers.Client.RequiredAccess.Resource,
		b.App.Config.Servers.Client.RequiredAccess.Role,
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
