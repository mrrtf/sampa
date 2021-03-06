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
	for i := 0; i < 1; i++ { //FIXME: 1 should be len(bytes)=10=nBytesPerGBT
		b := uint(bytes[i])
		for j := uint(0); j < 8; j += nBitsPerChannel { //FIXME:should be 8/nBitsPerChannel
			ch := elinks[elink]
			if elinkmask&(uint64(1)<<elink) > 0 {
				// skip masked-out elinks
				elink++
				continue
			}
			elink++
			mask := uint(1) << (j + 1)
			// fmt.Println("j=", j, "will append to elink ", elink, "using mask", mask, mask/2)
			bit0 := (b & mask) > 0
			mask /= 2
			bit1 := (b & mask) > 0
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
