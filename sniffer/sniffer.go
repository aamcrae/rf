// Program to read a file of captured timings and create an output message
package main

import (
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aamcrae/rf/io"
	"github.com/aamcrae/rf/message"
)

var tolerance = flag.Int("tolerance", 20, "Percent tolerance")
var min_msg = flag.Int("min", 10, "Mininum number of changes")
var max_msg = flag.Int("max", 140, "Maximum number of changes")
var base_time = flag.Int("base", 0, "Microseconds for bit period")
var gpio = flag.Int("gpio", 15, "Input GPIO number for capture")

func main() {
	flag.Parse()
	inp, err := io.NewReceiver(uint(*gpio))
	if err != nil {
		log.Fatalf("GPIO %d receiver failed: %v", *gpio, err)
	}
	defer inp.Close()
	c, err := inp.Start()
	if err != nil {
		log.Fatalf("GPIO %d receiver failed: %v", *gpio, err)
	}
	fmt.Printf("Starting reader - hit enter to exit\n")
	var wg sync.WaitGroup
	wg.Add(1)
	go reader(c, &wg)
	fmt.Scanln()
	inp.Stop()
	wg.Wait()
	fmt.Printf("Reader finished\n")
}

func reader(c <-chan time.Duration, wg *sync.WaitGroup) {
	l := message.NewListener()
	for {
		d := <-c
		if d == 0 {
			wg.Done()
			return
		}
		m := l.Next(int(d.Microseconds()))
		if m != nil {
			fmt.Printf("Msg, len %d\n", len(m))
		}
	}
}
