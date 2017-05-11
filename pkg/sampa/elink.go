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
	sdh        SampaDataHeader
}

func NewELink() *elink {
	return &elink{BitSet: *(bitset.New(100000)), checkpoint: HeaderSize, indata: false, nsync: 0}
}

func (p *elink) String() string {
	return fmt.Sprintf("len %d checkpoint %d indata %v nsync %d %s",
		p.Length(), p.checkpoint, p.indata, p.nsync, p.BitSet.StringLSBRight())
}

// Append adds 1 bit at the end of the bitset.
// If the resulting bitset's length reaches the checkpoint
// then the bitset is further processed by the Process method
func (p *elink) AppendBit(bit bool) ([]Cluster, error) {
	err := p.BitSet.Append(bit)
	if err != nil {
		return nil, err
	}
	if p.Length() != p.checkpoint {
		return nil, nil
	}
	return p.Process(), nil
}

// Append adds two bits at the end of the bitset.
// If the resulting bitset's length reaches the checkpoint
// then the bitset is further processed by the Process method
func (p *elink) Append(bit0, bit1 bool) ([]Cluster, error) {
	clusters0, err := p.AppendBit(bit0)
	if err != nil {
		return nil, err
	}
	clusters1, err := p.AppendBit(bit1)
	if err != nil {
		return nil, err
	}

	return append(clusters0, clusters1...), nil
}

// findSync tries to find a sync word in the last 50
// bits of the current elink bitset.
func (p *elink) findSync() {
	if p.nsync != 0 {
		panic("wrong logic 2")
	}
	if p.indata == true {
		panic("wrong logic 3")
	}

	sdh := SampaDataHeader{BitSet: *(p.BitSet.Last(HeaderSize))}

	if !sdh.IsEqual(SyncPattern.BitSet) {
		p.checkpoint++
		return
	}
	if sdh.PKT() != uint8(SyncPKT) {
		log.Fatal("something's really wrong : a sync packet MUST have the correct packet type !")
	}

	log.Println("findSync: found sync", p.nsync)
	p.Clear()
	p.checkpoint = HeaderSize
	p.nsync++
}

// Process attempts to interpret the current bitset
// as either a Sampa header or Sampa data
// If it's neither, then set the checkpoint at
// the current length + 2 bits
func (p *elink) Process() []Cluster {
	if p.Length() != p.checkpoint {
		panic("wrong logic somewhere")
	}

	// first things first : we must find the sync pattern, otherwise
	// just continue
	if p.nsync == 0 {
		p.findSync()
		return nil
	}

	if p.indata {
		// data mode, just decode ourselves into
		// a set of sample clusters
		// log.Fatal("in data !")
		log.Println("will decode - length=", p.Length())
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

	p.sdh = SampaDataHeader{BitSet: *(p.BitSet.Last(HeaderSize))}
	// fmt.Println(sdh.StringAnnotated("\n"))
	switch uint(p.sdh.PKT()) {
	case DataPKT:
		fmt.Println("DATA:", p.sdh.StringAnnotated("\n"))
		// log.Fatal("YEP")
		dataToGo := p.sdh.NumWords()
		fmt.Println(">>>", dataToGo, " 10-bits words to read")
		fmt.Println("ELINK:", p)
		p.Clear()
		p.checkpoint = int(dataToGo * 10)
		p.indata = true
		return nil
	case SyncPKT:
		p.nsync++
		log.Println("found sync ", p.nsync)
		p.Clear()
		p.checkpoint = HeaderSize
		return nil
	case HeartBeatPKT:
		log.Println("HEARTBEAT found. Should be do sth about it ?")
		p.Clear()
		p.checkpoint = HeaderSize
		return nil
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

func (p *elink) ForceClear() {
	if p.indata {
		return
	}
	p.nsync = 0
	p.Clear()
	p.indata = false
}

// Decode returns a slice of SAMPA Cluster (cluster in the sense of
// set of ADC values).
func (p *elink) Decode() []Cluster {
	tb := p.Split()
	fmt.Println("Decode : slice size=", len(tb))
	for _, t := range tb {
		fmt.Printf("[%v] ", t)
	}
	fmt.Println()
	// TODO: each cluster must contain Hadd and CHadd from p.Hadd and p.CHadd
	return nil
}

func (p *elink) IsEmpty() bool {
	return p.Length() == 0
}
