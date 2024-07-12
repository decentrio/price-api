package user

import (
	"gorm.io/gorm"

	types "github.com/decentrio/price-api/types/user"
)

type Keeper struct {
	dbHandler *gorm.DB
	types.UnimplementedUserQueryServer
}

func NewKeeper(db *gorm.DB) *Keeper {
	return &Keeper{
		dbHandler: db,
	}
}
