// Program to read a file of captured timings and create an output message
package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/scanner"
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

const (
	GCD1 = iota	// Min
	GCD2 = iota	// Min / 2
	GCD3 = iota	// Min / 3
	GCD4 = iota	// Min / 4
	MAX_GCD = iota
)

type Message struct {
	index int
	timings []int
	str string
	rle string
}

func main() {
	flag.Parse()
	msgs, err := readMessages(*input)
	if err != nil {
		log.Fatalf("%s: %v", *input, err)
	}
	if len(msgs) == 0 {
		log.Fatalf("No messages found to process")
	}
	if *verbose {
		fmt.Printf("# of messages: %d\n", len(msgs))
	}
	base := *base_time
	if base == 0 {
		// Analyse the message and try and determine a sensible bit period
		base = findBaseTime(msgs)
	}
	// Create a map holding commonly decoded strings
	str_count := make(map[string]int)
	str_message := make(map[string]*Message)
	for _, m := range msgs {
		m.Normalise(base)
		str_count[m.str] = str_count[m.str] + 1
		str_message[m.str] = m
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
				if v == msg_count[len(msg_count) - l - 1] {
					fmt.Fprintf(f, "%s-%d", *input, l)
					sep := ' '
					for _, t := range str_message[s].timings {
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
			fmt.Printf("%3d (%3d): %s\n", c, len(s), summary(s))
		}
	}
}

func readMessages(input string) ([]*Message, error) {
	var msgs []*Message
	var current *Message
	f, err := os.Open(input)
	if err != nil {
		return msgs, err
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
			// Check for end of message gap.
			if int(v) > *gap {
				// Save current message, if any
				if current != nil {
					if len(current.timings) < *min_msg || len(current.timings) >= *max_msg {
						// Skip out-of-range message
						if *verbose {
							fmt.Printf("Skipping message of length %d\n", len(current.timings))
						}
					} else {
						msgs = append(msgs, current)
					}
				}
				current = new(Message)
				current.index = len(msgs)
			// Ignore values before an inter-message gap is seen.
			} else if current != nil {
				current.timings = append(current.timings, int(v))
			}
		}
	}
	// Skip last message, since it is likely truncated.
	return msgs, nil
}

func findBaseTime(msgs []*Message) int {
	if len(msgs) < *min_messages {
		log.Fatalf("Need at least %d messages for estimating sync time", *min_messages)
	}
	var gcds [MAX_GCD]int
	var avg [MAX_GCD]int
	var quality[MAX_GCD]int
	for _, m := range msgs {
		gcd, est_base, q := m.Analyse()
		gcds[gcd]++
		avg[gcd] += est_base
		quality[gcd] += q
	}
	count := -1
	best := -1
	for i, v := range gcds {
		if v != 0 {
			if count < v {
				count = v
				best = i
			}
		}
	}
	if *verbose {
		fmt.Printf("GCD #%d (value %d), count %d, quality %d%%\n", best, avg[best]/count, count, quality[best]/count)
	}
	return avg[best]/count
}

// Analyse one message, trying to estimate the base sync time by
// finding the closest GCD of the message segments.
func (m *Message) Analyse() (int, int, int) {
	st := make([]int, len(m.timings))
	copy(st, m.timings)
	sort.Ints(st)
	total := 0
	count := 0
	cutoff := st[0] + st[0] * 20 / 100
	// Average the minimum values, up to 20% larger.
	for _, v := range st {
		if v > cutoff {
			break;
		}
		total += v
		count++
	}
	min := total/count
	if *verbose {
		fmt.Printf("Min = %d (%d samples)\n", min, count)
	}
	best := -1
	gcds := []int{min, min/2, min/3, min/4}
	matches := make([]int, len(gcds))
	bestGcd := 0
	for g, i := range gcds {
		allow := (*tolerance * i) / 100
		for _, v := range m.timings {
			if _, match := gcd(v, i, allow); match {
				matches[g]++
			}
		}
		if best < matches[g] {
			best = matches[g]
			bestGcd = g
		}
	}
	quality := matches[bestGcd] * 100 / len(m.timings)
	if *verbose {
		fmt.Printf("msg %d, base = %d (quality %d, matched %d)\n", m.index, gcds[bestGcd], quality, matches[bestGcd])
	}
	return bestGcd, gcds[bestGcd], quality
}

func (m *Message) Normalise(base int) {
	var s strings.Builder
	bit := 1
	c := '1'
	r := base/2
	bits := 0
	for i, t := range m.timings {
		count := (t + r) / base
		m.timings[i] = count * base
		for i := 0; i < count; i++ {
			s.WriteRune(c)
		}
		bits += count
		bit ^= 1
		c = '0' + rune(bit)
	}
	m.str = s.String()
}

// Return true if base is close to a factor of v.
func gcd(v, base, t int) (int, bool) {
	d := int(math.Round(float64(v)/float64(base)))
	n := d * base
	return d, n < (v+t) && n > (v-t)
}

// Return a run-length encoded form of 0/1 string.
func summary(s string) string {
	var st strings.Builder
	last := rune(s[0])
	count := 0
	for _, c := range s {
		if last != c {
			st.WriteRune('0' + rune(count))
			count = 1
			last = c
		} else {
			if count == 9 {
				st.WriteString("90")
				count = 0
			}
			count++
		}
	}
	return st.String()
}

