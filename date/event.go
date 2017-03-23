package date

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"log"

	"github.com/fatih/color"
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

var headerSize uint32

func init() {
	var h EventHeaderType
	headerSize = uint32(binary.Size(h))
}

// EventType is a simple DATE event = header + payload
type EventType struct {
	header  EventHeaderType
	payload []byte
	size    int // used size of payload
}

func (event EventType) Header() EventHeaderType {
	return event.header
}

func (event EventType) Data32() EventDataType {
	if !event.HasPayload() {
		return nil
	}
	return event.payload32[10:]
}

func (event EventType) Data() []byte {
	if !event.HasPayload() {
		return nil
	}
	return event.payload[40:]
}

func (event EventType) HasPayload32() bool {
	return event.header.EventSize > headerSize &&
		(len(event.payload32) > 10)
}

func (event EventType) HasPayload() bool {
	return event.header.EventSize > headerSize &&
		(len(event.payload) > 40)
}

// start of packet (SOP = 0 0 0x1)
func (event EventType) SOP32() (EventDataType, error) {
	var s EventDataType
	if !event.HasPayload() {
		return nil, nil
	}
	s = event.payload32[7:10]
	asExpected := s[0] == 0 &&
		s[1] == 0 &&
		s[2] == 1
	if !asExpected {
		return s, errors.New("unexpected sop")
	}
	return s, nil
}

// Triplets converts the bytes starting at payload[pos]
// into 3 32-bits values
func (event EventType) triplet(pos int) (uint32, uint32, uint32) {
	x := event.Data()[pos : pos+12]
	return binary.LittleEndian.Uint32(x[0:4]), binary.LittleEndian.Uint32(x[4:8]), binary.LittleEndian.Uint32(x[4:12])
}

// start of packet (SOP = 0x000000000000000000000001)
func (event EventType) SOP() ([]byte, error) {
	if !event.HasPayload() {
		return nil, nil
	}
	// a,b,c := event.triplet(28)
	s := event.payload[28:40]

	asExpected := (s[3] == 1)

	for i := 0; i < 12 && i != 3 && asExpected; i++ {
		if s[i] != 0 {
			asExpected = false
		}
	}
	if !asExpected {
		hex.Dump(s)
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

func (event EventType) String() string {
	v := event.header.String()
	v += "\n---\n"

	if event.HasPayload32() {
		v += blue("payload  ") + event.payload32[0:7].String(7)
		v += "***\n"
		sop32, err := event.SOP32()
		if err == nil && sop32 != nil {
			v += blue("sop      " + sop32.String(3))
			size := len(event.payload32[10:]) / 3
			if size != 8192 {
				// this test is probably valid only for the SOLAR tests
				log.Printf(red("Was expecting %d bytes, got %d"), 8192, size)
			}
		} else if sop32 != nil {
			v += red("sop      " + sop32.String(3))
			v += red("extra\n" + event.payload32[10:20].String(5))
		}
	}

	if event.HasPayload() {
		sop, err := event.SOP()
		a, b, c := event.triplet(0)
		v += "data=\n"
		v += red(StringPerLine(event.Data()[0:80], 5))
		v += yellow(fmt.Sprintf("%08X %08X %08X\n", a, b, c))
		v += hex.EncodeToString(sop)
		if err != nil {
			nbadsop++
			if nbadsop > 3 {
				log.Fatal(err)
			}
		}
	}
	return v
}
