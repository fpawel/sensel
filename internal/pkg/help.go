package pkg

import (
	"fmt"
	"strconv"
	"time"
)

func FormatFloatTrimNulls(v float64, prec int) string {
	s := strconv.FormatFloat(v, 'f', prec, 64)
	if v != float64(int64(v)) {
		for len(s) > 0 && s[len(s)-1] == '0' {
			s = s[:len(s)-1]
		}
	}
	for len(s) > 0 && s[len(s)-1] == '.' {
		s = s[:len(s)-1]
	}
	return s
}

func FormatFloat(v float64, prec int) string {
	return strconv.FormatFloat(v, 'f', prec, 64)
}

func FormatDuration(d time.Duration) string {

	d = d.Round(time.Second)

	h := d / time.Hour
	d -= h * time.Hour

	m := d / time.Minute
	d -= m * time.Minute

	s := d / time.Second

	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
