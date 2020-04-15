package pkg

import "strconv"

func FormatFloat(v float64, prec int) string {
	s := strconv.FormatFloat(v, 'f', prec, 64)
	for len(s) > 0 && s[len(s)-1] == '0' {
		s = s[:len(s)-1]
	}
	for len(s) > 0 && s[len(s)-1] == '.' {
		s = s[:len(s)-1]
	}
	return s
}
