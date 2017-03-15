package main

import (
	"flag"
	"log"
	"os"

	"github.com/aphecetche/sampa/edm"
	"github.com/fatih/color"
)

var flagMaxNofEvents int
var flagNoColor bool

var nevents = 0

func init() {
	flag.IntVar(&flagMaxNofEvents, "n", 0, "max number of events to read")
	flag.BoolVar(&flagNoColor, "no-color", false, "Disable color output")
}

func main() {

	flag.Parse()

	if flagNoColor {
		color.NoColor = true
	}

	if flag.NArg() == 0 {
		flag.Usage()
		return
	}

	inputFileName := flag.Args()[0]

	file, err := os.Open(inputFileName)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}

	var nevents int

	for {
		event, err := edm.GetEvent(file)
		if err != nil {
			break
		}
		nevents++

		edm.ProcessEvent(event)

		if flagMaxNofEvents > 0 && nevents >= flagMaxNofEvents {
			break
		}
	}
}
