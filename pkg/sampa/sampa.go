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
	Append(bit0, bit1 bool) ([]Cluster, error)
	Clear()
	ForceClear()
	IsEmpty() bool
}

const (
	HeaderSize int = 50
	// nBitsPerChannel is the number of bits a channel uses in a 80-bits GBT word
	nBitsPerChannel int = 2
	nBytesPerGBT    int = 10
)

// Dispatch splits the 10 bytes composing a 80 bits GBT word
// into n elink data groups of 80/n bits
func Dispatch(bytes []byte, elinks []ELink) error {
	if len(bytes) != nBytesPerGBT {
		return ErrIncorrectSize
	}
	elink := 0
	for i := 0; i < 1; i++ { //FIXME: 1 should be len(bytes)=10=nBytesPerGBT
		b := uint(bytes[i])
		for j := uint(0); j < 1; j++ { //FIXME: 1 should be 8/nBitsPerChannel
			ch := elinks[elink]
			elink++
			mask := uint(1) << (j + 1)
			bit0 := (b & mask) > 0
			mask /= 2
			bit1 := (b & mask) > 0
			clusters, err := ch.Append(bit0, bit1)
			if err != nil {
				log.Fatalf("Dispatch error : byte %d elink %d", b, i)
			}
			if len(clusters) > 0 {
				fmt.Println(clusters)
			}
		}
	}
	return nil
}
