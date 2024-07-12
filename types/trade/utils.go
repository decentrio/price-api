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
