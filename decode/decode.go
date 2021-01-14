// Program to read a file of captured timings and create an output message
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"text/scanner"

	"github.com/aamcrae/rf/io"
	"github.com/aamcrae/rf/message"
)

var verbose = flag.Bool("v", false, "Log more information")
var gap = flag.Int("gap", 4000, "Inter-message gap time")
var input = flag.String("input", "", "Input file to be read")
var output = flag.String("output", "", "File for writing message strings")
var tolerance = flag.Int("tolerance", 20, "Percent tolerance")
var min_msg = flag.Int("min", 10, "Mininum number of changes")
var max_msg = flag.Int("max", 140, "Maximum number of changes")
var min_messages = flag.Int("messages", 3, "Mininum number of messages")
var base_time = flag.Int("base", 0, "Microseconds for bit period")
var output_limit = flag.Int("limit", 1, "Number of output messages to save")
var capture = flag.Int("capture", 500, "Number of signals to capture")
var gpio = flag.Int("gpio", 15, "Input GPIO number for capture")

func main() {
	flag.Parse()
	var timings []int
	var err error
	var name string
	if len(*input) > 0 {
		timings, err = readMessages(*input)
		if err != nil {
			log.Fatalf("%s: %v", *input, err)
		}
		name = *input
	} else {
		timings, err = rxCapture(*capture)
		if *verbose {
			for _, n := range timings {
				fmt.Printf("%d,", n)
			}
			fmt.Printf("\n")
		}
		name = "capture"
	}
	msgs := message.Decode(timings)
	if len(msgs) == 0 {
		log.Fatalf("No messages found to process")
	}
	if *verbose {
		fmt.Printf("# of messages: %d\n", len(msgs))
	}
	base := *base_time
	if base == 0 {
		// Analyse the message and try and determine a sensible bit period
		if len(msgs) < *min_messages {
			log.Fatalf("Need at least %d messages for estimating sync time", *min_messages)
		}
		base = message.BaseTime(msgs, *tolerance)
	}
	// Create a map holding commonly decoded strings
	str_count := make(map[string]int)
	str_message := make(map[string]*message.Message)
	for _, m := range msgs {
		m.Normalise(base)
		str_count[m.Str] = str_count[m.Str] + 1
		str_message[m.Str] = m
	}
	var msg_count []int
	for _, v := range str_count {
		msg_count = append(msg_count, v)
	}
	sort.Ints(msg_count)
	if len(*output) != 0 {
		f, err := os.Create(*output)
		if err != nil {
			log.Fatalf("%s: %v", *output, err)
		}
		defer f.Close()
		for s, v := range str_count {
			for l := 0; l < *output_limit; l++ {
				if v == msg_count[len(msg_count)-l-1] {
					fmt.Fprintf(f, "%s-%d", name, l)
					sep := ' '
					for _, t := range str_message[s].Timings {
						fmt.Fprintf(f, "%c%d", sep, t)
						sep = ','
					}
					fmt.Fprint(f, "\n")
				}
			}
		}
	}
	if *verbose {
		for s, c := range str_count {
			fmt.Printf("%3d (%3d): %s\n", c, len(s), message.Summary(s))
		}
	}
}

func rxCapture(max int) ([]int, error) {
	inp, err := io.NewReceiver(uint(*gpio))
	if err != nil {
		return nil, err
	}
	defer inp.Close()
	dm, err := inp.Read(max)
	if err != nil {
		return nil, err
	}
	tm := make([]int, len(dm))
	for i := range dm {
		tm[i] = int(dm[i].Microseconds())
	}
	return tm, nil
}

func readMessages(input string) ([]int, error) {
	var t []int
	f, err := os.Open(input)
	if err != nil {
		return t, err
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
				fmt.Printf("Illegal value: %s\n", s.TokenText())
				continue
			}
			t = append(t, int(v))
		}
	}
	return t, nil
}
