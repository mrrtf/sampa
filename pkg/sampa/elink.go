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
	id         int
}

func NewELink(id int) *elink {
	return &elink{id: id, BitSet: *(bitset.New(100000)), checkpoint: HeaderSize, indata: false, nsync: 0}
}

func (p *elink) Id() int {
	return p.id
}

func (p *elink) String() string {
	return fmt.Sprintf("ELink %d len %d checkpoint %d indata %v nsync %d %s", p.id, p.Length(), p.checkpoint, p.indata, p.nsync, p.BitSet.StringLSBRight())
}

// Append adds 1 bit at the end of the bitset.
// If the resulting bitset's length reaches the checkpoint
// then the bitset is further processed by the Process method
func (p *elink) AppendBit(bit bool) (*Packet, error) {
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
func (p *elink) Append(bit0, bit1 bool) (*Packet, error) {
	packet0, err := p.AppendBit(bit0)
	if err != nil {
		return nil, err
	}
	packet1, err := p.AppendBit(bit1)
	if err != nil {
		return nil, err
	}

	if packet1 != nil && packet0 != nil {
		log.Fatal("only one of the two AppendBit method should have returned a packet!!!")
	}
	if packet1 != nil {
		return packet1, nil
	}
	return packet0, nil
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

	log.Println("findSync: found sync #", p.nsync, " for elink #", p.id)
	p.Clear()
	p.checkpoint = HeaderSize
	p.nsync++
}

// Process attempts to interpret the current bitset
// as either a Sampa header or Sampa data
// If it's neither, then set the checkpoint at
// the current length + 2 bits
func (p *elink) Process() *Packet {
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
		// a set of sampa packets
		packet := p.GetPacket()
		p.Clear()
		p.checkpoint = HeaderSize
		p.indata = false
		return &packet
	}

	// looking for a header
	if p.checkpoint != HeaderSize {
		panic(fmt.Sprintf("wrong logic 5 checkpoint %d HeaderSize %d", p.checkpoint, HeaderSize))
	}

	p.sdh = SampaDataHeader{BitSet: *(p.BitSet.Last(HeaderSize))}
	// fmt.Println("ELink ", p.id, " ", p.sdh.StringAnnotated("-"))
	switch uint(p.sdh.PKT()) {
	case DataTruncatedPKT, DataTruncatedTriggerTooEarlyPKT, DataTriggerTooEarlyPKT, DataTriggerTooEarlyNumWordsPKT:
		// data with a problem is still data, i.e. there will
		// probably be some data words to read in
		fallthrough
	case DataPKT:
		// log.Println("ELink", p.id, "DATA", p.sdh.StringAnnotated(" "))
		// log.Println(p)
		dataToGo := p.sdh.NumWords()
		p.Clear()
		p.checkpoint = int(dataToGo * 10)
		p.indata = true
		return nil
	case SyncPKT:
		p.nsync++
		p.Clear()
		p.checkpoint = HeaderSize
		return nil
	case HeartBeatPKT:
		log.Printf("ELink #%d : HEARTBEAT found. Should be do sth about it  ?\n", p.id)
		log.Println(p)
		p.Clear()
		p.checkpoint = HeaderSize
		return nil
	default:
		log.Printf("ELink %d Got a PKT=%d\n", p.id, p.sdh.PKT())
		log.Println(p)
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

// Decode returns a SAMPA Packet
func (p *elink) GetPacket() Packet {
	// log.Printf("ELink %d PKT %d GetPacket\n", p.id, p.sdh.PKT())
	// log.Println(p.sdh.StringAnnotated(" "))
	tb := p.Split()
	i := 0
	packet := Packet{sdh: p.sdh, elink: p.id}
	// if data is truncated, do not even try to add anything
	// to the packet
	if uint(p.sdh.PKT()) != DataPKT {
		return packet
	}
	for i < len(tb) {
		nwords := tb[i]
		timestamp := tb[i+1]
		packet.AddCluster(timestamp, tb[i+2:i+2+nwords])
		i += nwords + 2
	}
	return packet

}

func (p *elink) IsEmpty() bool {
	return p.Length() == 0
}
