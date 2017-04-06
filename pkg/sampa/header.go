package sampa

import (
	"errors"
	"fmt"
	"log"

	"github.com/mrrtf/sampa/pkg/bitset"
)

const (
	HammingFirstBit  int = 0
	HammingLastBit   int = 5
	PBit             int = 6
	PKTFirstBit      int = 7
	PKTLastBit       int = 9
	NumWordsFirstBit int = 10
	NumWordsLastBit  int = 19
	HaddFirstBit     int = 20
	HaddLastBit      int = 23
	CHaddFirstBit    int = 24
	CHaddLastBit     int = 28
	BXcountFirstBit  int = 29
	BXcountLastBit   int = 48
	DPBit            int = 49
)

// SampaDataHeader is the fixed length SAMPA header (50 bits)
//
//  6 bits hamming code
//  1 bit parity (odd) of header including hamming
//  3 bits packet type
// 10 bits number of 10 bit words in data payload
//  4 bits hardware address of chip
//  5 bits channel address
// 20 bits bunch-crossing counter (40MHz counter)
//  1 bit parity (odd) of data payload
type SampaDataHeader struct {
	bitset.BitSet
}

func (sdh *SampaDataHeader) SetHamming(v uint) error {
	if v > (1<<uint(HammingLastBit-HammingFirstBit+1))-1 {
		return errors.New(fmt.Sprintf("Hamming should be %d bits",
			HammingLastBit-HammingFirstBit+1))
	}
	sdh.SetRangeFromUint8(HammingFirstBit, HammingLastBit, uint8(v))
	return nil
}

func (sdh *SampaDataHeader) SetP(v bool) {
	sdh.Set(PBit, v)
}

func (sdh *SampaDataHeader) SetPKT(v uint) error {
	if v > (1<<uint(PKTLastBit-PKTFirstBit+1))-1 {
		return errors.New(fmt.Sprintf("PKT should be %d bits",
			PKTLastBit-PKTFirstBit+1))
	}
	sdh.SetRangeFromUint8(PKTFirstBit, PKTLastBit, uint8(v))
	return nil
}

func (sdh *SampaDataHeader) SetNumWords(v uint) error {
	if v > (1<<uint(NumWordsLastBit-NumWordsFirstBit+1))-1 {
		return errors.New(fmt.Sprintf("Hamming should be %d bits",
			NumWordsLastBit-NumWordsFirstBit+1))
	}
	sdh.SetRangeFromUint16(NumWordsFirstBit, NumWordsLastBit, uint16(v))
	return nil
}

func (sdh *SampaDataHeader) SetHadd(v uint) error {
	if v > (1<<uint(HaddLastBit-HaddFirstBit+1))-1 {
		return errors.New(fmt.Sprintf("Hadd should be %d bits",
			HaddLastBit-HaddFirstBit+1))
	}
	sdh.SetRangeFromUint8(HaddFirstBit, HaddLastBit, uint8(v))
	return nil
}

func (sdh *SampaDataHeader) SetCHadd(v uint) error {
	if v > (1<<uint(CHaddLastBit-CHaddFirstBit+1))-1 {
		return errors.New(fmt.Sprintf("CHadd should be %d bits",
			CHaddLastBit-CHaddFirstBit+1))
	}
	sdh.SetRangeFromUint8(CHaddFirstBit, CHaddLastBit, uint8(v))
	return nil
}

func (sdh *SampaDataHeader) SetBXcount(v uint) error {
	if v > (1<<uint(BXcountLastBit-BXcountFirstBit+1))-1 {
		return errors.New(fmt.Sprintf("BXcount should be %d bits",
			BXcountLastBit-BXcountFirstBit+1))
	}
	sdh.SetRangeFromUint32(BXcountFirstBit, BXcountLastBit, uint32(v))
	return nil
}

func (sdh *SampaDataHeader) SetDP(v bool) {
	sdh.Set(DPBit, v)
}

func (sdh *SampaDataHeader) Hamming() uint8 {
	return sdh.Uint8(HammingFirstBit, HammingLastBit)
}

func (sdh *SampaDataHeader) P() bool {
	return sdh.Get(PBit)
}

func (sdh *SampaDataHeader) PKT() uint8 {
	return sdh.Uint8(PKTFirstBit, PKTLastBit)
}

func (sdh *SampaDataHeader) NumWords() uint16 {
	return sdh.Uint16(NumWordsFirstBit, NumWordsLastBit)
}

func (sdh *SampaDataHeader) Hadd() uint8 {
	return sdh.Uint8(HaddFirstBit, HaddLastBit)
}

func (sdh *SampaDataHeader) CHadd() uint8 {
	return sdh.Uint8(CHaddFirstBit, CHaddLastBit)
}

func (sdh *SampaDataHeader) BXcount() uint32 {
	return sdh.Uint32(BXcountFirstBit, BXcountLastBit)
}

func (sdh *SampaDataHeader) DP() bool {
	return sdh.Get(DPBit)
}

func (sdh *SampaDataHeader) StringAnnotated(sep string) string {
	s := fmt.Sprintf("Hamming %d 0b%b 0x%X", sdh.Hamming(), sdh.Hamming(), sdh.Hamming())
	s += sep
	s += fmt.Sprintf("P        %v", sdh.P())
	s += sep
	s += fmt.Sprintf("PKT      %d 0b%b 0x%X", sdh.PKT(), sdh.PKT(), sdh.PKT())
	s += sep
	s += fmt.Sprintf("NumWords %d 0b%b 0x%X", sdh.NumWords(), sdh.NumWords(), sdh.NumWords())
	s += sep
	s += fmt.Sprintf("Hadd     %d 0b%b 0x%X", sdh.Hadd(), sdh.Hadd(), sdh.Hadd())
	s += sep
	s += fmt.Sprintf("CHadd    %d 0b%b 0x%X", sdh.CHadd(), sdh.CHadd(), sdh.CHadd())
	s += sep
	s += fmt.Sprintf("BXcount  %d 0b%b 0x%X", sdh.BXcount(), sdh.BXcount(), sdh.BXcount())
	s += sep
	s += fmt.Sprintf("DP       %v", sdh.DP())
	return s
}

const (
	HeartBeatPKT                    uint = 0
	DataTruncatedPKT                uint = 1
	SyncPKT                         uint = 2
	DataTruncatedTriggerTooEarlyPKT uint = 3
	DataPKT                         uint = 4
	DataNumWordsPKT                 uint = 5
	DataTriggerTooEarlyPKT          uint = 6
	DataTriggerTooEarlyNumWordsPKT  uint = 7
)

var SyncPattern SampaDataHeader

func init() {
	SyncPattern = SampaDataHeader{*bitset.New(50)}
	SyncPattern.SetHamming(0x13)
	SyncPattern.SetP(false)
	SyncPattern.SetPKT(SyncPKT)
	SyncPattern.SetNumWords(0)
	SyncPattern.SetHadd(0xF)
	SyncPattern.SetCHadd(0)
	SyncPattern.SetBXcount(0xAAAAA)
	SyncPattern.SetDP(false)
	if SyncPattern.Length() != 50 {
		log.Fatal("sync pattern is not 50 bits as expected")
	}
	var sp uint64 = SyncPattern.Uint64(0, -1)
	if sp != 0x1555540F00113 {
		log.Fatal(fmt.Sprintf("SyncPattern expected to be 0x1555540F00113 but is %x", sp))
	}
	fmt.Printf("SYNC is assumed to be %X = %s ; count=%d\n", SyncPattern.Uint64(0, -1),
		SyncPattern.String(), SyncPattern.Count())
}
