package sampa

import (
	"errors"
	"fmt"
	"log"
)

var (
	ErrIncorrectSize = errors.New("sampa: incorrect GBT size")
)

type ELink interface {
	Append(bit0, bit1 bool) (*Packet, error)
	// Clear()
	// ForceClear()
	IsEmpty() bool
	Id() int
	Size() int
}

const (
	HeaderSize int = 50
	// nBitsPerChannel is the number of bits a channel uses in a 80-bits GBT word
	nBitsPerChannel uint = 2
	nBytesPerGBT    int  = 10
)

// Dispatch splits the 10 bytes composing a 80 bits GBT word
// into n elink data groups of 80/n bits
func Dispatch(bytes []byte, elinks []ELink, elinkmask uint64) error {
	if len(bytes) != nBytesPerGBT {
		return ErrIncorrectSize
	}
	var elink uint64 = 0
	nbytes := 1 /*len(bytes)*/
	// FIXME:
	// either be fast enough and use len(bytes) whatever the number of sampa per solar is
	// (i.e. absorb zeros easily),
	// or use some configuration to know how many sampas are to be read...
	for i := 0; i < nbytes; i++ {
		b := uint(bytes[i])
		for j := uint(0); j < 8; j += nBitsPerChannel {
			ch := elinks[elink]
			// fmt.Println("elink(", ch.Id(), ")=", elink, ch.Size())
			if elinkmask&(uint64(1)<<elink) > 0 {
				// skip masked-out elinks
				elink++
				continue
			}
			maskbit0 := uint(1) << (j + 1)
			bit0 := (b & maskbit0) > 0
			maskbit1 := uint(1) << j
			bit1 := (b & maskbit1) > 0
			// fmt.Println("j=", j, "will append to elink(Id=)", ch.Id(), elink, " of size ", ch.Size(), "using masks ", maskbit0, maskbit1)
			elink++
			// if ch.IsEmpty() && bit0 == bit1 && bit0 == false {
			// 	// only start to append bits if there's a 1
			// 	continue
			// }
			packet, err := ch.Append(bit0, bit1)
			if err != nil {
				log.Fatalf("Dispatch error : byte %d elink %d", b, i)
			}
			if packet != nil {
				fmt.Println(packet.String())
			}
		}
	}
	return nil
}
