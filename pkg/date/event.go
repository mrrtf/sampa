package date

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"

	"github.com/fatih/color"
)

const (
	maxPayloadSize = 1024 * 1024
)

var yellow = color.New(color.FgYellow).SprintFunc()
var blue = color.New(color.FgBlue).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()

type EventHeaderType struct {
	EventSize         uint32    //  [ 0: 4[
	EventMagic        uint32    //    4: 8
	HeaderSize        uint32    //    8:12
	Version           uint32    //   12:16
	EventType         uint32    //   16:20
	RunNumber         uint32    //   20:24
	EventID           uint64    //   24:32
	Trigger           [2]uint64 //   32:48
	Detectors         uint32    //   48:52
	Attributes        [3]uint32 //   52:64
	Ldc               uint32    //   64:68
	Gdc               uint32    //   68:72
	TimeStampSec      uint32    //   72:76
	TimeStampMicroSec uint32    //   76:80
}

type EquipmentHeaderType struct {
	Size       uint32    // [ 0: 4[
	Type       uint32    //   4: 8
	Id         uint32    //   8:12
	Attributes [3]uint32 //  12:24
	ElemSize   uint32    //  24:28
}

var headerSize uint32
var equipmentHeaderSize int

func init() {
	var h EventHeaderType
	var he EquipmentHeaderType
	headerSize = uint32(binary.Size(h))
	equipmentHeaderSize = binary.Size(he)
	if equipmentHeaderSize != 28 {
		log.Fatal("oups equipmentHeaderSize=", equipmentHeaderSize)
	}
}

// EventType is a simple DATE event = header + payload
// TODO: should be more = header + { equipmentHeader,payload }
type EventType struct {
	header  EventHeaderType
	payload []byte
	size    int // used size of payload
}

func NewEvent() *EventType {
	return &EventType{
		header:  EventHeaderType{},
		payload: make([]byte, maxPayloadSize),
		size:    0}
}

func (event *EventType) Header() EventHeaderType {
	return event.header
}

func (event *EventType) OnlyHeader(header EventHeaderType) {
	event.header = header
	event.size = 0
}

func (event *EventType) Data() []byte {
	if !event.HasPayload() {
		return nil
	}
	return event.payload[equipmentHeaderSize+nDateBytesPerGBT:]
}

func (event *EventType) HasPayload() bool {
	return event.header.EventSize > headerSize &&
		(len(event.payload) > equipmentHeaderSize+nDateBytesPerGBT)
}

// Quartets converts the bytes starting at payload[pos]
// into 4 32-bits values
func (event *EventType) quartet(pos int) (uint32, uint32, uint32, uint32) {
	x := event.payload[pos : pos+16]
	return binary.LittleEndian.Uint32(x[0:4]), binary.LittleEndian.Uint32(x[4:8]), binary.LittleEndian.Uint32(x[8:12]), binary.LittleEndian.Uint32(x[12:16])
}

// start of packet (SOP = 0x000000000000000000000000000001)
func (event *EventType) SOP() ([]byte, error) {
	if !event.HasPayload() {
		return nil, nil
	}
	a, b, c, d := event.quartet(28)
	if a != 0 || b != 0 || c != 0 || d != 1 {
		return event.payload[28:44], errors.New(fmt.Sprintf("unexpected sop %08X %08X %08X %08X", a, b, c, d))
	}
	return event.payload[28:44], nil
}

func (h EventHeaderType) String() string {

	v := fmt.Sprintf("\n%s ", blue("eveSize "))
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

func StringPerLine(buf []byte, perline int) string {
	v := ""
	offset := 0
	m := len(buf)
	for offset < m {
		for b := 0; b < perline && offset < m; b++ {
			v += fmt.Sprintf("%02X%02X%02X%02X ", buf[offset+3],
				buf[offset+2], buf[offset+1], buf[offset])
			offset += 4
		}
		v += "\n"
	}
	return v
}

var nbadsop = 0

func (event *EventType) String() string {
	v := event.header.String()
	v += "\n---\n"

	if event.HasPayload() {
		sop, err := event.SOP()
		v += blue("SOP ") + StringPerLine(sop, 4)
		if err != nil {
			nbadsop++
			if nbadsop > 10 {
				log.Fatal(err)
			}
		}
		// v += blue("DATA\n") + StringPerLine(event.Data()[:16*5], 4)
	}
	return v
}
