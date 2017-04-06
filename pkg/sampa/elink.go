package sampa

import (
	"fmt"
	"log"

	"github.com/mrrtf/sampa/pkg/bitset"
)

// elink is a bitset with some helper methods to
// split it into 10-bits ints
type elink struct {
	bitset.BitSet
	checkpoint int
	indata     bool
	nsync      int
}

func NewELink() *elink {
	return &elink{BitSet: *(bitset.New(100000)), checkpoint: HeaderSize, indata: false, nsync: 0}
}

func (p *elink) String() string {
	return fmt.Sprintf("len %d checkpoint %d indata %v nsync %d",
		p.Length(), p.checkpoint, p.indata, p.nsync)
}

// Append adds two bits at the end of the bitset.
// If the resulting bitset's length reaches the checkpoint
// then the bitset is further processed by the Process method
func (p *elink) Append(bit0, bit1 bool) ([]Cluster, error) {
	err := p.BitSet.Append(bit0)
	if err != nil {
		return nil, err
	}
	err = p.BitSet.Append(bit1)
	if err != nil {
		return nil, err
	}
	// fmt.Println(p)
	if p.Length() != p.checkpoint {
		return nil, nil
	}
	return p.Process(), nil
}

func (p *elink) checkSync() {
	if p.nsync != 0 {
		panic("wrong logic 2")
	}
	if p.indata == true {
		panic("wrong logic 3")
	}
	sdh := SampaDataHeader{BitSet: *(p.BitSet.Last(HeaderSize))}
	if !sdh.IsEqual(SyncPattern.BitSet) {
		p.checkpoint += nBitsPerChannel
		return
	}
	if sdh.PKT() != uint8(SyncPKT) {
		log.Fatal("something's really wrong : a sync packet MUST have the correct packet type !")
	}
	p.Clear()
	p.checkpoint = HeaderSize
	p.nsync++
}

// Process attempts to interpret the current bitset
// as either a Sampa header or Sampa data
// If it's neither, then set the checkpoint at
// the current length + 2 bits
func (p *elink) Process() []Cluster {
	// fmt.Printf("===> Process %s\n", p.String())
	if p.Length() != p.checkpoint {
		panic("wrong logic somewhere")
	}

	// first things first : we must find the sync pattern, otherwise
	// just continue
	if p.nsync == 0 {
		p.checkSync()
		return nil
	}

	if p.indata {
		// data mode, just decode ourselves into
		// a set of sample clusters
		log.Fatal("in data !")
		clusters := p.Decode()
		p.Clear()
		p.checkpoint = HeaderSize
		p.indata = false
		return clusters
	}

	// looking for a header
	if p.checkpoint != HeaderSize {
		panic(fmt.Sprintf("wrong logic 5 checkpoint %d HeaderSize %d", p.checkpoint, HeaderSize))
	}

	sdh := SampaDataHeader{BitSet: *(p.BitSet.Last(HeaderSize))}
	fmt.Println(sdh.StringAnnotated("\n"))
	switch uint(sdh.PKT()) {
	case DataPKT:
		log.Fatal("DATA PACKET")
		fmt.Println(sdh.StringAnnotated("\n"))
		dataToGo := sdh.NumWords()
		fmt.Println(dataToGo, " 10-bits words to read")
		p.Clear()
		p.checkpoint = int(dataToGo * 10)
		p.indata = true
		return nil
	case HeartBeatPKT:
		log.Println("HEARBEAT found. Should be do sth about it ?")
		fallthrough
	case SyncPKT:
		p.nsync++
		fallthrough
	default:
		p.Clear()
		p.checkpoint = HeaderSize
		return nil
	}
	return nil
}

// Split splits the elink bitset into a slice of 10-bits integers
func (p *elink) Split() []int {
	tenbits := make([]int, p.BitSet.Length()/10)
	var i int = 0
	for offset := 0; offset < p.BitSet.Length(); offset += 10 {
		tenbits[i] = int(p.Uint16(offset, offset+10) & 0x3FF)
		i++
	}
	return tenbits
}

func (p *elink) Clear() {
	p.BitSet.Clear()
}

func (p *elink) Decode() []Cluster {
	tb := p.Split()
	fmt.Println("Decode : slice size=", len(tb))
	for _, t := range tb {
		fmt.Printf("%v ", t)
	}
	fmt.Println()
	return nil
}
