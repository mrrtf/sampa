package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/aphecetche/bitset"
)

type eventHeaderType struct {
	EventSize         uint32
	EventMagic        uint32
	HeaderSize        uint32
	Version           uint32
	EventType         uint32
	RunNumber         uint32
	EventID           uint64
	Trigger           [2]uint64
	Detectors         uint32
	Attributes        [3]uint32
	Ldc               uint32
	Gdc               uint32
	TimeStampSec      uint32
	TimeStampMicroSec uint32
}

type eventDataType []byte

// EventType is a simple DATE event = header + payload
type EventType struct {
	header  eventHeaderType
	payload eventDataType
}

func (event EventType) data() eventDataType {
	if !event.hasPayload() {
		return nil
	}
	return event.payload[40:]
}

func (event EventType) hasPayload() bool {
	return event.header.EventSize > uint32(binary.Size(event.header)) &&
		len(event.payload) > 40
}

// start of packet (SOP = 0 0 0x1)
func (event EventType) sop() (eventDataType, error) {
	var s eventDataType
	if !event.hasPayload() {
		return nil, nil
	}
	s = event.payload[28:40]
	asExpected := binary.LittleEndian.Uint32(s[:4]) == 0 &&
		binary.LittleEndian.Uint32(s[4:8]) == 0 &&
		binary.LittleEndian.Uint32(s[8:]) == 1
	if !asExpected {
		return s, errors.New("unexpected sop")
	}
	return s, nil
}

func (h eventHeaderType) String() string {

	v := fmt.Sprintf("%s ", blue("eveSize "))
	v += fmt.Sprintf("%08X", h.EventSize)
	v += fmt.Sprintf(" %s ", blue("magic   "))
	v += fmt.Sprintf("%08X", h.EventMagic)
	v += fmt.Sprintf(" %s ", blue("headSize"))
	v += fmt.Sprintf("%08X", h.HeaderSize)
	v += fmt.Sprintf(" %s ", blue("version "))
	v += fmt.Sprintf("%08X\n", h.Version)

	v += fmt.Sprintf("%s ", blue("eveType "))
	v += fmt.Sprintf("%08X", h.EventType)
	v += fmt.Sprintf(" %s ", blue("run     "))
	v += fmt.Sprintf("%08X", h.RunNumber)
	v += fmt.Sprintf(" %s ", blue("id      "))
	v += fmt.Sprintf("%016X\n", h.EventID)

	v += fmt.Sprintf("%s ", blue("trigger "))
	v += fmt.Sprintf("%016X%016X\n", h.Trigger[0], h.Trigger[1])

	v += fmt.Sprintf("%s ", blue("dets    "))
	v += fmt.Sprintf("%08X", h.Detectors)
	v += fmt.Sprintf(" %s ", blue("attr    "))
	v += fmt.Sprintf("%08X%08X%08X\n", h.Attributes[0],
		h.Attributes[1], h.Attributes[2])

	v += fmt.Sprintf("%s ", blue("LDC     "))
	v += fmt.Sprintf("%08X", h.Ldc)
	v += fmt.Sprintf(" %s ", blue("GDC     "))
	v += fmt.Sprintf("%08X", h.Gdc)
	v += fmt.Sprintf(" %s ", blue("time(s)"))
	v += fmt.Sprintf("%08X", h.TimeStampSec)
	v += fmt.Sprintf(" %s ", blue("time(us)"))
	v += fmt.Sprintf("%08X", h.TimeStampMicroSec)
	return v
}

func (buf eventDataType) String(perline int) string {
	v := ""
	offset := 0
	m := len(buf)
	for offset < m {
		for b := 0; b < perline && offset < m; b++ {
			v += fmt.Sprintf("%02X%02X%02X%02X ", buf[offset+3], buf[offset+2], buf[offset+1], buf[offset])
			offset += 4
		}
		v += "\n"
	}
	return v
}

func (event EventType) String() string {
	v := event.header.String()
	v += "\n---\n"
	if event.hasPayload() {
		v += blue("payload  ") + event.payload[0:28].String(7)
		v += "***\n"
		sop, err := event.sop()
		if err == nil && sop != nil {
			v += blue("sop      " + sop.String(3))
			size := len(event.payload[40:]) / 3 / 4
			if size != 8192 {
				// this test is probably valid only for the SOLAR tests
				log.Printf(red("Was expecting %d bytes, got %d"), 8192, size)
			}
		} else if sop != nil {
			v += red("sop      " + sop.String(3))
			v += red("extra\n" + event.payload[40:80].String(5))
		}
	}
	return v
}

// GetEvent returns the next DATE event found in reader
func getEvent(r io.Reader) (event EventType, err error) {
	const magic uint32 = 0xDA1E5AFE
	var header eventHeaderType
	var headerSize = int64(binary.Size(header))
	err = binary.Read(r, binary.LittleEndian, &header)
	if err != nil {
		return EventType{}, err
	}

	if header.EventMagic != magic {
		log.Fatal("not a magic word where I expected it")
	}

	if int64(header.EventSize) <= headerSize {

		return EventType{header: header, payload: nil}, nil
	}

	buf := make([]byte, int64(header.EventSize)-headerSize)

	n, err := r.Read(buf)

	if err != nil {
		return EventType{}, err
	}

	if n != len(buf) {
		log.Fatal("Could not read full event")
	}
	return EventType{header: header, payload: buf}, nil
}

var bs0 = bitset.NewBitSet(1024)

func GBTword(data eventDataType) (*bitset.BitSet, error) {

	if len(data) != 12 {
		return nil, errors.New("3 32-bits words expected")
	}
	bs := bitset.NewBitSet(80)

	return bs, nil
}

func processEvent(event EventType) {
	fmt.Println(event)

	if !event.hasPayload() {
		return
	}

	// sop, err := event.sop()
	_, err := event.sop()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("would treat the data here")
	n := 5

	fmt.Println(event.data()[:72].String(3))

	for i := 0; i < n; i++ {

		bs, err := GBTword(event.data()[i : i+12])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(bs)

	}

	fmt.Println()
}
