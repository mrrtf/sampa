package edm

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
)

type EventHeaderType struct {
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

type EventDataType []uint32

// EventType is a simple DATE event = header + payload
type EventType struct {
	header  EventHeaderType
	payload EventDataType
}

func (event EventType) Data() EventDataType {
	if !event.HasPayload() {
		return nil
	}
	return event.payload[10:]
}

func (event EventType) HasPayload() bool {
	return event.header.EventSize > uint32(binary.Size(event.header)) &&
		len(event.payload) > 10
}

// start of packet (SOP = 0 0 0x1)
func (event EventType) SOP() (EventDataType, error) {
	var s EventDataType
	if !event.HasPayload() {
		return nil, nil
	}
	s = event.payload[7:10]
	asExpected := s[0] == 0 &&
		s[1] == 0 &&
		s[2] == 1
	if !asExpected {
		return s, errors.New("unexpected sop")
	}
	return s, nil
}

func (h EventHeaderType) String() string {

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

func (buf EventDataType) String(perline int) string {
	v := ""
	offset := 0
	m := len(buf)
	for offset < m {
		for b := 0; b < perline && offset < m; b++ {
			v += fmt.Sprintf("%08X ", buf[offset])
			offset++
		}
		v += "\n"
	}
	return v
}

func (event EventType) String() string {
	v := event.header.String()
	v += "\n---\n"

	if event.HasPayload() {
		v += blue("payload  ") + event.payload[0:7].String(7)
		v += "***\n"
		sop, err := event.SOP()
		if err == nil && sop != nil {
			v += blue("sop      " + sop.String(3))
			size := len(event.payload[10:]) / 3
			if size != 8192 {
				// this test is probably valid only for the SOLAR tests
				log.Printf(red("Was expecting %d bytes, got %d"), 8192, size)
			}
		} else if sop != nil {
			v += red("sop      " + sop.String(3))
			v += red("extra\n" + event.payload[10:20].String(5))
		}
	}
	return v
}

// GetEvent returns the next DATE event found in reader
func GetEvent(r io.Reader) (event EventType, err error) {
	const magic uint32 = 0xDA1E5AFE
	var header EventHeaderType
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

	buf := make([]uint32, (int64(header.EventSize)-headerSize)/4)

	err = binary.Read(r, binary.LittleEndian, buf)

	if err != nil {
		return EventType{}, err
	}

	return EventType{header: header, payload: buf}, nil
}
