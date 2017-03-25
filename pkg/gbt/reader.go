package gbt

import "github.com/aphecetche/sampa/pkg/bitset"

// Word is a 80-bits bitset
type Word struct {
	bitset.BitSet
}

func NewWord() *Word {
	return &Word{BitSet: *(bitset.New(80))}
}

type Reader interface {
	Word() (g Word, err error)
}
