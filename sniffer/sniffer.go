// Program to read a file of captured timings and create an output message
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"text/scanner"
	"time"

	"github.com/aamcrae/rf/io"
	"github.com/aamcrae/rf/message"
)

var tolerance = flag.Int("tolerance", 20, "Percent tolerance")
var referenceFile = flag.String("messages", "", "Database of RF messages for reference")
var gap = flag.Int("gap", 4000, "Mininum number of changes")
var min_msg = flag.Int("min", 10, "Mininum number of changes")
var max_msg = flag.Int("max", 140, "Maximum number of changes")
var base_time = flag.Int("base", 0, "Microseconds for bit period")
var round = flag.Int("round", 10, "Round up base time")
var gpio = flag.Int("gpio", 15, "Input GPIO number for capture")
var input = flag.String("input", "", "Input file to be read")
var debounce = flag.Int("debounce", 100, "Minimum time for transition")
var tag = flag.String("tag", "tag", "Message tag for output")
var output = flag.String("output", "", "Output filename")

type msg struct {
	base     message.Base
	messages []message.Raw
}

var tags map[string][]message.Raw
var lenMap = make(map[int]*msg)
var messages []message.Raw
var baseAll message.Base

func main() {
	flag.Parse()
	baseAll.Tolerance = *tolerance
	if len(*referenceFile) > 0 {
		var err error
		tags, err = message.ReadTagFile(*referenceFile)
		if err != nil {
			log.Fatalf("%s: %v", *referenceFile, err)
		}
	}
	l := message.NewListener()
	l.Gap = *gap
	l.MinLen = *min_msg
	l.MaxLen = *max_msg
	l.MinPulse = *debounce
	if len(*input) > 0 {
		readFromFile(*input, l)
	} else {
		capture(l)
	}
	fmt.Printf("Noise skipped msgs = %d, overflow = %d, runts = %d, min timing = %d\n", l.Noise, l.Overflow, l.Runt, l.ShortestPulse)
	if len(*output) > 0 {
		f, err := os.OpenFile(*output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		for _, m := range messages {
			m.Write(f, *tag)
		}
	}
}

func readFromFile(input string, l *message.Listener) {
	f, err := os.Open(input)
	if err != nil {
		log.Fatalf("%s: %v", input, err)
	}
	defer f.Close()
	var s scanner.Scanner
	s.Init(f)
	s.Whitespace |= 1 << ','
	s.Mode |= scanner.ScanInts
	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		if tok == scanner.Int {
			v, err := strconv.ParseInt(s.TokenText(), 10, 32)
			if err != nil {
				log.Fatalf("Illegal value: %s: %v\n", s.TokenText(), err)
			}
			m := l.Next(int(v))
			if m != nil {
				newMessage(m)
			}
		}
	}
}

func capture(l *message.Listener) {
	inp, err := io.NewReceiver(uint(*gpio))
	if err != nil {
		log.Fatalf("GPIO %d receiver failed: %v", *gpio, err)
	}
	defer inp.Close()
	c, err := inp.Start()
	if err != nil {
		log.Fatalf("GPIO %d receiver failed: %v", *gpio, err)
	}
	fmt.Printf("Starting Capture - hit enter to exit\n")
	var wg sync.WaitGroup
	wg.Add(1)
	go reader(c, &wg, l)
	fmt.Scanln()
	inp.Stop()
	wg.Wait()
	fmt.Printf("Capture finished\n")
}

func reader(c <-chan time.Duration, wg *sync.WaitGroup, l *message.Listener) {
	for {
		d := <-c
		if d == 0 {
			wg.Done()
			return
		}
		m := l.Next(int(d.Microseconds()))
		if m != nil {
			newMessage(m)
		}
	}
}

func newMessage(m message.Raw) {
	baseAll.Add(m)
	l := len(m)
	mp, ok := lenMap[l]
	if !ok {
		mp = new(msg)
		mp.base.Tolerance = *tolerance
		lenMap[l] = mp
	}
	mp.messages = append(mp.messages, m)
	messages = append(messages, m)
	mp.base.Add(m)
	base, quality := mp.base.EstimateBase(*round)
	fmt.Printf("len %d, %d messages, estimated base %d (quality %d)\n", l, len(mp.messages), base, quality)
}
