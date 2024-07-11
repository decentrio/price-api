package user

import (
	types "github.com/decentrio/price-api/types/user"
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
	types.RegisterUserQueryServer(server, am.keeper)
}
