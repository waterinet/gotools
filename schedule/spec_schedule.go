package cron

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

type SpecSchedule struct {
	Second, Minute, Hour, Day, Month, Dow uint64
}

type bounds struct {
	min, max int
}

type brange struct {
	min, max, step int
	all            bool
}

var (
	seconds = bounds{0, 59}
	minutes = bounds{0, 59}
	hours   = bounds{0, 23}
	days    = bounds{1, 31}
	months  = bounds{1, 12}
	dows    = bounds{0, 6}
)

// (*) all  : "* 0 * * * *" every first minute
// (-) range: "* 0-30 * * * *" every minite in the range [0,30] of an hour(0,1,2...30)
// (/) step : "* 0-30/2 * * * *" every two minites in the range [0,30] of an hour(0,2,4...30)
// (,) list : "* 1,3,5 * * * *" the minutes {1,3,5} of an hour
func (ss *SpecSchedule) Parse(spec string) error {
	fields := strings.Fields(spec)
	if len(fields) != 6 {
		return errors.New("invalid spec string")
	}

	var err error
	if ss.Second, err = parseField(fields[0], seconds); err != nil {
		return errors.New("Second: " + err.Error())
	}
	if ss.Minute, err = parseField(fields[1], minutes); err != nil {
		return errors.New("Minute: " + err.Error())
	}
	if ss.Hour, err = parseField(fields[2], hours); err != nil {
		return errors.New("Hour: " + err.Error())
	}
	if ss.Day, err = parseField(fields[3], days); err != nil {
		return errors.New("Day: " + err.Error())
	}
	if ss.Month, err = parseField(fields[4], months); err != nil {
		return errors.New("Month: " + err.Error())
	}
	if ss.Dow, err = parseField(fields[5], dows); err != nil {
		return errors.New("Dow: " + err.Error())
	}
	return nil
}

func (ss *SpecSchedule) Next(current time.Time) time.Time {
	t := current.Add(time.Second - time.Duration(current.Nanosecond()))

	for !testBit(ss.Month, 63) && !testBit(ss.Month, int(t.Month())) {
		t = t.AddDate(0, 1, 0)
		// set day to be first of that month
		t = t.AddDate(0, 0, 1-t.Day())
		//fmt.Println("testMonth")
	}
	// day of week is also checked here
	for !testBit(ss.Day, 63) && !testBit(ss.Day, t.Day()) && !testBit(ss.Dow, int(t.Weekday())) {
		t = t.AddDate(0, 0, 1)
		// set hour to be first of that day
		t = t.Add(-time.Duration(t.Hour()))
		//fmt.Println("testDay")
	}
	for !testBit(ss.Hour, 63) && !testBit(ss.Hour, t.Hour()) {
		t = t.Add(time.Hour)
		// set minute to be first of that hour
		t = t.Add(-time.Duration(t.Minute()))
		//fmt.Println("testHour")
	}
	for !testBit(ss.Minute, 63) && !testBit(ss.Minute, t.Minute()) {
		t = t.Add(time.Minute)
		// set second to be first of that minute
		t = t.Add(-time.Duration(t.Second()))
		//fmt.Println("testMinute")
	}
	for !testBit(ss.Second, 63) && !testBit(ss.Second, t.Second()) {
		t = t.Add(time.Second)
		//fmt.Println("testSecond", t.Second())
	}

	return t
}

func parseField(field string, b bounds) (uint64, error) {
	var val uint64 = 0
	parts := strings.Split(field, ",")
	for _, s := range parts {
		if r, err := getRange(s, b); err != nil {
			return val, err
		} else {
			setBits(&val, r)
		}
	}
	return val, nil
}

func getRange(s string, b bounds) (brange, error) {
	if s == "*" {
		return brange{b.min, b.max, 1, true}, nil
	}
	remain := s
	r := brange{0, 0, 1, false}
	// get step
	if strings.Contains(remain, "/") {
		parts := strings.Split(remain, "/")
		if len(parts) != 2 {
			return r, errors.New("invalid step")
		}
		if step, err := strconv.Atoi(parts[1]); err != nil {
			return r, errors.New("invalid step")
		} else {
			r.step = step
			remain = parts[0]
		}
	}
	// get min,max
	if strings.Contains(remain, "-") {
		parts := strings.Split(remain, "-")
		if len(parts) != 2 {
			return r, errors.New("invalid range")
		}
		if min, err := strconv.Atoi(parts[0]); err != nil {
			return r, errors.New("invalid range: min")
		} else if max, err := strconv.Atoi(parts[1]); err != nil {
			return r, errors.New("invalid range: max")
		} else if min < b.min || max > b.max {
			return r, errors.New("invalid range: out of bounds")
		} else {
			r.min = min
			r.max = max
			return r, nil
		}
	}
	// individual value
	if val, err := strconv.Atoi(remain); err != nil {
		return r, errors.New("invalid value")
	} else if val < b.min || val > b.max {
		return r, errors.New("invalid value: out of bounds")
	} else {
		r.min = val
		r.max = val
		return r, nil
	}
}

func setBits(val *uint64, r brange) {
	for i := r.min; i <= r.max; i += r.step {
		var v uint64 = 1
		*val |= v << uint(i)
	}
	if r.all {
		// use highest bit as all flag
		var v uint64 = 1
		*val |= v << 63
	}
}

func testBit(a uint64, p int) bool {
	var b uint64 = 1
	return a&(b<<uint(p)) > 0
}
