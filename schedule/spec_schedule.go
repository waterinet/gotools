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
	if ss.Second, err = parseField(fields[0], bounds{0, 59}); err != nil {
		return errors.New("Second: " + err.Error())
	}
	if ss.Minute, err = parseField(fields[1], bounds{0, 59}); err != nil {
		return errors.New("Minute: " + err.Error())
	}
	if ss.Hour, err = parseField(fields[2], bounds{0, 23}); err != nil {
		return errors.New("Hour: " + err.Error())
	}
	if ss.Day, err = parseField(fields[3], bounds{1, 31}); err != nil {
		return errors.New("Day: " + err.Error())
	}
	if ss.Month, err = parseField(fields[4], bounds{1, 12}); err != nil {
		return errors.New("Month: " + err.Error())
	}
	if ss.Dow, err = parseField(fields[5], bounds{0, 6}); err != nil {
		return errors.New("Dow: " + err.Error())
	}
	return nil
}

func (ss *SpecSchedule) Next(current time.Time) time.Time {
	t := current.Add(time.Second - time.Duration(current.Nanosecond()))
	// indicate a field has been added
	added := false
	yearLimit := t.Year() + 1

REDO:
	if t.Year() > yearLimit {
		return time.Time{}
	}
	// test all-bit for quick check
	for !testBit(ss.Month, 63) && !testBit(ss.Month, int(t.Month())) {
		// if added, reset lower fields
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.Local)
		}
		t = t.AddDate(0, 1, 0)
		// the addition led to a carry, check higher fields
		if t.Month() == time.January {
			goto REDO
		}
	}
	// day of week is also checked here
	for !testDay(ss, t) {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
		}
		t = t.AddDate(0, 0, 1)
		if t.Day() == 1 {
			goto REDO
		}
	}
	for !testBit(ss.Hour, 63) && !testBit(ss.Hour, t.Hour()) {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, time.Local)
		}
		t = t.Add(time.Hour)
		if t.Hour() == 0 {
			goto REDO
		}
	}
	for !testBit(ss.Minute, 63) && !testBit(ss.Minute, t.Minute()) {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, time.Local)
		}
		t = t.Add(time.Minute)
		if t.Minute() == 0 {
			goto REDO
		}
	}
	for !testBit(ss.Second, 63) && !testBit(ss.Second, t.Second()) {
		if !added {
			added = true
		}
		t = t.Add(time.Second)
		if t.Second() == 0 {
			goto REDO
		}
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

func testDay(ss *SpecSchedule, t time.Time) bool {
	if testBit(ss.Dow, int(t.Weekday())) {
		if testBit(ss.Day, 63) || testBit(ss.Day, t.Day()) {
			return true
		}
	}
	return false
}
