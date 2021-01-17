// Program to capture RF 433 signals
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/aamcrae/rf/io"
)

var verbose = flag.Bool("v", false, "Log more information")
var samples = flag.Int("samples", 500, "Number of samples")
var gpio = flag.Int("gpio", 5, "Output GPIO number") // PRU Unit 1, P8_42

func main() {
	flag.Parse()

	inp, err := io.NewReceiver(uint(*gpio))
	if err != nil {
		log.Fatalf("%s", err)
	}
	tm, err := inp.Read(*samples)
	if err != nil {
		log.Fatalf("%s", err)
	}
	for _, n := range tm {
		fmt.Printf("%d,", n.Microseconds())
	}
	fmt.Printf("\n")
}
