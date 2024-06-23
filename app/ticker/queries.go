package ticker

import (
	"context"
	"strconv"

	app "github.com/decentrio/price-api/app"
	types "github.com/decentrio/price-api/types/ticker"
)

const (
	SymbolTickerId = "PHO_USDC"

	TickerId          = "PHO-GAX5TXB5RYJNLBUR477PEXM4X75APK2PGMTN6KEFQSESGWFXEAKFSXJO_USDC-GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN"
	PHOTokenContract  = "PHO-GAX5TXB5RYJNLBUR477PEXM4X75APK2PGMTN6KEFQSESGWFXEAKFSXJO"
	USDCTokenContract = "USDC-GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN"
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

	// it should be only 1 ticker for now so we hard code here
	tickerInfos[0].TickerId = TickerId
	tickerInfos[0].BaseCurrency = PHOTokenContract
	tickerInfos[0].TargetCurrency = USDCTokenContract

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
		BaseVolume:     strconv.FormatFloat(baseVol, 'f', 7, 64),
		TargetVolume:   strconv.FormatFloat(targetVol, 'f', 7, 64),
		LiquidityInUsd: strconv.FormatFloat(liquidUsd, 'f', 7, 64),
	}
}
