package util

import (
	"time"
)

const (
	TIME_LAYOUT = "2006-01-02 15:04:05"
)

func UTC2Loc(utc time.Time) time.Time {
	y, mo, d := utc.Date()
	h, mi, s := utc.Clock()
	t := time.Date(y, mo, d, h, mi, s, utc.Nanosecond(), time.Local)
	return t
}

func IsSameDay(d1, d2 time.Time) bool {
	return d1.Year() == d2.Year() && d1.Month() == d2.Month() && d1.Day() == d2.Day()
}

func Insert(slice, insertion []interface{}, index int) []interface{} {

	if index == len(slice) || index == -1 {
		slice = append(slice, insertion...)
		return slice
	}

	slice = append(slice, insertion...)
	copy(slice[index+len(insertion):], slice[index:])
	copy(slice[index:len(insertion)+index], insertion)
	return slice
}
