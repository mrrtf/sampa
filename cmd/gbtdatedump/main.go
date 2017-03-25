package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime/pprof"

	"github.com/aphecetche/sampa/pkg/bitset"
	"github.com/aphecetche/sampa/pkg/date"
	"github.com/aphecetche/sampa/pkg/sampa"
	"github.com/fatih/color"
)

var flagMaxGBTwords int
var flagNoColor bool
var flagCpuProfile string
var flagMemProfile string
var flagMaxEvents int

var NumberOfProcessedEvents int = 0
var elinks []sampa.Payload
var gbt *bitset.BitSet
var inData bool
var nextCheckPoint int

func init() {
	for i := 0; i < 40; i++ {
		elinks = append(elinks, sampa.Payload{BitSet: *(bitset.New(100000))})
	}
	log.Println(len(elinks), "elinks created")
	gbt = bitset.New(80)
	inData = false
	nextCheckPoint = 0
	flag.IntVar(&flagMaxGBTwords, "nw", 0, "max number of GBT words to read")
	flag.IntVar(&flagMaxEvents, "n", 0, "max number of DATE events to read")
	flag.BoolVar(&flagNoColor, "no-color", false, "Disable color output")
	flag.StringVar(&flagCpuProfile, "cpuprofile", "", "write cpu profile to file")
	flag.StringVar(&flagMemProfile, "memprofile", "", "write memory profile to this file")
	log.SetFlags(log.Llongfile)
	log.SetOutput(ioutil.Discard)
}

func main() {
	flag.Parse()
	if flagCpuProfile != "" {
		f, err := os.Create(flagCpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	if flagNoColor {
		color.NoColor = true
	}
	if flag.NArg() == 0 {
		flag.Usage()
		return
	}
	inputFileName := flag.Args()[0]
	r := date.NewReader(inputFileName)
	defer func() {
		fmt.Printf("Happy ending. I've read %d events.",
			r.NofEvents())
	}()
	if r == nil {
		log.Fatal("cannot read file", inputFileName)
	}
	log.Println("Reading from ", inputFileName)
	s := 0
	for n := 0; ; {
		if flagMaxGBTwords > 0 && n >= flagMaxGBTwords {
			break
		}
		if flagMaxEvents > 0 && r.NofEvents() >= flagMaxEvents {
			break
		}
		g, err := r.GBT()
		n++
		s += g.Size() // just to use g for the momemt
		if err != nil {
			if err == io.EOF {
				break
			}
			if err == date.ErrEmptyEvent ||
				err == date.ErrInvalidSOP ||
				err == date.ErrEndOfEvent {
				continue
			}
			log.Fatal(err)
		}
		if n > 100000 && flagMemProfile != "" {
			f, err := os.Create(flagMemProfile)
			if err != nil {
				log.Fatal(err)
			}
			pprof.WriteHeapProfile(f)
			f.Close()
			return
		}
	}
}
