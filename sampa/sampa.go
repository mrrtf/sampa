package sampa

import (
	"fmt"

	"github.com/aphecetche/bitset"
)

// Cluster describes a Sampa cluster, i.e. a set of
// ADC samples.
// Note that we use ints whereas each value really
// is 10 bits (or 20 bits for samples in sum mode)
type Cluster struct {
	n       int   // number of samples
	ts      int   // timestamp
	samples []int // samples
}

// Payload is a bitset with some helper methods to
// split it into 10-bits ints
type Payload struct {
	bitset.BitSet
}

func Decode(clusters []Cluster, data *Payload) {
	tb := data.Split()
	fmt.Println("Decode : slice size=", len(tb))
	for _, t := range tb {
		fmt.Printf("%v ", t)
	}
	fmt.Println()
}

// Split splits the payload bitset into a slice of 10-bits integers
func (buf *Payload) Split() []int {
	tenbits := make([]int, buf.BitSet.Length()/10)
	var i int = 0
	for offset := 0; offset < buf.BitSet.Length(); offset += 10 {
		tenbits[i] = int(buf.Uint16(offset, offset+10) & 0x3FF)
		i++
	}
	return tenbits
}

func (payload *Payload) Clear() {
	payload.BitSet.Clear()
}
