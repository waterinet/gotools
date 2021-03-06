package cron

import (
	"fmt"
	"testing"
	"time"
)

func TestSpecSchedule(t *testing.T) {

	ss := new(SpecSchedule)
	testSpec(ss, "* * * * * *")
	testSpec(ss, "1,3,5 1-10 * * * *")
	testSpec(ss, "0-4,8-12 * * * * *")
	testSpec(ss, "0 30 23 * 10,11,12 1-5")
	testSpec(ss, "*")
	testSpec(ss, "-1-2 * * * * *")
	testSpec(ss, "* 0,1, * * * 0-8")
	testSpec(ss, "60 * * * * *")
	testSpec(ss, "0 0 0 0-40 * *")
	testSpec(ss, "0 0 0 31 2 *")
	testSpec(ss, "*/1 * * * * *")
	testSpec(ss, "0 */5 9-18/2 * * *")
	testSpec(ss, "*/30 */60 * * * *")
	testSpec(ss, "*/ * * * * *")
	testSpec(ss, "* /* * * * *")
	testSpec(ss, "* * */* * * *")
	fmt.Println("=======================")
	testNext(ss, "0 30 23 * * *")
	testNext(ss, "0 30 0 * * 1-5")
	testNext(ss, "0 0-59/30 11-13 * * *")
	testNext(ss, "0 0 0 29 2 *")
	testNext(ss, "0 0,59 12 * * 5")
}

func testSpec(ss *SpecSchedule, spec string) {
	fmt.Printf(">>\"%s\"\n", spec)
	if err := ss.Parse(spec); err != nil {
		fmt.Println("Parse:", err)
	} else {
		fmt.Printf("%b\n%b\n%b\n%b\n%b\n%b\n", ss.Second, ss.Minute, ss.Hour, ss.Day, ss.Month, ss.Dow)
	}
}

func testNext(ss *SpecSchedule, spec string) {
	if err := ss.Parse(spec); err != nil {
		fmt.Println("Run:", err)
	}
	fmt.Printf(">>\"%s\"\n", spec)
	count := 0
	next := time.Now()
	for {
		next = ss.Next(next)
		fmt.Println("next:", next)
		if next.Before(time.Now()) {
			fmt.Println("error:", "invalid next time")
			break
		}
		tiker := time.NewTicker(time.Second * 1)
		<-tiker.C
		count++
		if count == 10 {
			break
		}
	}
}
