package cron

import "time"

type FixedIntervalSchedule struct {
	Interval time.Duration
}

// The smallest unit supported is second
func (fs *FixedIntervalSchedule) Every(interval time.Duration) error {
	if interval < time.Second {
		fs.Interval = time.Second
		return nil
	}
	fs.Interval = time.Second * (interval / time.Second)
	return nil
}

// Target time should be rounded up to second
func (fs *FixedIntervalSchedule) Next(current time.Time) time.Time {
	return current.Add(fs.Interval - time.Duration(current.Nanosecond()))
}
