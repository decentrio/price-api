package trade

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	app "github.com/decentrio/price-api/app"
	tickertypes "github.com/decentrio/price-api/types/ticker"
	types "github.com/decentrio/price-api/types/trade"
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
		return &types.TradesResponse{}, err
	}

	var tradeInfos []*types.TradeInfo

	for _, trade := range trades {
		tradeInfos = append(tradeInfos, convertToInfo(trade))
	}

	return &types.TradesResponse{
		Trades: tradeInfos,
	}, nil
}

func (k Keeper) AdvancedTrades(ctx context.Context, request *types.AdvancedTradesRequest) (*types.AdvancedTradesResponse, error) {
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
	if request.Address != "" {
		query = query.Where("maker = ?", request.Address)
	}
	if request.PoolId != "" {
		var ticker tickertypes.Ticker
		err := k.dbHandler.Table(app.TICKER_TABLE).
			Where("pool_id = ?", request.PoolId).
			First(&ticker).Error
		if err != nil {
			return &types.AdvancedTradesResponse{}, err
		}
		query = query.Where("ticker_id = ?", ticker.TickerId)
	}

	err := query.Order("trade_timestamp DESC").Find(&trades).Error
	if err != nil {
		return &types.AdvancedTradesResponse{}, err
	}

	var tradeInfos []*types.TradeInfo

	for _, trade := range trades {
		tradeInfos = append(tradeInfos, convertToInfo(trade))
	}

	return &types.AdvancedTradesResponse{
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

func (k Keeper) Shares(ctx context.Context, request *types.SharesRequest) (*types.SharesResponse, error) {
	var data types.Share
	err := k.dbHandler.Table(app.SHARE_TABLE).
		Where("pool_address = ?", request.PoolId).
		Where("account_address = ?", request.Address).
		Order("tx_time DESC").
		First(&data).Error
	if err != nil {
		return &types.SharesResponse{}, err
	}

	return &types.SharesResponse{
		Shares: data.Amount,
	}, nil
}

func (k Keeper) LockShares(ctx context.Context, request *types.LockSharesRequest) (*types.LockSharesResponse, error) {
	var data types.Stake
	err := k.dbHandler.Table(app.STAKE_TABLE).
		Where("pool_address = ?", request.PoolId).
		Where("account_address = ?", request.Address).
		Order("tx_time DESC").
		First(&data).Error
	if err != nil {
		return &types.LockSharesResponse{}, err
	}

	return &types.LockSharesResponse{
		Shares: data.Amount,
	}, nil
}

// historical trading volume
func (k Keeper) TradingVolumePerWeek(ctx context.Context, request *types.TradingVolumePerWeekRequest) (*types.TradingVolumePerWeekResponse, error) {
	if request.ContractId == "" {
		return &types.TradingVolumePerWeekResponse{}, fmt.Errorf("error empty contract id")
	}
	// get ticker_id by contract_id
	var ticker tickertypes.Ticker
	k.dbHandler.Table(app.TICKER_TABLE).Where("pool_id = ?", request.ContractId).Scan(&ticker)

	from := request.From
	to := request.To

	curTime := time.Now().Unix()

	if from == 0 {
		from = uint64(curTime - YearSec + WeekSec - curTime%WeekSec)
	}
	if to == 0 {
		to = uint64(curTime)
	}

	var vols []*types.VolWeek
	err := k.dbHandler.Table(app.TRADE_TABLE).
		Select("DATE_PART('week', to_timestamp(trade_timestamp)) AS week, "+
			"EXTRACT(YEAR FROM to_timestamp(trade_timestamp)) AS year, "+
			"COALESCE(SUM(base_volume), 0) AS base_volume, "+
			"COALESCE(SUM(target_volume)) AS target_volume, "+
			"COALESCE(SUM(volume_in_usd)) AS usd_volume").
		Where("ticker_id = ?", ticker.TickerId).
		Where("trade_timestamp >= ?", from).
		Where("trade_timestamp <= ?", to).
		Group("year, week").
		Order("year DESC, week DESC").
		Scan(&vols).Error
	if err != nil {
		return &types.TradingVolumePerWeekResponse{}, err
	}
	volInfos := make([]*types.TradeVolumeByWeek, 0)
	for _, vol := range vols {
		volInfos = append(volInfos, &types.TradeVolumeByWeek{
			Week: &types.Week{
				Week: uint32(vol.Week),
				Year: uint32(vol.Year),
			},
			TokenAVolume: uint64(vol.BaseVolume),
			TokenBVolume: uint64(vol.TargetVolume),
			UsdVolume: vol.USDVolume / 10000000,
		})
	}
	return &types.TradingVolumePerWeekResponse{
		TradingVolume: volInfos,
	}, nil
}

func (k Keeper) TradingVolumePerMonth(ctx context.Context, request *types.TradingVolumePerMonthRequest) (*types.TradingVolumePerMonthResponse, error) {
	if request.ContractId == "" {
		return &types.TradingVolumePerMonthResponse{}, fmt.Errorf("error empty contract id")
	}
	// get ticker_id by contract_id
	var ticker tickertypes.Ticker
	k.dbHandler.Table(app.TICKER_TABLE).Where("pool_id = ?", request.ContractId).Scan(&ticker)

	from := request.From
	to := request.To

	curTime := time.Now().Unix()

	if from == 0 {
		from = uint64(time.Now().AddDate(0, -12, 0).Unix())
	}
	if to == 0 {
		to = uint64(curTime)
	}
	var vols []*types.VolMonth
	err := k.dbHandler.Table(app.TRADE_TABLE).
		Select("EXTRACT(MONTH FROM to_timestamp(trade_timestamp)) AS month, "+
			"EXTRACT(YEAR FROM to_timestamp(trade_timestamp)) AS year, "+
			"COALESCE(SUM(base_volume), 0) AS base_volume, "+
			"COALESCE(SUM(target_volume)) AS target_volume, "+
			"COALESCE(SUM(volume_in_usd)) AS usd_volume").
		Where("ticker_id = ?", ticker.TickerId).
		Where("trade_timestamp >= ?", from).
		Where("trade_timestamp <= ?", to).
		Group("year, month").
		Order("year DESC, month DESC").
		Scan(&vols).Error
	if err != nil {
		return &types.TradingVolumePerMonthResponse{}, err
	}
	volInfos := make([]*types.TradeVolumeByMonth, 0)
	for _, vol := range vols {
		volInfos = append(volInfos, &types.TradeVolumeByMonth{
			Month: &types.Month{
				Month: uint32(vol.Month),
				Year: uint32(vol.Year),
			},
			TokenAVolume: uint64(vol.BaseVolume),
			TokenBVolume: uint64(vol.TargetVolume),
			UsdVolume: vol.USDVolume / 10000000,
		})
	}
	return &types.TradingVolumePerMonthResponse{
		TradingVolume: volInfos,
	}, nil
}

func (k Keeper) TradingVolumePerDay(ctx context.Context, request *types.TradingVolumePerDayRequest) (*types.TradingVolumePerDayResponse, error) {
	if request.ContractId == "" {
		return &types.TradingVolumePerDayResponse{}, fmt.Errorf("error empty contract id")
	}
	// get ticker_id by contract_id
	var ticker tickertypes.Ticker
	k.dbHandler.Table(app.TICKER_TABLE).Where("pool_id = ?", request.ContractId).Scan(&ticker)

	from := request.From
	to := request.To

	curTime := time.Now().Unix()

	if from == 0 {
		from = uint64(curTime - WeekSec + DaySec - curTime%DaySec)
	}
	if to == 0 {
		to = uint64(curTime)
	}
	var vols []*types.VolDay
	err := k.dbHandler.Table(app.TRADE_TABLE).
		Select("EXTRACT(DAY FROM to_timestamp(trade_timestamp)) AS day, "+
			"EXTRACT(MONTH FROM to_timestamp(trade_timestamp)) AS month, "+
			"EXTRACT(YEAR FROM to_timestamp(trade_timestamp)) AS year, "+
			"COALESCE(SUM(base_volume), 0) AS base_volume, "+
			"COALESCE(SUM(target_volume)) AS target_volume, "+
			"COALESCE(SUM(volume_in_usd)) AS usd_volume").
		Where("ticker_id = ?", ticker.TickerId).
		Where("trade_timestamp >= ?", from).
		Where("trade_timestamp <= ?", to).
		Group("year, month, day").
		Order("year DESC, month DESC, day DESC").
		Scan(&vols).Error
	if err != nil {
		return &types.TradingVolumePerDayResponse{}, err
	}
	volInfos := make([]*types.TradeVolumeByDate, 0)
	for _, vol := range vols {
		volInfos = append(volInfos, &types.TradeVolumeByDate{
			Date: &types.Date{
				Day: uint32(vol.Day),
				Month: uint32(vol.Month),
				Year: uint32(vol.Year),
			},
			TokenAVolume: uint64(vol.BaseVolume),
			TokenBVolume: uint64(vol.TargetVolume),
			UsdVolume: vol.USDVolume / 10000000,
		})
	}
	return &types.TradingVolumePerDayResponse{
		TradingVolume: volInfos,
	}, nil
}

func (k Keeper) TradingVolumePerHour(ctx context.Context, request *types.TradingVolumePerHourRequest) (*types.TradingVolumePerHourResponse, error) {
	if request.ContractId == "" {
		return &types.TradingVolumePerHourResponse{}, fmt.Errorf("error empty contract id")
	}
	// get ticker_id by contract_id
	var ticker tickertypes.Ticker
	k.dbHandler.Table(app.TICKER_TABLE).Where("pool_id = ?", request.ContractId).Scan(&ticker)

	from := request.From
	to := request.To

	curTime := time.Now().Unix()

	if from == 0 {
		from = uint64(curTime - DaySec + HourSec - curTime%HourSec)
	}
	if to == 0 {
		to = uint64(curTime)
	}
	var vols []*types.VolHour
	err := k.dbHandler.Table(app.TRADE_TABLE).
		Select("EXTRACT(HOUR FROM to_timestamp(trade_timestamp)) AS hour, "+
			"EXTRACT(DAY FROM to_timestamp(trade_timestamp)) AS day, "+
			"EXTRACT(MONTH FROM to_timestamp(trade_timestamp)) AS month, "+
			"EXTRACT(YEAR FROM to_timestamp(trade_timestamp)) AS year, "+
			"COALESCE(SUM(base_volume), 0) AS base_volume, "+
			"COALESCE(SUM(target_volume)) AS target_volume, "+
			"COALESCE(SUM(volume_in_usd)) AS usd_volume").
		Where("ticker_id = ?", ticker.TickerId).
		Where("trade_timestamp >= ?", from).
		Where("trade_timestamp <= ?", to).
		Group("year, month, day, hour").
		Order("year DESC, month DESC, day DESC, hour DESC").
		Scan(&vols).Error
	if err != nil {
		return &types.TradingVolumePerHourResponse{}, err
	}
	volInfos := make([]*types.TradeVolumeByHour, 0)
	for _, vol := range vols {
		volInfos = append(volInfos, &types.TradeVolumeByHour{
			Time: &types.TimeHour{
				Hour: uint32(vol.Hour),
				Date: &types.Date{
					Day: uint32(vol.Day),
					Month: uint32(vol.Month),
					Year: uint32(vol.Year),
				},
			},
			TokenAVolume: uint64(vol.BaseVolume),
			TokenBVolume: uint64(vol.TargetVolume),
			UsdVolume: vol.USDVolume / 10000000,
		})
	}
	return &types.TradingVolumePerHourResponse{
		TradingVolume: volInfos,
	}, nil
}

func (k Keeper) PriceGraph(ctx context.Context, request *types.PriceGraphRequest) (*types.PriceGraphResponse, error) {
	if request.ContractId == "" {
		return &types.PriceGraphResponse{}, fmt.Errorf("error empty contract id")
	}

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
	if request.ContractId == "" {
		return &types.PriceGraphLastWeekResponse{}, fmt.Errorf("error empty contract id")
	}

	to := time.Now().Unix()
	from := time.Now().Add(-Week).Unix()

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
	if request.ContractId == "" {
		return &types.PriceGraphLastMonthResponse{}, fmt.Errorf("error empty contract id")
	}

	to := time.Now().Unix()
	from := time.Now().Add(-Month).Unix()

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
	if request.ContractId == "" {
		return &types.PriceGraphLastYearResponse{}, fmt.Errorf("error empty contract id")
	}

	to := time.Now().Unix()
	from := time.Now().Add(-Year).Unix()

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
	if request.Address == "" {
		return &types.TradeHistoricalResponse{}, fmt.Errorf("error empty address")
	}

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

func (k Keeper) LastWeekTradeHistoricals(ctx context.Context, request *types.LastWeekTradeHistoricalRequest) (*types.LastWeekTradeHistoricalResponse, error) {
	to := time.Now().Unix()
	from := time.Now().Add(-Week).Unix()

	if request.Address == "" {
		return &types.LastWeekTradeHistoricalResponse{}, fmt.Errorf("error empty address")
	}

	var trades []*types.Trade

	query := k.dbHandler.Table(app.TRADE_TABLE).
		Order("trade_timestamp DESC").
		Where("maker = ?", request.Address).
		Where("trade_timestamp >= ?", from).
		Where("trade_timestamp <= ?", to)

	err := query.Find(&trades).Error
	if err != nil {
		return &types.LastWeekTradeHistoricalResponse{}, err
	}

	var tradeInfos []*types.TradeInfo

	for _, trade := range trades {
		tradeInfos = append(tradeInfos, convertToInfo(trade))
	}

	return &types.LastWeekTradeHistoricalResponse{
		Trades: tradeInfos,
	}, nil
}

func (k Keeper) LastMonthTradeHistoricals(ctx context.Context, request *types.LastMonthTradeHistoricalRequest) (*types.LastMonthTradeHistoricalResponse, error) {
	to := time.Now().Unix()
	from := time.Now().Add(-Month).Unix()

	if request.Address == "" {
		return &types.LastMonthTradeHistoricalResponse{}, fmt.Errorf("error empty address")
	}

	var trades []*types.Trade

	query := k.dbHandler.Table(app.TRADE_TABLE).
		Order("trade_timestamp DESC").
		Where("maker = ?", request.Address).
		Where("trade_timestamp >= ?", from).
		Where("trade_timestamp <= ?", to)

	err := query.Find(&trades).Error
	if err != nil {
		return &types.LastMonthTradeHistoricalResponse{}, err
	}

	var tradeInfos []*types.TradeInfo

	for _, trade := range trades {
		tradeInfos = append(tradeInfos, convertToInfo(trade))
	}

	return &types.LastMonthTradeHistoricalResponse{
		Trades: tradeInfos,
	}, nil
}

func (k Keeper) LastYearTradeHistoricals(ctx context.Context, request *types.LastYearTradeHistoricalRequest) (*types.LastYearTradeHistoricalResponse, error) {
	to := time.Now().Unix()
	from := time.Now().Add(-Year).Unix()

	if request.Address == "" {
		return &types.LastYearTradeHistoricalResponse{}, fmt.Errorf("error empty address")
	}

	var trades []*types.Trade

	query := k.dbHandler.Table(app.TRADE_TABLE).
		Order("trade_timestamp DESC").
		Where("maker = ?", request.Address).
		Where("trade_timestamp >= ?", from).
		Where("trade_timestamp <= ?", to)

	err := query.Find(&trades).Error
	if err != nil {
		return &types.LastYearTradeHistoricalResponse{}, err
	}

	var tradeInfos []*types.TradeInfo

	for _, trade := range trades {
		tradeInfos = append(tradeInfos, convertToInfo(trade))
	}

	return &types.LastYearTradeHistoricalResponse{
		Trades: tradeInfos,
	}, nil
}

func (k Keeper) TotalTrades(ctx context.Context, request *types.TotalTradesRequest) (*types.TotalTradesResponse, error) {
	count := int64(0)
	err := k.dbHandler.Table(app.TRADE_TABLE).Count(&count).Error
	if err != nil {
		return &types.TotalTradesResponse{
			TotalTrades: count,
		}, err
	}

	return &types.TotalTradesResponse{
		TotalTrades: count,
	}, nil
}

func (k Keeper) MostTraded(ctx context.Context, request *types.MostTradedRequest) (*types.MostTradedResponse, error) {
	var tokens []*types.Token
	err := k.dbHandler.Table(app.TOKEN_TABLE).Find(&tokens).Error
	if err != nil {
		return &types.MostTradedResponse{}, err
	}
	maxVol := float64(0)
	maxIdx := 0
	oneDayAgo := time.Now().Unix() - 86400
	for idx, token := range tokens {
		tickers := []string{}
		err := k.dbHandler.Table(app.TICKER_TABLE).
			Where("base_currency = ?", token.TokenName).
			Or("target_currency = ?", token.TokenName).
			Select("ticker_id").
			Find(&tickers).Error
		if err != nil {
			return &types.MostTradedResponse{}, err
		}
		volume := float64(64)
		for _, ticker := range tickers {
			poolVol := float64(0)
			err = k.dbHandler.Table(app.TRADE_TABLE).
				Where("ticker_id = ?", ticker).
				Where("trade_timestamp >= ?", oneDayAgo).
				Select("COALESCE(sum(volume_in_usd), 0) as total").Scan(&poolVol).Error
			if err != nil {
				return &types.MostTradedResponse{}, err
			}
			volume += poolVol
		}
		if maxVol < volume {
			maxIdx = idx
			maxVol = volume
		}
	}
	return &types.MostTradedResponse{
		Asset:     tokens[maxIdx].Symbol,
		UsdVolume: maxVol / 10000000,
	}, nil
}
