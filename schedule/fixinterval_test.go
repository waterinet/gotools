package cron

import (
	"fmt"
	"testing"
	"time"
)

func TestFixSchedule(t *testing.T) {

	return
	fs := new(FixedIntervalSchedule)
	fs.Every(time.Second*5 + time.Millisecond*10 + time.Nanosecond*100)
	fmt.Println("Interval:", fs.Interval)

	now := time.Now()
	for i := 0; i < 10; i++ {
		next := fs.Next(now)
		fmt.Println("next:", next)
		now = <-time.After(next.Sub(now))
		fmt.Println("now :", now)
	}
}
