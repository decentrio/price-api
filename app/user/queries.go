package user

import (
	"context"
	"fmt"

	app "github.com/decentrio/price-api/app"
	types "github.com/decentrio/price-api/types/user"
)

var _ types.UserQueryServer = Keeper{}

func (k Keeper) Activities(ctx context.Context, request *types.ActivitiesRequest) (*types.ActivitiesResponse, error) {
	if request.Address == "" {
		return &types.ActivitiesResponse{}, fmt.Errorf("error empty address")
	}

	var activityInfos []*types.ActivityInfo
	err := k.dbHandler.Table(app.ACTIVITIES_TABLE).
		Order("time_stamp DESC").
		Where("address = ?", request.Address).
		Find(&activityInfos).Error
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

func convertToInfo(act *types.ActivityInfo) *types.Activity {
	baseVol := float64(act.BaseVolume) / 10000000.0
	targetVol := float64(act.TargetVolume) / 10000000.0

	return &types.Activity{
		User:           act.User,
		ActionType:     act.ActionType,
		BaseCurrency:   act.BaseCurrency,
		BaseVolume:     baseVol,
		TargetCurrency: act.TargetCurrency,
		TargetVolume:   targetVol,
		TimeStamp:      act.TimeStamp,
	}
}
