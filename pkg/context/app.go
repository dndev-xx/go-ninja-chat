package context

import (
	"context"
	"flag"
	"fmt"

	swag "github.com/getkin/kin-openapi/openapi3"
	"go.uber.org/zap"

	keycloakclient "github.com/dndev-xx/go-ninja-chat/internal/clients/keycloak"
	"github.com/dndev-xx/go-ninja-chat/internal/config"
	"github.com/dndev-xx/go-ninja-chat/internal/logger"
	repoChats "github.com/dndev-xx/go-ninja-chat/internal/repositories/chats"
	repoJobs "github.com/dndev-xx/go-ninja-chat/internal/repositories/jobs"
	repo "github.com/dndev-xx/go-ninja-chat/internal/repositories/messages"
	repoProblems "github.com/dndev-xx/go-ninja-chat/internal/repositories/problems"
	serverclient "github.com/dndev-xx/go-ninja-chat/internal/server-client"
	servererror "github.com/dndev-xx/go-ninja-chat/internal/server-client/errhandler"
	h "github.com/dndev-xx/go-ninja-chat/internal/server-client/v1"
	sw "github.com/dndev-xx/go-ninja-chat/internal/server-client/v1/pkg"
	serverdebug "github.com/dndev-xx/go-ninja-chat/internal/server-debug"
	msgProducer "github.com/dndev-xx/go-ninja-chat/internal/services/msg-producer"
	obox "github.com/dndev-xx/go-ninja-chat/internal/services/outbox"
	regMsgProd "github.com/dndev-xx/go-ninja-chat/internal/services/outbox/jobs/send-client-message"
	"github.com/dndev-xx/go-ninja-chat/internal/store"
	db "github.com/dndev-xx/go-ninja-chat/internal/store"
	usecase "github.com/dndev-xx/go-ninja-chat/internal/usecase/client/get-history"
	usecaseMsg "github.com/dndev-xx/go-ninja-chat/internal/usecase/client/send-message"
)

var configPath = flag.String("config", "configs/config.toml", "Path to config file")

type AppContext struct {
	context      context.Context
	Config       *config.Config
	Logger       *zap.Logger
	DebugServer  *serverdebug.Server
	Swagger      *swag.T
	ClientServer *serverclient.Server
	Stores       *store.Client
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
	if b.err != nil {
		return b
	}
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
	if b.err != nil {
		return b
	}
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
	if b.err != nil {
		return b
	}
	swagger, err := sw.GetSwagger()
	if err != nil {
		b.err = fmt.Errorf("load swagger spec: %v", err)
		return b
	}
	b.App.Swagger = swagger
	return b
}

func (b *AppBuilder) WithStoresDB() Builder {
	if b.err != nil {
		return b
	}
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
	if b.err != nil {
		return b
	}
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
	jobRepo, err := repoJobs.New(repoJobs.NewOptions(db))
	if err != nil {
		b.err = fmt.Errorf("create v1 repository %v", err)
		return b
	}
	msgProd, err := msgProducer.New(msgProducer.NewOptions(
		msgProducer.NewKafkaWriter(
			b.App.Config.Services.MsgProducer.Brokers,
			b.App.Config.Services.MsgProducer.Topic,
			16,
		),
		msgProducer.WithEncryptKey(b.App.Config.Services.MsgProducer.EncryptKey),
	))
	if err != nil {
		b.err = fmt.Errorf("create v1 usecase %v", err)
		return b
	}
	job, err := regMsgProd.New(regMsgProd.NewOptions(msgProd, msgRepo))
	if err != nil {
		b.err = fmt.Errorf("create v1 usecase %v", err)
		return b
	}
	outbox := obox.New(jobRepo, db, obox.Config{
		Workers:    10,
		IdleTime:   b.App.Config.Services.Outbox.IdleTime,
		ReserveFor: b.App.Config.Services.Outbox.ReserveFor,
		Logger:     b.App.Logger,
	})
	outbox.MustRegisterJob(job)
	go outbox.Start(b.App.context)
	usecaseHist, err := usecase.New(usecase.NewOptions(msgRepo))
	usecaseMsg, err := usecaseMsg.New(usecaseMsg.NewOptions(msgRepo, chatRepo, repoProblems, db, outbox))
	if err != nil {
		b.err = fmt.Errorf("create v1 usecase %v", err)
		return b
	}
	httpErrorHandler, err := servererror.New(servererror.NewOptions(
		b.App.Logger,
		b.App.Config.Global.IsProduction(),
		servererror.ResponseBuilder,
	))
	if err != nil {
		b.err = fmt.Errorf("create http error handler: %v", err)
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
		httpErrorHandler.Handle,
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
