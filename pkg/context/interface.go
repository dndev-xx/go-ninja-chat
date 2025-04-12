package context

type Builder interface {
	WithConfig() Builder
	WithLogger() Builder
	WithDebugHTTPSrv() Builder
	GetContext() (*AppContext, error)
}