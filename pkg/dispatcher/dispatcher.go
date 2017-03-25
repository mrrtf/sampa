package dispatcher

import (
	"fmt"

	"github.com/aphecetche/sampa/gbt"
	"github.com/aphecetche/sampa/sampa"
)

// nBitsPerChannel is the number of bits a channel uses in a 80-bits GBT word
const nBitsPerChannel int = 2

// FanOut splits a 80-bits GBT word into n elink data groups
// of 80/n bits
func FanOut(g gbt.Word, elinks []sampa.Payload) error {
	for ibit := 0; ibit < 2; /*80*/ ibit += nBitsPerChannel {
		ch := &elinks[ibit/nBitsPerChannel]
		for i := nBitsPerChannel - 1; i >= 0; i-- {
			err := ch.Append(g.Get(ibit + i))
			if err != nil {
				panic(fmt.Sprintf("FanOut ibit %v i %v err %v",
					ibit, i, err))
			}
		}
	}
	return nil
}
