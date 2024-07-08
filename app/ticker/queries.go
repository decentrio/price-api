package ticker

import (
	"context"

	app "github.com/decentrio/price-api/app"
	types "github.com/decentrio/price-api/types/ticker"
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

// Liquidity
func (k Keeper) PoolTotalLiquidityInUsd(ctx context.Context, request *types.PoolTotalLiquidityInUsdRequest) (*types.PoolTotalLiquidityInUsdResponse, error) {
	var ticker types.Ticker

	// get liquidity
	k.dbHandler.Table("tickers").Where("pool_id = ?", request.ContractId).Scan(&ticker)

	return &types.PoolTotalLiquidityInUsdResponse{
		Amount: float64(ticker.LiquidityInUsd) / 10000000.0,
	}, nil
}
