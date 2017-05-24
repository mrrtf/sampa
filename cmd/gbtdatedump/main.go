package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/fatih/color"
	"github.com/mrrtf/sampa/pkg/bitset"
	"github.com/mrrtf/sampa/pkg/date"
	"github.com/mrrtf/sampa/pkg/sampa"
)

var flagMaxGBTwords int
var flagNoColor bool
var flagCpuProfile string
var flagMemProfile string
var flagMaxEvents int
var flagNoDispatch bool
var flagMaskELink uint64
var NumberOfProcessedEvents int = 0
var elinks []sampa.ELink
var gbt *bitset.BitSet
var inData bool
var nextCheckPoint int

func init() {
	for i := 0; i < 40; i++ {
		elinks = append(elinks, sampa.NewELink(i))
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
	flag.BoolVar(&flagNoDispatch, "no-dispatch", false, "Disable GBT to elink dispatching")
	flag.Uint64Var(&flagMaskELink, "elink-mask", 0, "40 bits mask to describe which elinks to skip in decoding (default none)")
	log.SetFlags(log.Llongfile)
	// log.SetOutput(ioutil.Discard)
}

func main() {

	fmt.Println(runtime.GOMAXPROCS(1))

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
		fmt.Printf("Happy ending. I've read %d events and %d GBT words\n",
			r.NofEvents(), r.NofGBTwords())
	}()
	if r == nil {
		log.Fatal("cannot read file", inputFileName)
	}
	log.Println("Reading from ", inputFileName)
	ten := make([]byte, 10)
	for {
		if flagMaxGBTwords > 0 && r.NofGBTwords() >= flagMaxGBTwords {
			break
		}
		if flagMaxEvents > 0 && r.NofEvents() >= flagMaxEvents {
			break
		}

		n, err := r.Read(ten)

		if err != nil {

			if err == io.EOF {
				break
			}
			if err == date.ErrEndOfEvent {
				// fmt.Println("end of event ", r.NofEvents())
				continue
			}
			log.Fatal(err)
		}

		// fmt.Println("GBT=", r.GBTAsString())
		if n != 10 {
			log.Fatalf("Could not read the expected 10 bytes, but %d ones", n)
		}
		if len(ten) != n {
			log.Fatalf("inconsistent slice returned : size is %d while I was expecting %d", len(ten), n)
		}
		if flagNoDispatch {
			continue
		}

		err = sampa.Dispatch(ten, elinks, flagMaskELink)

		if err != nil {
			log.Printf("ten size is %d", len(ten))
			log.Fatal(err)
		}
		if r.NofGBTwords() > 100000 && flagMemProfile != "" {
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

func dumpElinks(elinks []sampa.ELink) {
	for i := 0; i < len(elinks); i++ {
		e := elinks[i]
		if !e.IsEmpty() {
			fmt.Println("elink ", i, e)
		}
		i++
	}
}
