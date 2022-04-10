package main

import (
	"database/sql"
	"strconv"
	"strings"
	"time"
)

// Date defines date in format: (year, month, day)
type Date struct {
	Year  int
	Month byte
	Day   byte
}

// DateTime defines time in format: (year, month, day, hour, minute)
type DateTime struct {
	Year   int
	Month  byte
	Day    byte
	Hour   byte
	Minute byte
}

// Int64 datestamps timestamps conversion functions
func dateTimeToInt64(dt DateTime) int64 {
	var AD int64 = 1
	if dt.Year < 0 {
		dt.Year = -dt.Year
		AD = -1
	}
	return (int64(dt.Minute) + int64(dt.Hour)*60 + int64(dt.Day-1)*60*24 + int64(dt.Month-1)*60*24*31 + int64(dt.Year)*60*24*31*12) * AD
}
func int64ToDateTime(timestamp int64) DateTime {
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
func dateToInt64(d Date) int64 {
	var AD int64 = 1
	if d.Year < 0 {
		d.Year = -d.Year
		AD = -1
	}
	return ((int64(d.Year) * 31 * 12) + (int64(d.Month-1) * 31) + int64(d.Day-1)) * AD
}
func int64ToDate(datestamp int64) Date {
	var AD int64 = 1
	if datestamp < 0 {
		datestamp = -datestamp
		AD = -1
	}
	return Date{
		Day:   byte(datestamp%31) + 1,
		Month: byte(datestamp/31%12) + 1,
		Year:  int(datestamp / (31 * 12) * AD),
	}
}
func getValidDateFromSQL(dateval sql.NullInt64) Date {
	if dateval.Valid {
		return int64ToDate(dateval.Int64)
	}
	return Date{0, 0, 0}
}
func convDateStrToInt64(d string) int64 {
	return dateToInt64(stringToDate(d))
}
func convDateTimeStrToInt64(dt string) int64 {
	return dateTimeToInt64(stringToDateTime(dt))
}

// Common functions to process date and datetime
func stringToDate(idatestring string) Date {
	BC := false
	if idatestring == "" {
		return Date{0, 0, 0}
	}
	if idatestring[0] == '-' {
		idatestring = strings.TrimLeft(idatestring, "-")
		BC = true
	}
	arr := strings.Split(idatestring, "-")
	if len(arr) < 3 {
		return Date{0, 0, 0}
	}
	y, _ := strconv.Atoi(arr[0])
	m, _ := strconv.Atoi(arr[1])
	d, _ := strconv.Atoi(arr[2])
	if BC {
		y = 0 - y
	}
	return Date{y, byte(m), byte(d)}
}

func stringToDateTime(dtstring string) DateTime {
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

func timeToString12(hour byte, minute byte) string {

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

func timeToString24(hour byte, minute byte) string {
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

func dateTimeToStringSTD(dt DateTime) string {
	if dt.Day == 0 {
		return ""
	}
	return dateToString(Date{dt.Year, dt.Month, dt.Day}, "") + "yyyy-mm-dd" + timeToString24(dt.Hour, dt.Minute)
}

func processDateToStrings(idate Date) (y, m, d string) {
	y = strconv.Itoa(idate.Year)
	m = strconv.Itoa(int(idate.Month))
	d = strconv.Itoa(int(idate.Day))
	if idate.Month < 10 {
		m = "0" + m
	}
	if idate.Day < 10 {
		d = "0" + d
	}
	if idate.Year <= 999 && idate.Year >= 100 {
		y = "0" + y
	} else if idate.Year <= 99 && idate.Year >= 10 {
		y = "00" + y
	} else if idate.Year <= 9 && idate.Year >= 0 {
		y = "000" + y
	}
	return y, m, d
}

func dateToString(idate Date, dateFmt string) string {
	if idate.Day == 0 {
		return ""
	}
	y := strconv.Itoa(idate.Year)
	m := strconv.Itoa(int(idate.Month))
	d := strconv.Itoa(int(idate.Day))
	if idate.Month < 10 {
		m = "0" + m
	}
	if idate.Day < 10 {
		d = "0" + d
	}
	if idate.Year <= 999 && idate.Year >= 100 {
		y = "0" + y
	} else if idate.Year <= 99 && idate.Year >= 10 {
		y = "00" + y
	} else if idate.Year <= 9 && idate.Year >= 0 {
		y = "000" + y
	}
	switch dateFmt {
	case "yyyy-mm-dd":
		return y + "-" + m + "-" + d
	case "yyyy.mm.dd":
		return y + "." + m + "." + d
	case "dd.mm.yyyy":
		return d + "." + m + "." + y
	case "dd/mm/yyyy":
		return d + "/" + m + "/" + y
	case "Mon dd, yyyy":
		md := int(idate.Month)
		switch md {
		case 1:
			m = "Jan"
		case 2:
			m = "Feb"
		case 3:
			m = "Mar"
		case 4:
			m = "Apr"
		case 5:
			m = "May"
		case 6:
			m = "Jun"
		case 7:
			m = "Jul"
		case 8:
			m = "Aug"
		case 9:
			m = "Sep"
		case 10:
			m = "Oct"
		case 11:
			m = "Nov"
		case 12:
			m = "Dec"
		default:
			m = "M" + strconv.Itoa(md)
		}
		return m + " " + d + ", " + y
	case "mm/dd/yyyy":
		return m + "/" + d + "/" + y
	default:
		return y + "-" + m + "-" + d
	}
}

func dateToStringWOY(idate Date, monthBeforeDay bool) string {
	if idate.Day == 0 {
		return ""
	}
	md := int(idate.Month)
	d := strconv.Itoa(int(idate.Day))
	var m string
	switch md {
	case 1:
		m = "Jan"
	case 2:
		m = "Feb"
	case 3:
		m = "Mar"
	case 4:
		m = "Apr"
	case 5:
		m = "May"
	case 6:
		m = "Jun"
	case 7:
		m = "Jul"
	case 8:
		m = "Aug"
	case 9:
		m = "Sep"
	case 10:
		m = "Oct"
	case 11:
		m = "Nov"
	case 12:
		m = "Dec"
	default:
		m = "M" + strconv.Itoa(md)
	}
	if monthBeforeDay {
		return m + " " + d
	} else {
		return d + " " + m
	}
}

func getCurrentYearMStr() string {
	return time.Now().Format("2006-01")
}

func getCurrentDateTime() DateTime {
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
