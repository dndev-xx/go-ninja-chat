package context

type Builder interface {
	WithConfig() Builder
	WithLogger() Builder
	WithDebugHTTPSrv() Builder
	WithSwagger() Builder
	WithClientHTTPSrv() Builder
	GetContext() (*AppContext, error)
}