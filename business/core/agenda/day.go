package agenda

import (
	"fmt"
	"strconv"
)

var (
	DaySunday    = Day{0}
	DayMonday    = Day{1}
	DayTuesday   = Day{2}
	DayWednesday = Day{3}
	DayThursday  = Day{4}
	DayFriday    = Day{5}
	DaySaturday  = Day{6}
)

var days = map[uint]Day{
	DaySunday.dow:    DaySunday,
	DayMonday.dow:    DayMonday,
	DayTuesday.dow:   DayTuesday,
	DayWednesday.dow: DayWednesday,
	DayThursday.dow:  DayThursday,
	DayFriday.dow:    DayFriday,
	DaySaturday.dow:  DaySaturday,
}

type Day struct {
	dow uint
}

func ParseDay(value uint) (Day, error) {
	d, exists := days[value]
	if !exists {
		return Day{}, fmt.Errorf("invalid day of week: %q", value)
	}

	return d, nil
}

func (d Day) DayOfWeedk() uint {
	return d.dow
}

func (d *Day) UnmarshalText(data []byte) error {
	dow, err := strconv.Atoi(string(data))
	if err != nil {
		return fmt.Errorf("converting day to int: %w", err)
	}

	dp, err := ParseDay(uint(dow))
	if err != nil {
		return err
	}

	d.dow = dp.dow
	return nil
}

func (d Day) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

func (d Day) String() string {
	return fmt.Sprintf("%d", d.dow)
}

func (d Day) Equal(d2 Day) bool {
	return d.dow == d2.dow
}
