package date

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aphecetche/sampa/pkg/gbt"
)

var (
	ErrEmptyEvent = errors.New("date: empty event")
	ErrInvalidSOP = errors.New("date: invalid start of packet")
	ErrEndOfEvent = errors.New("date: end of event")
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
	event   *EventType
	pos     int
	gbtword gbt.Word
	headBuf []byte
	header  EventHeaderType
	nevents int
	ngbt    int
}

// NewReader returns a DateReader object ready
// to read from the given filename
func NewReader(filename string) *DateReader {
	file, err := os.Open(filename)
	if err != nil {
		log.Println(err)
		return nil
	}
	return &DateReader{r: bufio.NewReaderSize(file, readingBufferSize), event: NewEvent(), pos: -1, gbtword: *gbt.NewWord(), headBuf: make([]byte, headerSize), nevents: 0, ngbt: 0}
	// 	return &DateReader{r: file, event: NewEvent(), pos: -1, gbtword: *gbt.NewWord(), headBuf: make([]byte, headerSize), nevents: 0, ngbt: 0}
}

func (dr *DateReader) NofEvents() int {
	return dr.nevents
}

// GBT returns a single 80-bit GBT word (and nil) or an
// empty word and an error (e.g. to signal EOF or any other
// reason)
// The DATE events without payload or with incorrect start
// of (sampa) packet are simply skipped
func (dr *DateReader) GBT() (g gbt.Word, err error) {

	if dr.pos < 0 {
		err = dr.GetNextEvent()
		if err != nil {
			return gbt.Word{}, err
		}
		dr.nevents++
		// log.Println(dr)
		if !dr.event.HasPayload() {
			// skip to next event
			// log.Println("Event without payload. Skipping")
			return gbt.Word{}, ErrEmptyEvent
		}
		_, err := dr.event.SOP()
		if err != nil {
			// invalid SOP, skip to next event
			// log.Println("Event with invalid SOP. Skipping")
			return gbt.Word{}, ErrInvalidSOP
		}
		dr.pos = 0
		dr.ngbt++
		// log.Printf("SOE %d len of data %d size %d", dr.event.Header().EventID,
		// 	len(dr.event.Data()), dr.event.size)
	}
	if dr.pos+3 >= dr.event.size {
		// log.Println("EOE reached. Going to next event")
		dr.pos = -1
		return gbt.Word{}, ErrEndOfEvent
	}

	dr.Data2GBT(dr.pos)
	// log.Println("GBT from data=", dr.gbtword.StringLSBRight())
	dr.pos += 3 * 4
	return dr.gbtword, nil
}

func Data2GBT(data []byte, g *gbt.Word) {
	d := []byte{
		data[8],
		data[9],
		data[10],
		data[11],
		data[4],
		data[5],
		data[6],
		data[7],
		data[0],
		data[1]}
	g.SetFromBytes(d)
}

// GBTword convert 3 32-bits (DATE) words into a 80-bits bitset
func (dr *DateReader) Data2GBT(pos int) {
	Data2GBT(dr.event.Data()[pos:pos+12], &dr.gbtword)
	dr.ngbt++
}

func (dr *DateReader) mustFillHeader() {
	// this ain't pretty but is (much) faster than
	// using the binary.Read on the header struct itself...
	dr.header.EventSize = binary.LittleEndian.Uint32(dr.headBuf[:4])
	dr.header.EventMagic = binary.LittleEndian.Uint32(dr.headBuf[4:8])
	dr.header.HeaderSize = binary.LittleEndian.Uint32(dr.headBuf[8:12])
	dr.header.Version = binary.LittleEndian.Uint32(dr.headBuf[12:16])
	dr.header.EventType = binary.LittleEndian.Uint32(dr.headBuf[16:20])
	dr.header.RunNumber = binary.LittleEndian.Uint32(dr.headBuf[20:24])
	dr.header.EventID = binary.LittleEndian.Uint64(dr.headBuf[24:32])
	dr.header.Trigger[0] = binary.LittleEndian.Uint64(dr.headBuf[32:40])
	dr.header.Trigger[1] = binary.LittleEndian.Uint64(dr.headBuf[40:48])
	dr.header.Detectors = binary.LittleEndian.Uint32(dr.headBuf[48:52])
	dr.header.Attributes[0] = binary.LittleEndian.Uint32(dr.headBuf[52:56])
	dr.header.Attributes[1] = binary.LittleEndian.Uint32(dr.headBuf[56:60])
	dr.header.Attributes[2] = binary.LittleEndian.Uint32(dr.headBuf[60:64])
	dr.header.Ldc = binary.LittleEndian.Uint32(dr.headBuf[64:68])
	dr.header.Gdc = binary.LittleEndian.Uint32(dr.headBuf[68:72])
	dr.header.TimeStampSec = binary.LittleEndian.Uint32(dr.headBuf[72:76])
	dr.header.TimeStampMicroSec = binary.LittleEndian.Uint32(dr.headBuf[76:80])
	if dr.header.EventMagic != magic {
		log.Fatalf("no magic word (%X) where I expected it, found %X instead. b=%v", magic, dr.header.EventMagic, dr.headBuf)
	}
}

// GetNextEvent returns the next DATE event found
func (dr *DateReader) GetNextEvent() (err error) {
	n, err := dr.r.Read(dr.headBuf)
	if err != nil {
		return err
	}
	if n != int(headerSize) {
		log.Fatalf("Read %d bytes and not %d as expected", n, headerSize)
	}

	dr.mustFillHeader()
	dr.event.OnlyHeader(dr.header)

	if dr.header.EventSize <= headerSize {
		// emty event, we skip it
		return ErrEmptyEvent
	}

	ndatabytes := int(dr.header.EventSize - headerSize)
	n, err = dr.r.Read(dr.event.payload[:ndatabytes])

	if n != ndatabytes {
		log.Fatalf("Could only read %d out of %d bytes expected", n, ndatabytes)
	}
	if err != nil {
		dr.pos = -1
		return err
	}

	dr.event.size = ndatabytes

	return nil
}

func (dr *DateReader) String() string {
	v := "\n---------------------"
	v += fmt.Sprintf("Read %d events", dr.NofEvents())
	v += fmt.Sprintf(" and %d GBT words. Pos %d\n", dr.ngbt, dr.pos)
	v += fmt.Sprintf("Last known event is :")
	v += fmt.Sprintf(dr.event.String())
	return v
}

// func Data2GBTbis(data []byte, g *gbt.Word) {
// 	b2b := [12]byte{64, 72, 0xFF, 0xFF, 32, 40, 48, 56, 0, 8, 16, 24}
// 	if g.Size() != 80 {
// 		log.Fatal("GBT word should be 80 bits")
// 	}
// 	if len(data) != 12 {
// 		log.Fatal("buf should be 12 bytes exactly")
// 	}
// 	for b := 0; b < len(b2b); b++ {
// 		bit := int(b2b[b])
// 		if bit < g.Size() {
// 			g.SetRangeFromUint8(bit, bit+7, data[b])
// 		}
// 	}
// }
