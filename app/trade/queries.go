package trade

import (
	"context"
	"sort"
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

func (k Keeper) TradingVolumePerHour(ctx context.Context, request *types.TradingVolumePerHourRequest) (*types.TradingVolumePerHourResponse, error) {
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
		return &types.TradingVolumePerHourResponse{}, nil
	}

	// calculate volume in hour
	tradingVolumes := make(map[*types.TimeHour]*types.TradeVolumeByHour)

	for _, trade := range trades {
		year, month, day := time.Unix(int64(trade.TradeTimestamp), 0).Date()
		hour, _, _ := time.Unix(int64(trade.TradeTimestamp), 0).Clock()

		tradingHour := &types.TimeHour{
			Hour: uint32(hour),
			Date: &types.Date{
				Year:  uint32(year),
				Month: uint32(month),
				Day:   uint32(day),
			},
		}

		tradingVolume, found := tradingVolumes[tradingHour]
		if found {
			tradingVolume.TokenAVolume += trade.BaseVolume
			tradingVolume.TokenBVolume += trade.TargetVolume
			tradingVolumes[tradingHour] = tradingVolume
		} else {
			tradingVolume := &types.TradeVolumeByHour{
				Time:         tradingHour,
				TokenAVolume: trade.BaseVolume,
				TokenBVolume: trade.TargetVolume,
			}

			tradingVolumes[tradingHour] = tradingVolume
		}

	}

	vals := maps.Values(tradingVolumes)

	return &types.TradingVolumePerHourResponse{
		TradingVolume: vals,
	}, nil
}

func (k Keeper) PriceGraph(ctx context.Context, request *types.PriceGraphRequest) (*types.PriceGraphResponse, error) {
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
		return &types.PriceGraphResponse{}, nil
	}

	var priceGraphs []*types.PriceGraph
	for _, trade := range trades {
		pg := &types.PriceGraph{
			TimeStamp: trade.TradeTimestamp,
			Price:     trade.Price,
		}
		priceGraphs = append(priceGraphs, pg)
	}

	sort.Slice(priceGraphs, func(i, j int) bool {
		return priceGraphs[i].TimeStamp < priceGraphs[j].TimeStamp
	})

	return &types.PriceGraphResponse{
		Graph: priceGraphs,
	}, nil
}

func (k Keeper) PriceGraphLastWeek(ctx context.Context, request *types.PriceGraphLastWeekRequest) (*types.PriceGraphLastWeekResponse, error) {
	to := time.Now()
	from := to.Add(-Week)

	// get ticker_id by contract_id
	var ticker tickertypes.Ticker
	k.dbHandler.Table(app.TICKER_TABLE).Where("pool_id = ?", request.ContractId).Scan(&ticker)

	var trades []*types.Trade
	query := k.dbHandler.Table(app.TRADE_TABLE).Order("trade_timestamp DESC").
		Where("ticker_id = ?", ticker.TickerId).
		Where("trade_timestamp >= ?", from).
		Where("trade_timestamp <= ?", to)

	err := query.Find(&trades).Error
	if err != nil {
		return &types.PriceGraphLastWeekResponse{}, nil
	}

	var priceGraphs []*types.PriceGraph
	for _, trade := range trades {
		pg := &types.PriceGraph{
			TimeStamp: trade.TradeTimestamp,
			Price:     trade.Price,
		}
		priceGraphs = append(priceGraphs, pg)
	}

	sort.Slice(priceGraphs, func(i, j int) bool {
		return priceGraphs[i].TimeStamp < priceGraphs[j].TimeStamp
	})

	return &types.PriceGraphLastWeekResponse{
		Graph: priceGraphs,
	}, nil
}

func (k Keeper) PriceGraphLastMonth(ctx context.Context, request *types.PriceGraphLastMonthRequest) (*types.PriceGraphLastMonthResponse, error) {
	to := time.Now()
	from := to.Add(-Month)

	// get ticker_id by contract_id
	var ticker tickertypes.Ticker
	k.dbHandler.Table(app.TICKER_TABLE).Where("pool_id = ?", request.ContractId).Scan(&ticker)

	var trades []*types.Trade
	query := k.dbHandler.Table(app.TRADE_TABLE).Order("trade_timestamp DESC").
		Where("ticker_id = ?", ticker.TickerId).
		Where("trade_timestamp >= ?", from).
		Where("trade_timestamp <= ?", to)

	err := query.Find(&trades).Error
	if err != nil {
		return &types.PriceGraphLastMonthResponse{}, nil
	}

	var priceGraphs []*types.PriceGraph
	for _, trade := range trades {
		pg := &types.PriceGraph{
			TimeStamp: trade.TradeTimestamp,
			Price:     trade.Price,
		}
		priceGraphs = append(priceGraphs, pg)
	}

	sort.Slice(priceGraphs, func(i, j int) bool {
		return priceGraphs[i].TimeStamp < priceGraphs[j].TimeStamp
	})

	return &types.PriceGraphLastMonthResponse{
		Graph: priceGraphs,
	}, nil
}

func (k Keeper) PriceGraphLastYear(ctx context.Context, request *types.PriceGraphLastYearRequest) (*types.PriceGraphLastYearResponse, error) {
	to := time.Now()
	from := to.Add(-Year)

	// get ticker_id by contract_id
	var ticker tickertypes.Ticker
	k.dbHandler.Table(app.TICKER_TABLE).Where("pool_id = ?", request.ContractId).Scan(&ticker)

	var trades []*types.Trade
	query := k.dbHandler.Table(app.TRADE_TABLE).Order("trade_timestamp DESC").
		Where("ticker_id = ?", ticker.TickerId).
		Where("trade_timestamp >= ?", from).
		Where("trade_timestamp <= ?", to)

	err := query.Find(&trades).Error
	if err != nil {
		return &types.PriceGraphLastYearResponse{}, nil
	}

	var priceGraphs []*types.PriceGraph
	for _, trade := range trades {
		pg := &types.PriceGraph{
			TimeStamp: trade.TradeTimestamp,
			Price:     trade.Price,
		}
		priceGraphs = append(priceGraphs, pg)
	}

	sort.Slice(priceGraphs, func(i, j int) bool {
		return priceGraphs[i].TimeStamp < priceGraphs[j].TimeStamp
	})

	return &types.PriceGraphLastYearResponse{
		Graph: priceGraphs,
	}, nil
}

func (k Keeper) TradeHistoricals(ctx context.Context, request *types.TradeHistoricalRequest) (*types.TradeHistoricalResponse, error) {
	var trades []*types.Trade

	query := k.dbHandler.Table(app.TRADE_TABLE).Order("trade_timestamp DESC")

	if request.Address != "" {
		query = query.Where("maker = ?", request.Address)
	}
	if request.From != 0 {
		query = query.Where("trade_timestamp >= ?", request.From)
	}
	if request.To != 0 {
		query = query.Where("trade_timestamp <= ?", request.To)
	}

	page := int(request.Page)
	if request.Page < 1 {
		page = 1
	}
	pageSize := int(request.PageSize)
	if request.PageSize < 1 {
		pageSize = app.PAGE_SIZE
	}

	offset := (page - 1) * pageSize
	query = query.Limit(int(pageSize)).Offset(offset)

	err := query.Find(&trades).Error
	if err != nil {
		return &types.TradeHistoricalResponse{}, nil
	}

	var tradeInfos []*types.TradeInfo

	for _, trade := range trades {
		tradeInfos = append(tradeInfos, convertToInfo(trade))
	}

	return &types.TradeHistoricalResponse{
		Trades: tradeInfos,
	}, nil
}
