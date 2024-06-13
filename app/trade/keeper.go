package trade

import (
	"gorm.io/gorm"

	types "github.com/decentrio/price-api/types/trade"
)

type Keeper struct {
	dbHandler *gorm.DB
	types.UnimplementedTradeQueryServer
}

func NewKeeper(db *gorm.DB) *Keeper {
	return &Keeper{
		dbHandler: db,
	}
}

var _ types.TradeQueryServer = Keeper{}
