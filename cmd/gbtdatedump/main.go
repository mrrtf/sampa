package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"

	"github.com/aphecetche/bitset"
	"github.com/aphecetche/sampa/date"
	"github.com/aphecetche/sampa/sampa"
	"github.com/fatih/color"
)

var flagMaxGBTwords int
var flagNoColor bool
var flagCpuProfile string

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
	flag.IntVar(&flagMaxGBTwords, "n", -1, "max number of GBT words to read")
	flag.BoolVar(&flagNoColor, "no-color", false, "Disable color output")
	flag.StringVar(&flagCpuProfile, "cpuprofile", "", "write cpu profile to file")
	log.SetFlags(log.Llongfile)
	// log.SetOutput(ioutil.Discard)
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
	if r == nil {
		log.Fatal("cannot read file", inputFileName)
	}
	log.Println("Reading from ", inputFileName)
	s := 0
	for n := 0; n < flagMaxGBTwords || flagMaxGBTwords < 0; n++ {
		g, err := r.GBT()
		s += g.Size()
		if err != nil {
			break
		}
	}
}
