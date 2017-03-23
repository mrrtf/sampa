package date

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aphecetche/sampa/gbt"
)

const (
	magic              uint32 = 0xDA1E5AFE
	readingBufferSize         = 1024 * 1024 * 1024
	maxEventBufferSize        = 1024 * 1024
)

// DateReader is meant to read GBT words from a DATE
// file. It is implementing the gbt.GBT interface
type DateReader struct {
	r       io.Reader
	event   EventType
	pos     int
	gbtword gbt.Word
	buf     []uint32
}

// NewReader returns a DateReader object ready
// to read from the given filename
func NewReader(filename string) *DateReader {
	file, err := os.Open(filename)
	if err != nil {
		log.Println(err)
		return nil
	}
	return &DateReader{r: bufio.NewReaderSize(file, readingBufferSize), event: EventType{}, pos: -1, gbtword: *gbt.NewWord()}
}

// GBT returns a single 80-bit GBT word (and nil) or an
// empty word and an error (e.g. to signal EOF or any other
// reason)
// The DATE events without payload or with incorrect start
// of (sampa) packet are simply skipped
func (dr *DateReader) GBT() (g gbt.Word, err error) {
	if dr.pos < 0 {
		err = dr.GetNextEvent2()
		if err != nil {
			log.Println("ERROR:", err)
			return gbt.Word{}, err
		}
		fmt.Println(dr.event.String())
		if !dr.event.HasPayload() {
			// skip to next event
			log.Println("Event without payload. Skipping")
			return dr.GBT()
		}
		_, err := dr.event.SOP()
		if err != nil {
			// invalid SOP, skip to next event
			log.Println("Event with invalid SOP. Skipping")
			return dr.GBT()
		}
		dr.pos = 0
		log.Printf("SOE %d", dr.event.Header().EventID)
	}
	if dr.pos+3 >= len(dr.event.Data()) {
		log.Println("EOE reached. Going to next event")
		dr.pos = -1
		return dr.GBT()
	}

	dr.Data2GBT(dr.pos)
	dr.pos += 3
	return dr.gbtword, nil
}

// GBTword convert 3 32-bits (DATE) words into a 80-bits bitset
func (dr *DateReader) Data2GBT(pos int) {
	// data := dr.event.Data()[pos : pos+3]
	_ = dr.event.Data()[pos : pos+3]
	// dr.gbtword.SetRangeFromUint32Fast(64, 79, data[0])
	// dr.gbtword.SetRangeFromUint32Fast(32, 64, data[1])
	// dr.gbtword.SetRangeFromUint32Fast(0, 31, data[2])
}

var header EventHeaderType

func (dr *DateReader) GetNextEvent() (err error) {
	err = binary.Read(dr.r, binary.LittleEndian, &header)
	if err != nil {
		dr.event = nil
		return err
	}
	if header.EventMagic != magic {
		log.Fatal("not a magic word where I expected it")
	}
	if header.EventSize <= headerSize {
		dr.event = &EventType{header: header, payload: nil}
		return nil
	}
	err = binary.Read(dr.r, binary.LittleEndian, eb32[:int(header.EventSize-headerSize)/4])
	if err != nil {
		dr.event = nil
		return err
	}
	dr.event = &EventType{header: header, payload32: eb32}
	return nil
}

var hb []byte
var eb []byte
var eb32 []uint32

func init() {
	hb = make([]byte, headerSize) // 2 32-bits words
	eb = make([]byte, maxEventBufferSize)
	eb32 = make([]uint32, maxEventBufferSize/4)
}

// GetNextEvent2 returns the next DATE event found
func (dr *DateReader) GetNextEvent2() (err error) {
	n, err := dr.r.Read(hb)
	if err != nil {
		return err
	}
	if n != int(headerSize) {
		log.Fatalf("Read %d bytes and not %d as expected", n, headerSize)
	}

	// this ain't pretty but is much faster than
	// using the binary.Read on the header struct itself...
	header.EventSize = binary.LittleEndian.Uint32(hb[:4])
	header.EventMagic = binary.LittleEndian.Uint32(hb[4:8])
	header.HeaderSize = binary.LittleEndian.Uint32(hb[8:12])
	header.Version = binary.LittleEndian.Uint32(hb[12:16])
	header.EventType = binary.LittleEndian.Uint32(hb[16:20])
	header.RunNumber = binary.LittleEndian.Uint32(hb[20:24])
	header.EventID = binary.LittleEndian.Uint64(hb[24:32])
	header.Trigger[0] = binary.LittleEndian.Uint64(hb[32:40])
	header.Trigger[1] = binary.LittleEndian.Uint64(hb[40:48])
	header.Detectors = binary.LittleEndian.Uint32(hb[48:52])
	header.Attributes[0] = binary.LittleEndian.Uint32(hb[52:56])
	header.Attributes[1] = binary.LittleEndian.Uint32(hb[56:60])
	header.Attributes[2] = binary.LittleEndian.Uint32(hb[60:64])
	header.Ldc = binary.LittleEndian.Uint32(hb[64:68])
	header.Gdc = binary.LittleEndian.Uint32(hb[68:72])
	header.TimeStampSec = binary.LittleEndian.Uint32(hb[72:76])
	header.TimeStampMicroSec = binary.LittleEndian.Uint32(hb[76:80])

	if header.EventMagic != magic {
		log.Fatalf("no magic word (%X) where I expected it, found %X instead. b=%v", magic, header.EventMagic, hb)
	}
	if header.EventSize <= headerSize {
		// emty event, we skip it
		dr.event = &EventType{header: header, payload: nil}
		return nil
	}

	ndatabytes := int(header.EventSize - headerSize)
	n, err = dr.r.Read(eb[:ndatabytes])

	if n != ndatabytes {
		log.Fatalf("Could only read %d out of %d bytes expected", n, ndatabytes)
	}
	if err != nil {
		dr.event = nil
		return err
	}

	// s := blue(StringPerLine(eb[:80], 7))
	// fmt.Println(s)

	dr.event = &EventType{header: header, payload32: nil, payload: make([]byte, ndatabytes)}

	copy(dr.event.payload, eb[:ndatabytes])

	return nil
}
