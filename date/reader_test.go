package date

import (
	"fmt"
	"log"
	"testing"
	"time"
)

const (
	NBENCH = 50
	NTEST  = 10000
)

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func get(nevents int) {
	defer timeTrack(time.Now(), fmt.Sprintf("get(%d)", nevents))
	dr := NewReader("/Users/laurent/o2/sampa/syn_then_trig_20170210_1833")
	i := 0
	for ; nevents >= 0; nevents-- {
		err := dr.GetNextEvent()
		if err != nil {
			break
		}
		i++
	}
	log.Printf("read %d events", i)
}

func get2(nevents int) {
	defer timeTrack(time.Now(), fmt.Sprintf("get2(%d)", nevents))
	dr := NewReader("/Users/laurent/o2/sampa/syn_then_trig_20170210_1833")
	i := 0
	for ; nevents >= 0; nevents-- {
		err := dr.GetNextEvent2()
		if err != nil {
			break
		}
		i++
	}
	log.Printf("read %d events", i)
}

// func TestGetNextEvent(t *testing.T) {
// 	get(NTEST)
// }

func TestGetNextEvent2(t *testing.T) {
	get2(NTEST)
}
func BenchmarkGetNextEvent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		get(NBENCH)
	}
}

func BenchmarkGetNextEvent2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		get2(NBENCH)
	}
}
