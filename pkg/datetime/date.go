package datetime

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

func DateToInt64(d Date) int64 {
	var AD int64 = 1
	if d.Year < 0 {
		d.Year = -d.Year
		AD = -1
	}
	return ((int64(d.Year) * 31 * 12) + (int64(d.Month-1) * 31) + int64(d.Day-1)) * AD
}
func Int64ToDate(datestamp int64) Date {
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
func GetValidDateFromSQL(dateval sql.NullInt64) Date {
	if dateval.Valid {
		return Int64ToDate(dateval.Int64)
	}
	return Date{0, 0, 0}
}
func ConvDateStrToInt64(d string) int64 {
	return DateToInt64(StringToDate(d))
}

// Common functions to process date and datetime
func StringToDate(idatestring string) Date {
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

func DateToString(idate Date, dateFmt string) string {
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

func DateToStringWOY(idate Date, monthBeforeDay bool) string {
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

func GetCurrentYearMStr() string {
	return time.Now().Format("2006-01")
}
