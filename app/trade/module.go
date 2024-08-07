package trade

import (
	"time"

	types "github.com/decentrio/price-api/types/trade"
	"google.golang.org/grpc"
)

const (
	Day   = time.Hour * 24
	Week  = Day * 7
	Month = Day * 30
	Year  = Day * 365

	DaySec   = 86400
	HourSec  = 3600
	MinSec   = 60
	WeekSec  = DaySec * 7
	MonthSec = DaySec * 30
	YearSec  = DaySec * 365
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
	types.RegisterTradeQueryServer(server, am.keeper)
}
