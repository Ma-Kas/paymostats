package cli

import "time"

type rangeSpec struct {
	label  string
	days   int // 0 = all time
	custom func() (time.Time, time.Time)
}

var choices = map[string]rangeSpec{
	"a": {label: "Last week", days: 7},
	"b": {label: "Last two weeks", days: 14},
	"c": {label: "Last month", days: 30},
	"d": {label: "Last 3 months", days: 90},
	"e": {label: "Last 6 months", days: 180},
	"f": {
		label: "Year to date",
		custom: func() (time.Time, time.Time) {
			now := time.Now()
			start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
			return start, now
		},
	},
	"g": {label: "All Time", days: 0},
}

func bounds(spec rangeSpec) (time.Time, time.Time) {
	if spec.custom != nil {
		return spec.custom()
	}
	end := time.Now().UTC()
	if spec.days == 0 {
		return time.Unix(0, 0), end
	}
	return end.AddDate(0, 0, -spec.days), end
}
