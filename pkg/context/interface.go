package context

import (
	"context"
)

type Builder interface {
	WithContext(ctx context.Context) Builder
	WithConfig() Builder
	WithLogger() Builder
	WithDebugHTTPSrv() Builder
	WithSwagger() Builder
	WithClientHTTPSrv() Builder
	WithStoresDB() Builder
	GetContext() (*AppContext, error)
}
