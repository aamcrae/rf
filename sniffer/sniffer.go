// Program to read a file of captured timings and create an output message
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"strconv"
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

type msg struct {
	base message.Base
	messages []message.Raw
}

var reference = make(map[int][]*message.Message)
var messages = make(map[int]*msg)
var baseAll message.Base

func main() {
	flag.Parse()
	baseAll.tolerance = *tolerance
	if len(*referenceFile) > 0 {
		rList, err := message.ReadMessageFile(*referenceFile)
		if err != nil {
			log.Fatalf("%s: %v", *referenceFile, err)
		}
		for _, r := range rList {
			reference[len(r.Raw)] = append(reference[len(r.Raw)], r)
			fmt.Printf("Ref msg %s, len %d\n", r.Name, len(r.Raw))
		}
	}
	l := message.NewListener()
	l.Gap = *gap
	l.Min = *min_msg
	l.Max = *max_msg
	if len(*input) > 0 {
		readFromFile(*input, l)
	} else {
		capture(l)
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
	mp, ok := messages[l]
	if !ok {
		mp = new(msg)
		mp.base.Tolerance = *tolerance
		messages[l] = mp
	}
	mp.messages = append(mp.messages, m)
	mp.base.Add(m)
	base, quality := mp.base.EstimateBase(*round)
	fmt.Printf("len %d, %d messages, estimated base %d (quality %d)\n", l, len(mp.messages), base, quality)
	for _, msg := range reference[l] {
		// Skip sync pulse when checking for equality
		b := *base_time
		if b == 0 {
			b = base
		}
		n := m.Normalise(b)
		nr := make([]int, len(n))
		for i, n := range n {
			nr[i] = b * n
		}
		rmatch := msg.MatchRaw(m, 1, *tolerance);
		nmatch := msg.MatchRaw(nr, 1, *tolerance)
		fmt.Printf("%s: len %d Raw %d (%d) Normalsed %d (%d)\n", msg.Name, len(m), rmatch * 100 / len(m), rmatch,  nmatch * 100 / len(m), nmatch)
	}
}
