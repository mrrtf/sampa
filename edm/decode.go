package edm

import (
	"errors"
	"fmt"
	"log"

	"github.com/aphecetche/bitset"
	"github.com/fatih/color"
)

var yellow = color.New(color.FgYellow).SprintFunc()
var blue = color.New(color.FgBlue).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()

var channels []bitset.BitSet

const nBitsPerChannel int = 100000

var bs0 = bitset.New(1024)

func GBTword(data EventDataType, bs *bitset.BitSet) error {

	if len(data) != 3 {
		return errors.New("3 32-bits words expected")
	}

	if bs.Size() != 80 {
		return errors.New(fmt.Sprintf("Expected 80 bits GBT words, not %d-bits one...", bs.Size()))
	}
	data2gbt := map[int][2]int{2: {0, 31}, 1: {32, 63}, 0: {64, 79}}
	for ix, bits := range data2gbt {
		for i := bits[0]; i <= bits[1]; i++ {
			bs.SetRangeFromUint32(bits[0], bits[1], data[ix])
		}
	}
	return nil
}

func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func dispatchGBTword(gbtWord *bitset.BitSet, channels []bitset.BitSet) error {

	const nbits int = 2

	var ibit int = 0

	// fmt.Println("GBT:", Reverse(gbtWord.String()), " (LSB on the right)")

	for ibit < 2 {

		for i := 0; i < nbits; i++ {

			bitContent := gbtWord.Get(ibit)

			err := channels[ibit/2].Append(bitContent)
			if err != nil {
				panic(err)
			}

			ibit++
		}
	}

	return nil
}

func ProcessEvent(event EventType) {

	if channels == nil {
		for i := 0; i < 40; i++ {
			channels = append(channels, *(bitset.New(nBitsPerChannel)))
		}
	} else {
		// FIXME: clear of channels should not occur at each event
		// but once a sync is found...
		for i := 0; i < 40; i++ {
			channels[0].Clear()
		}
	}

	fmt.Println(event)

	if !event.HasPayload() {
		return
	}

	_, err := event.SOP()

	if err != nil {
		fmt.Println(err)
		return
	}

	// n := 72
	// fmt.Println(event.data()[:n].String(3))

	gbt := bitset.New(80)

	// n := len(event.Data())
	n := 720

	for i := 0; i < n; i += 3 {

		err := GBTword(event.Data()[i:i+3], gbt)
		if err != nil {
			log.Fatal(err)
		}

		err = dispatchGBTword(gbt, channels)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("channels[0]=", channels[0].Length())

		ix, err := channels[0].Search(SyncPattern.BitSet)

		s := channels[0].String()

		if ix > 0 {

			log.Fatal("found !")
			fmt.Print(s[:ix])
			fmt.Print(blue(s[ix : ix+SyncPattern.Length()]))

			channels[0].PruneFirst(SyncPattern.Length())

			fmt.Print(blue("SYNC -"), channels[0].Length())
		}

		m := len(s) - 120
		if m < 0 {
			m = 0
		}
		fmt.Println(yellow(s[m:]))
	}

	fmt.Println()
}
