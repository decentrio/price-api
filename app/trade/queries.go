package trade

import (
	"context"
	"strings"
	"time"

	app "github.com/decentrio/price-api/app"
	tickertypes "github.com/decentrio/price-api/types/ticker"
	types "github.com/decentrio/price-api/types/trade"
	"golang.org/x/exp/maps"
)

var _ types.TradeQueryServer = Keeper{}

func (k Keeper) Trades(ctx context.Context, request *types.TradesRequest) (*types.TradesResponse, error) {
	var trades []*types.Trade

	query := k.dbHandler.Table(app.TRADE_TABLE).Order("trade_timestamp DESC")

	if request.TickerId != "" {
		query = query.Where("ticker_id = ?", request.TickerId)
	}
	if request.Type != "" {
		query = query.Where("trade_type = ?", strings.ToLower(request.Type))
	}
	if request.StartTime != 0 {
		query = query.Where("trade_timestamp >= ?", request.StartTime)
	}
	if request.EndTime != 0 {
		query = query.Where("trade_timestamp <= ?", request.EndTime)
	}
	if request.Limit != 0 {
		query = query.Limit(int(request.Limit))
	}

	err := query.Find(&trades).Error
	if err != nil {
		return &types.TradesResponse{}, nil
	}

	var tradeInfos []*types.TradeInfo

	for _, trade := range trades {
		tradeInfos = append(tradeInfos, convertToInfo(trade))
	}

	return &types.TradesResponse{
		Trades: tradeInfos,
	}, nil
}

func convertToInfo(trade *types.Trade) *types.TradeInfo {
	baseVol := float64(trade.BaseVolume) / 10000000.0
	targetVol := float64(trade.TargetVolume) / 10000000.0

	return &types.TradeInfo{
		TradeId:        trade.TradeId,
		Price:          trade.Price,
		BaseVolume:     baseVol,
		TargetVolume:   targetVol,
		TradeTimestamp: trade.TradeTimestamp * 1000,
		Type:           trade.TradeType,
		TickerId:       trade.TickerId,
	}
}

// historical trading volume
func (k Keeper) TradingVolumePerWeek(ctx context.Context, request *types.TradingVolumePerWeekRequest) (*types.TradingVolumePerWeekResponse, error) {
	// get ticker_id by contract_id
	var ticker tickertypes.Ticker
	k.dbHandler.Table(app.TICKER_TABLE).Where("pool_id = ?", request.ContractId).Scan(&ticker)

	var trades []*types.Trade
	query := k.dbHandler.Table(app.TRADE_TABLE).Order("trade_timestamp DESC").
		Where("ticker_id = ?", ticker.TickerId).
		Where("trade_timestamp >= ?", request.From).
		Where("trade_timestamp <= ?", request.To)

	err := query.Find(&trades).Error
	if err != nil {
		return &types.TradingVolumePerWeekResponse{}, nil
	}

	// calculate volume in week
	tradingVolumes := make(map[*types.Week]*types.TradeVolumeByWeek)

	for _, trade := range trades {
		year, week := time.Unix(int64(trade.TradeTimestamp), 0).ISOWeek()

		tradingWeek := &types.Week{
			Year: uint32(year),
			Week: uint32(week),
		}

		tradingVolume, found := tradingVolumes[tradingWeek]
		if found {
			tradingVolume.TokenAVolume += trade.BaseVolume
			tradingVolume.TokenBVolume += trade.TargetVolume
			tradingVolumes[tradingWeek] = tradingVolume
		} else {
			tradingVolume := &types.TradeVolumeByWeek{
				Week:         tradingWeek,
				TokenAVolume: trade.BaseVolume,
				TokenBVolume: trade.TargetVolume,
			}

			tradingVolumes[tradingWeek] = tradingVolume
		}

	}

	vals := maps.Values(tradingVolumes)

	return &types.TradingVolumePerWeekResponse{
		TradingVolume: vals,
	}, nil
}

func (k Keeper) TradingVolumePerMonth(ctx context.Context, request *types.TradingVolumePerMonthRequest) (*types.TradingVolumePerMonthResponse, error) {
	// get ticker_id by contract_id
	var ticker tickertypes.Ticker
	k.dbHandler.Table(app.TICKER_TABLE).Where("pool_id = ?", request.ContractId).Scan(&ticker)

	var trades []*types.Trade
	query := k.dbHandler.Table(app.TRADE_TABLE).Order("trade_timestamp DESC").
		Where("ticker_id = ?", ticker.TickerId).
		Where("trade_timestamp >= ?", request.From).
		Where("trade_timestamp <= ?", request.To)

	err := query.Find(&trades).Error
	if err != nil {
		return &types.TradingVolumePerMonthResponse{}, nil
	}

	// calculate volume in month
	tradingVolumes := make(map[*types.Month]*types.TradeVolumeByMonth)

	for _, trade := range trades {
		year := time.Unix(int64(trade.TradeTimestamp), 0).Year()
		month := time.Unix(int64(trade.TradeTimestamp), 0).Month()

		tradingMonth := &types.Month{
			Year:  uint32(year),
			Month: uint32(month),
		}

		tradingVolume, found := tradingVolumes[tradingMonth]
		if found {
			tradingVolume.TokenAVolume += trade.BaseVolume
			tradingVolume.TokenBVolume += trade.TargetVolume
			tradingVolumes[tradingMonth] = tradingVolume
		} else {
			tradingVolume := &types.TradeVolumeByMonth{
				Month:        tradingMonth,
				TokenAVolume: trade.BaseVolume,
				TokenBVolume: trade.TargetVolume,
			}

			tradingVolumes[tradingMonth] = tradingVolume
		}

	}

	vals := maps.Values(tradingVolumes)

	return &types.TradingVolumePerMonthResponse{
		TradingVolume: vals,
	}, nil
}

func (k Keeper) TradingVolumePerDay(ctx context.Context, request *types.TradingVolumePerDayRequest) (*types.TradingVolumePerDayResponse, error) {
	// get ticker_id by contract_id
	var ticker tickertypes.Ticker
	k.dbHandler.Table(app.TICKER_TABLE).Where("pool_id = ?", request.ContractId).Scan(&ticker)

	var trades []*types.Trade
	query := k.dbHandler.Table(app.TRADE_TABLE).Order("trade_timestamp DESC").
		Where("ticker_id = ?", ticker.TickerId).
		Where("trade_timestamp >= ?", request.From).
		Where("trade_timestamp <= ?", request.To)

	err := query.Find(&trades).Error
	if err != nil {
		return &types.TradingVolumePerDayResponse{}, nil
	}

	// calculate volume in month
	tradingVolumes := make(map[*types.Date]*types.TradeVolumeByDate)

	for _, trade := range trades {
		year, month, day := time.Unix(int64(trade.TradeTimestamp), 0).Date()

		tradingDay := &types.Date{
			Year:  uint32(year),
			Month: uint32(month),
			Day:   uint32(day),
		}

		tradingVolume, found := tradingVolumes[tradingDay]
		if found {
			tradingVolume.TokenAVolume += trade.BaseVolume
			tradingVolume.TokenBVolume += trade.TargetVolume
			tradingVolumes[tradingDay] = tradingVolume
		} else {
			tradingVolume := &types.TradeVolumeByDate{
				Date:         tradingDay,
				TokenAVolume: trade.BaseVolume,
				TokenBVolume: trade.TargetVolume,
			}

			tradingVolumes[tradingDay] = tradingVolume
		}

	}

	vals := maps.Values(tradingVolumes)

	return &types.TradingVolumePerDayResponse{
		TradingVolume: vals,
	}, nil
}
