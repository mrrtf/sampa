package date

import (
	"fmt"
	"io"
	"log"
	"os"
	"testing"
	"time"
)

const (
	NBENCH   = 50
	NTEST    = 10000000
	TESTFILE = "/Users/laurent/o2/sampa/syn_then_trig_20170210_1833"
)

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func get(nevents int) {
	defer timeTrack(time.Now(), fmt.Sprintf("get(%d)", nevents))
	dr := NewReader(TESTFILE)
	if dr == nil {
		return
	}
	i := 0
	for ; nevents >= 0; nevents-- {
		err := dr.GetNextEvent()
		if err == io.EOF {
			break
		}
		i++
	}
	log.Printf("read %d events", i)
}

// TestReadGBT test the reading of DATE file, but driven by GBT word,
// not by event (TestGetNextEvent)
func TestReadGBT(t *testing.T) {
	nloop := 10000 * 10000
	defer timeTrack(time.Now(), fmt.Sprintf("TestReadGBT(%d)", nloop))
	file, err := os.Open(TESTFILE)
	if err != nil {
		t.Skip("Input raw data file not there. Skipping test.")
	}
	defer file.Close()
	dr := NewReader(TESTFILE)
	defer func() {
		fmt.Println(dr)
	}()
	for ; nloop >= 0; nloop-- {
		err := dr.NextGBT()
		// log.Println("GBT=", g.StringLSBRight())
		if err != nil {
			// log.Println("loop=", nloop, "err=", err, "ngbt=", dr.ngbt)
			if err == io.EOF {
				// log.Println("Reached EOF")
				break
			}
		}
	}
}

func TestGetNextEvent(t *testing.T) {
	get(NTEST)
}
func BenchmarkGetNextEvent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		get(NBENCH)
	}
}
func BenchmarkData2GBT(b *testing.B) {
	data := []byte{0X2, 0x1, 0xBB, 0xAA, 0x06, 0x05, 0x04, 0x03, 0x10, 0x09, 0x08, 0x07}
	gbt := make([]byte, 12)
	dr := NewReader(TESTFILE)
	if dr == nil {
		return
	}
	b.Run("regular", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			dr.Data2GBTHelper(gbt, data)
		}
	})
}

// func TestData2GBT(t *testing.T) {
// 	// the AA and AB are outside of the 80 bits limits and so should not
// 	// appear in the GBT word, that's normal
// 	data := []byte{0X2, 0x1, 0xBB, 0xAA, 0x06, 0x05, 0x04, 0x03, 0x10, 0x09, 0x08, 0x07}
// 	g := gbt.NewWord()
// 	Data2GBT(data, g)
// 	expected := "00000001000000100000001100000100000001010000011000000111000010000000100100010000"
// 	s := g.BitSet.StringLSBRight()
// 	if s != expected {
// 		t.Errorf("Expecting %s got %s", expected, s)
// 	}
// 	data = []byte{0xFF, 0xFF, 0, 0, 0x80, 0, 0, 0xFF, 0x1, 0, 0, 0xFF}
// 	expected = "11111111111111111111111100000000000000001000000011111111000000000000000000000001"
//
// 	Data2GBT(data, g)
// 	s = g.BitSet.StringLSBRight()
// 	if s != expected {
// 		t.Errorf("Expecting %s got %s", expected, s)
// 	}
// }
//
