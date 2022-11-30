package currencies

import (
	"strconv"
	"strings"
)

// ProcessFormSumInt accepts decimal or non-decimal string and returns its integer representation multiplied by 100
func ProcessFormSumInt(s string) (sum int) {
	if !strings.Contains(s, ".") {
		s += ".00"
	} else if s[len(s)-2] == '.' {
		s += "0"
	}
	if strings.Contains(s, ".") && s[len(s)-3] != '.' {
		tsa := strings.Split(s, ".")
		if len(tsa[1]) == 1 {
			tsa[1] = tsa[1] + "0"
		}
		s = tsa[0] + tsa[1][:2]
	}
	sum, _ = strconv.Atoi(strings.Replace(s, ".", "", -1))
	return sum
}

// ToDecimalStr accepts decimal or non-decimal string and returns its string representation with two fraction digits
func ToDecimalStr(s string) string {
	var addminus = false
	if strings.HasPrefix(s, "-") {
		s = strings.TrimPrefix(s, "-")
		addminus = true
	}
	if len(s) == 1 {
		s = "0.0" + s
	} else if len(s) == 2 {
		s = "0." + s
	} else {
		i := len(s) - 2
		s = s[:i] + "." + s[i:]
	}
	if addminus {
		s = "-" + s
	}
	return s
}
