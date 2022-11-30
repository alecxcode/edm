package datetime

import (
	"strconv"
	"strings"
	"time"
)

// DateTime defines time in format: (year, month, day, hour, minute)
type DateTime struct {
	Year   int
	Month  byte
	Day    byte
	Hour   byte
	Minute byte
}

// Int64 timestamps conversion functions
func DateTimeToInt64(dt DateTime) int64 {
	var AD int64 = 1
	if dt.Year < 0 {
		dt.Year = -dt.Year
		AD = -1
	}
	return (int64(dt.Minute) + int64(dt.Hour)*60 + int64(dt.Day-1)*60*24 + int64(dt.Month-1)*60*24*31 + int64(dt.Year)*60*24*31*12) * AD
}
func Int64ToDateTime(timestamp int64) DateTime {
	if timestamp == 0 {
		return DateTime{0, 0, 0, 0, 0}
	}
	var AD int64 = 1
	if timestamp < 0 {
		timestamp = -timestamp
		AD = -1
	}
	return DateTime{
		Minute: byte(timestamp % 60),
		Hour:   byte(timestamp / 60 % 24),
		Day:    byte(timestamp/(60*24)%31) + 1,
		Month:  byte(timestamp/(60*24*31)%12) + 1,
		Year:   int(timestamp / (60 * 24 * 31 * 12) * AD),
	}
}

func ConvDateTimeStrToInt64(dt string) int64 {
	return DateTimeToInt64(StringToDateTime(dt))
}

func StringToDateTime(dtstring string) DateTime {
	BC := false
	if dtstring == "" {
		return DateTime{0, 0, 0, 0, 0}
	}
	if dtstring[0] == '-' {
		dtstring = strings.TrimLeft(dtstring, "-")
		BC = true
	}
	dtstring = strings.Replace(dtstring, "T", " ", 1)
	dtarr := strings.Split(dtstring, " ")
	if len(dtarr) < 2 {
		return DateTime{0, 0, 0, 0, 0}
	}
	darr := strings.Split(dtarr[0], "-")
	if len(darr) < 3 {
		return DateTime{0, 0, 0, 0, 0}
	}
	y, _ := strconv.Atoi(darr[0])
	m, _ := strconv.Atoi(darr[1])
	d, _ := strconv.Atoi(darr[2])
	if BC {
		y = 0 - y
	}
	tarr := strings.Split(dtarr[1], ":")
	if len(tarr) < 2 {
		return DateTime{y, byte(m), byte(d), 0, 0}
	}
	hour, _ := strconv.Atoi(tarr[0])
	minute, _ := strconv.Atoi(tarr[1])
	return DateTime{y, byte(m), byte(d), byte(hour), byte(minute)}
}

func TimeToString12(hour byte, minute byte) string {

	if hour == 0 && minute == 0 {
		return "12 midnight"
	}
	if hour == 12 && minute == 0 {
		return "12 noon"
	}

	var ampm string
	if hour < 12 {
		ampm = " am"
	} else {
		ampm = " pm"
	}
	if hour > 12 {
		hour = hour - 12
	}
	if hour == 0 {
		hour = 12
	}
	h := strconv.Itoa(int(hour))
	m := strconv.Itoa(int(minute))
	if minute < 10 {
		m = "0" + m
	}

	return h + ":" + m + ampm
}

func TimeToString24(hour byte, minute byte) string {
	h := strconv.Itoa(int(hour))
	m := strconv.Itoa(int(minute))
	if hour < 10 {
		h = "0" + h
	}
	if minute < 10 {
		m = "0" + m
	}
	return h + ":" + m
}

func DateTimeToStringSTD(dt DateTime) string {
	if dt.Day == 0 {
		return ""
	}
	return DateToString(Date{dt.Year, dt.Month, dt.Day}, "") + "yyyy-mm-dd" + TimeToString24(dt.Hour, dt.Minute)
}

func GetCurrentDateTime() DateTime {
	t := time.Now()
	y, m, d := t.Date()
	hour := t.Hour()
	minute := t.Minute()
	return DateTime{y, byte(m), byte(d), byte(hour), byte(minute)}
}

// UTC functions for possible future use
// func getCurrentDateTimeUTC() DateTime {
// 	t := time.Now().UTC()
// 	y, m, d := t.Date()
// 	hour := t.Hour()
// 	minute := t.Minute()
// 	return DateTime{y, byte(m), byte(d), byte(hour), byte(minute)}
// }
// func getUTCDiff() int64 {
// 	return dateTimeToInt64(getCurrentDateTime()) - dateTimeToInt64(getCurrentDateTimeUTC())
// }
// func (dt DateTime) minus(minutesdiff int64) DateTime {
// 	if dt.Day == 0 {
// 		return DateTime{0, 0, 0, 0, 0}
// 	}
// 	return int64ToDateTime(dateTimeToInt64(dt) - minutesdiff)
// }
// func (dt DateTime) plus(minutesdiff int64) DateTime {
// 	if dt.Day == 0 {
// 		return DateTime{0, 0, 0, 0, 0}
// 	}
// 	return int64ToDateTime(dateTimeToInt64(dt) + minutesdiff)
// }
