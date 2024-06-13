package ticker

import (
	"gorm.io/gorm"

	types "github.com/decentrio/price-api/types/ticker"
)

type Keeper struct {
	dbHandler *gorm.DB
	types.UnimplementedTickerQueryServer
}

func NewKeeper(db *gorm.DB) *Keeper {
	return &Keeper{
		dbHandler: db,
	}
}

var _ types.TickerQueryServer = Keeper{}

