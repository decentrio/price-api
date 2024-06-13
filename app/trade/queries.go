package trade

import (
	"context"
	"strconv"
	"strings"

	app "github.com/decentrio/price-api/app"
	types "github.com/decentrio/price-api/types/trade"
)

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
		BaseVolume:     strconv.FormatFloat(baseVol, 'f', 7, 64),
		TargetVolume:   strconv.FormatFloat(targetVol, 'f', 7, 64),
		TradeTimestamp: trade.TradeTimestamp * 1000,
		Type:           trade.TradeType,
	}
}
