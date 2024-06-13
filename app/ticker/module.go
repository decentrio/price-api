package ticker

import (
	types "github.com/decentrio/price-api/types/ticker"
	"google.golang.org/grpc"
)

type AppModule struct {
	keeper Keeper
}

func NewAppModule(
	keeper Keeper,
) AppModule {
	return AppModule{
		keeper: keeper,
	}
}

func (am AppModule) RegisterServices(server *grpc.Server) {
	types.RegisterTickerQueryServer(server, am.keeper)
}
