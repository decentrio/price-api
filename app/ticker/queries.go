package ticker

import (
	"context"

	app "github.com/decentrio/price-api/app"
	types "github.com/decentrio/price-api/types/ticker"
	tradetypes "github.com/decentrio/price-api/types/trade"
)

var _ types.TickerQueryServer = Keeper{}

const (
	UsdcTokenName = "USDC-GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN"
)

func (k Keeper) Tickers(ctx context.Context, request *types.TickersRequest) (*types.TickersResponse, error) {
	var tickers []*types.Ticker

	err := k.dbHandler.Table(app.TICKER_TABLE).
		Find(&tickers).Error

	if err != nil {
		return &types.TickersResponse{}, err
	}

	var tickerInfos []*types.TickerInfo

	for _, ticker := range tickers {
		tickerInfos = append(tickerInfos, convertToInfo(ticker))
	}

	return &types.TickersResponse{
		Tickers: tickerInfos,
	}, nil
}

func convertToInfo(ticker *types.Ticker) *types.TickerInfo {
	baseVol := float64(ticker.BaseVolume) / 10000000.0
	targetVol := float64(ticker.TargetVolume) / 10000000.0
	liquidUsd := float64(ticker.LiquidityInUsd) / 10000000.0

	return &types.TickerInfo{
		TickerId:       ticker.TickerId,
		BaseCurrency:   ticker.BaseCurrency,
		TargetCurrency: ticker.TargetCurrency,
		PoolId:         ticker.PoolId,
		LastPrice:      ticker.LastPrice,
		BaseVolume:     baseVol,
		TargetVolume:   targetVol,
		High:           ticker.High,
		Low:            ticker.Low,
		LiquidityInUsd: liquidUsd,
	}
}

func (k Keeper) Price(ctx context.Context, request *types.TokenPriceRequest) (*types.TokenPriceResponse, error) {
	var token tradetypes.Token

	err := k.dbHandler.Table(app.TOKEN_TABLE).Where("symbol = ?", request.Name).
		First(&token).Error
	if err != nil {
		return &types.TokenPriceResponse{}, err
	}

	return &types.TokenPriceResponse{
		Price: token.PriceInUsd,
	}, nil
}

func (k Keeper) PoolShare(ctx context.Context, request *types.PoolShareRequest) (*types.PoolShareResponse, error) {
	var pool tradetypes.Pool

	err := k.dbHandler.Table(app.POOL_TABLE).Where("pool_address = ?", request.ContractId).
		Order("tx_time DESC").
		First(&pool).Error
	if err != nil {
		return &types.PoolShareResponse{}, err
	}

	return &types.PoolShareResponse{
		Amount: float64(pool.Share) / 10000000.0,
	}, nil
}

func (k Keeper) PoolReserveA(ctx context.Context, request *types.PoolReserveARequest) (*types.PoolReserveAResponse, error) {
	var pool tradetypes.Pool
	err := k.dbHandler.Table(app.POOL_TABLE).Where("pool_address = ?", request.ContractId).
		Order("tx_time DESC").
		First(&pool).Error
	if err != nil {
		return &types.PoolReserveAResponse{}, err
	}
	return &types.PoolReserveAResponse{
		Amount: float64(pool.ReserveA) / 10000000.0,
	}, nil
}

func (k Keeper) PoolReserveB(ctx context.Context, request *types.PoolReserveBRequest) (*types.PoolReserveBResponse, error) {
	var pool tradetypes.Pool
	err := k.dbHandler.Table(app.POOL_TABLE).Where("pool_address = ?", request.ContractId).
		Order("tx_time DESC").
		First(&pool).Error
	if err != nil {
		return &types.PoolReserveBResponse{}, err
	}
	return &types.PoolReserveBResponse{
		Amount: float64(pool.ReserveB) / 10000000.0,
	}, nil
}

func (k Keeper) PoolTotalLiquidityInUsd(ctx context.Context, request *types.PoolTotalLiquidityInUsdRequest) (*types.PoolTotalLiquidityInUsdResponse, error) {
	var pool tradetypes.Pool
	err := k.dbHandler.Table(app.POOL_TABLE).Where("pool_address = ?", request.ContractId).
		Order("tx_time DESC").
		First(&pool).Error
	if err != nil {
		return &types.PoolTotalLiquidityInUsdResponse{}, err
	}
	var ticker types.Ticker
	err = k.dbHandler.Table(app.TICKER_TABLE).Where("pool_id = ?", request.ContractId).Scan(&ticker).Error
	if err != nil {
		return &types.PoolTotalLiquidityInUsdResponse{}, err
	}
	return &types.PoolTotalLiquidityInUsdResponse{
		Amount: float64(pool.ReserveA) * 2 / 10000000.0 * k.getTokenPriceInUsd(ticker.BaseCurrency),
	}, nil
}

func (k Keeper) getTokenPriceInUsd(tokenName string) float64 {
	if tokenName == UsdcTokenName {
		return float64(1)
	} else {
		var token tradetypes.Token
		if err := k.dbHandler.Table(app.TOKEN_TABLE).Where("token_name = ?", tokenName).Scan(&token).Error; err != nil {
			return 0
		}
		return token.PriceInUsd
	}
}
