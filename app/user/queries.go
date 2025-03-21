package user

import (
	"context"
	"fmt"
	"time"

	app "github.com/decentrio/price-api/app"
	types "github.com/decentrio/price-api/types/user"
)

var _ types.UserQueryServer = Keeper{}

func (k Keeper) Activities(ctx context.Context, request *types.ActivitiesRequest) (*types.ActivitiesResponse, error) {
	if request.Address == "" {
		return &types.ActivitiesResponse{}, fmt.Errorf("error empty address")
	}

	var activityInfos []*types.ActivityInfo
	query := k.dbHandler.Table(app.ACTIVITIES_TABLE).
		Order("timestamp DESC").
		Where("address = ?", request.Address).
		Distinct("address", "timestamp")
		
	if request.From != 0 {
		query = query.Where("timestamp >= ?", request.From)
	} else {
		query = query.Where("timestamp >= ?", time.Now().Unix() - 86400)
	}

	if request.To != 0 {
		query = query.Where("timestamp <= ?", request.To)
	} else {
		query = query.Where("timestamp <= ?", time.Now().Unix())
	}

	err := query.Find(&activityInfos).Error
	if err != nil {
		return &types.ActivitiesResponse{}, err
	}

	var activities []*types.Activity

	for _, activity := range activityInfos {
		activities = append(activities, convertToInfo(activity))
	}

	return &types.ActivitiesResponse{
		Activities: activities,
	}, nil
}

func (k Keeper) TotalUsers(ctx context.Context, request *types.TotalUsersRequest) (*types.TotalUsersResponse, error) {
	total := int64(0)
	user24H := int64(0)
	oneDayAgo := time.Now().Add(-24 * time.Hour).Unix()
	err := k.dbHandler.Table(app.ACTIVITIES_TABLE).Select("Count(distinct address)").Scan(&total).Error
	if err != nil {
		return &types.TotalUsersResponse{
			TotalUsers:    total,
			UsersLast_24H: user24H,
		}, err
	}
	err = k.dbHandler.Table(app.ACTIVITIES_TABLE).Where("timestamp >= ?", oneDayAgo).Select("Count(distinct address)").Scan(&user24H).Error
	if err != nil {
		return &types.TotalUsersResponse{
			TotalUsers:    total,
			UsersLast_24H: user24H,
		}, err
	}
	return &types.TotalUsersResponse{
		TotalUsers:    total,
		UsersLast_24H: user24H,
	}, nil
}

func convertToInfo(act *types.ActivityInfo) *types.Activity {
	baseVol := float64(act.BaseVolume) / 10000000.0
	targetVol := float64(act.TargetVolume) / 10000000.0

	return &types.Activity{
		User:           act.Address,
		ActionType:     act.ActionType,
		BaseCurrency:   act.BaseCurrency,
		BaseVolume:     baseVol,
		TargetCurrency: act.TargetCurrency,
		TargetVolume:   targetVol,
		TimeStamp:      act.Timestamp,
	}
}
