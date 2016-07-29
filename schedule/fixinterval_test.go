package cron

import (
	"fmt"
	"testing"
	"time"
)

func TestFixSchedule(t *testing.T) {

	fs := new(FixedIntervalSchedule)
	fs.Every(time.Second*5 + time.Millisecond*10 + time.Nanosecond*100)
	fmt.Println("Interval:", fs.Interval)

	count := 0
	next := time.Now()
	for {
		next = fs.Next(next)
		fmt.Println("next:", next)
		<-time.After(time.Second * 1)
		count++
		if count == 10 {
			break
		}
	}
}
