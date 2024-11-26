package trade

import "fmt"

func (x *TimeHour) ToString() string {
	return fmt.Sprintf("%s-H%d", x.Date.ToString(), x.Hour)
}

func (x *Date) ToString() string {
	return fmt.Sprintf("Y%d-M%d-D%d", x.Year, x.Month, x.Day)
}

func (x *Week) ToString() string {
	return fmt.Sprintf("Y%d-W%d", x.Year, x.Week)
}

func (x *Month) ToString() string {
	return fmt.Sprintf("Y%d-M%d", x.Year, x.Month)
}

type VolDay struct {
	Day          int
	Month        int
	Year         int
	BaseVolume   int64
	TargetVolume int64
	USDVolume    float64
}

type VolHour struct {
	Hour         int
	Day          int
	Month        int
	Year         int
	BaseVolume   int64
	TargetVolume int64
	USDVolume    float64
}

type VolWeek struct {
	Week         int
	Year         int
	BaseVolume   int64
	TargetVolume int64
	USDVolume    float64
}

type VolMonth struct {
	Month        int
	Year         int
	BaseVolume   int64
	TargetVolume int64
	USDVolume    float64
}
