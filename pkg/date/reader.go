package date

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

var (
	ErrEmptyEvent = errors.New("date: empty event")
	ErrInvalidSOP = errors.New("date: invalid start of packet")
	ErrEndOfEvent = errors.New("date: end of event")
)

const (
	magic            uint32 = 0xDA1E5AFE
	nDateWordsPerGBT        = 4 // 4 x 32 bits words
	nDateBytesPerGBT        = 4 * nDateWordsPerGBT
)

// DateReader is meant to read GBT words from a DATE
// file.
type DateReader struct {
	r     io.Reader
	event *EventType
	pos   int
	// gbtword gbt.Word
	gbt     []byte
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
	return &DateReader{r: bufio.NewReader(file), event: NewEvent(), pos: -1, gbt: make([]byte, 10), headBuf: make([]byte, headerSize), nevents: 0, ngbt: 0}
}

var gbtcount int = 0

// Read reads 10 bytes from the underlying stream
// FIXME: this is not really satisfying the Read interface
// (e.g. we really expect p to be of len 10, nothing else...)
func (dr *DateReader) Read(p []byte) (n int, err error) {
	if len(p) != 10 {
		log.Fatal("DateReader.Read method is so far only able to deal with 10 bytes slices")
	}
	for {
		err = dr.NextGBT()
		if err == io.EOF {
			return 0, err
		}
		if err == nil {
			copy(p, dr.gbt)
			gbtcount++
			// fmt.Printf("GBT word %d = %s\n", gbtcount, dr.GBTAsString())
			return len(dr.gbt), nil
		}
		if err == ErrEndOfEvent {
			gbtcount = 0
			return 0, err
		}
	}
}

// NextGBT advances to the next 10 bytes representing a single 80-bit GBT word
// and fills the gbt internal slice with those
//
// Note that the DATE events without payload or with incorrect start
// of (sampa) packet are simply skipped by NextGBT
func (dr *DateReader) NextGBT() (err error) {

	if dr.pos < 0 {
		err = dr.GetNextEvent()
		if err != nil {
			return err
		}
		// log.Println(dr)
		if !dr.event.HasPayload() {
			// skip to next event
			log.Println("Event without payload. Skipping")
			return ErrEmptyEvent
		}
		_, err := dr.event.SOP()
		if err != nil {
			// invalid SOP, skip to next event
			log.Println("Event with invalid SOP. Skipping")
			return ErrInvalidSOP
		}
		dr.pos = 0
		// log.Printf("SOE %d len of data %d size %d", dr.event.Header().EventID,
		// 	len(dr.event.Data()), dr.event.size)
	}

	endOfEvent := dr.pos+nDateWordsPerGBT >= dr.event.size-equipmentHeaderSize-2*nDateBytesPerGBT

	if endOfEvent {
		// log.Println("EOE reached. Going to next event")
		dr.pos = -1
		return ErrEndOfEvent
	}

	dr.Data2GBT(dr.pos)
	dr.pos += nDateBytesPerGBT

	if len(dr.gbt) != 10 {
		log.Fatal("dr.gbt is not of size 10")
	}
	return nil
}

func (dr *DateReader) Data2GBTHelper(gbt []byte, data []byte) {
	gbt[0] = data[8+4]
	gbt[1] = data[9+4]
	gbt[2] = data[10+4]
	gbt[3] = data[11+4]
	gbt[4] = data[4+4]
	gbt[5] = data[5+4]
	gbt[6] = data[6+4]
	gbt[7] = data[7+4]
	gbt[8] = data[0+4]
	gbt[9] = data[1+4]
	// gbt[0] = data[8]
	// gbt[1] = data[9]
	// gbt[2] = data[10]
	// gbt[3] = data[11]
	// gbt[4] = data[4]
	// gbt[5] = data[5]
	// gbt[6] = data[6]
	// gbt[7] = data[7]
	// gbt[8] = data[0]
	// gbt[9] = data[1]
}

// Data2GBT converts 3 32-bits (DATE) words into a 80-bits GBT word
// represented by a slice of 10 bytes
func (dr *DateReader) Data2GBT(pos int) {
	data := dr.event.Data()[pos : pos+nDateBytesPerGBT]
	dr.Data2GBTHelper(dr.gbt, data)
	dr.ngbt++
}

func (dr *DateReader) NofEvents() int {
	return dr.nevents
}

func (dr *DateReader) NofGBTwords() int {
	return dr.ngbt
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

// insure we only have one equipment,
// as this is the only thing we can deal with so far
func (dr *DateReader) mustHaveOnlyOneEquipment() {
	eqSize := binary.LittleEndian.Uint32(dr.event.payload[:4])
	if eqSize != uint32(dr.event.size) {
		log.Fatal(eqSize, "!=", uint32(dr.event.size)+headerSize)
	}
}

// GetNextEvent gets the next DATE event found
func (dr *DateReader) GetNextEvent() (err error) {
	n, err := dr.r.Read(dr.headBuf)
	if err != nil {
		return err
	}
	if n != int(headerSize) {
		log.Fatalf("Read %d bytes and not %d as expected", n, headerSize)
	}

	dr.nevents++

	dr.mustFillHeader()
	dr.event.OnlyHeader(dr.header)

	if dr.header.EventSize <= headerSize {
		// emty event, we skip it
		return ErrEmptyEvent
	}

	ndatabytes := int(dr.header.EventSize - headerSize)
	n, err = io.ReadFull(dr.r, dr.event.payload[:ndatabytes])

	if n != ndatabytes {
		log.Println(err)
		log.Fatalf("Could only read %d out of %d bytes expected", n, ndatabytes)
	}
	if err != nil {
		dr.pos = -1
		return err
	}

	dr.event.size = ndatabytes

	dr.mustHaveOnlyOneEquipment()

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

func (dr *DateReader) GBTAsString() string {
	v := ""
	for b := 0; b < len(dr.gbt); b++ {
		v += fmt.Sprintf("%0x ", dr.gbt[b])
	}
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
