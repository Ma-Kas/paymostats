package cli

import "time"

type rangeSpec struct {
	label string
	days  int // 0 = forever
}

var choices = map[string]rangeSpec{
	"a": {"Last two weeks", 14},
	"b": {"Last month", 30},
	"c": {"Last 3 months", 30},
	"d": {"Last 6 months", 180},
	"e": {"Forever", 0},
}

func bounds(days int) (time.Time, time.Time) {
	end := time.Now().UTC()
	if days == 0 {
		return time.Unix(0, 0), end
	}
	return end.AddDate(0, 0, -days), end
}
