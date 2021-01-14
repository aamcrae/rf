// Program to read a file of captured timings and create an output message
package message

import (
	"math"
	"sort"
	"strings"
)

const (
	GCD1    = iota // Min
	GCD2    = iota // Min / 2
	GCD3    = iota // Min / 3
	GCD4    = iota // Min / 4
	MAX_GCD = iota
)

type Message struct {
	Timings []int
	Str     string
	Rle     string
}

func Decode(t []int) []*Message {
	l := NewListener()
	var msgs []*Message
	for _, tv := range t {
		t := l.Next(tv)
		if t != nil {
			m := new(Message)
			m.Timings = t
			msgs = append(msgs, m)
		}
	}
	return msgs
}

func BaseTime(msgs []*Message, tolerance int) int {
	var gcds [MAX_GCD]int
	var avg [MAX_GCD]int
	var quality [MAX_GCD]int
	for _, m := range msgs {
		gcd, est_base, q := m.Analyse(tolerance)
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
	return avg[best] / count
}

// Analyse one message, trying to estimate the base sync time by
// finding the closest GCD of the message segments.
func (m *Message) Analyse(tolerance int) (int, int, int) {
	st := make([]int, len(m.Timings))
	copy(st, m.Timings)
	sort.Ints(st)
	total := 0
	count := 0
	cutoff := st[0] + st[0]*20/100
	// Average the minimum values, up to 20% larger.
	for _, v := range st {
		if v > cutoff {
			break
		}
		total += v
		count++
	}
	min := total / count
	best := -1
	gcds := []int{min, min / 2, min / 3, min / 4}
	matches := make([]int, len(gcds))
	bestGcd := 0
	for g, i := range gcds {
		allow := (tolerance * i) / 100
		for _, v := range m.Timings {
			if _, match := gcd(v, i, allow); match {
				matches[g]++
			}
		}
		if best < matches[g] {
			best = matches[g]
			bestGcd = g
		}
	}
	quality := matches[bestGcd] * 100 / len(m.Timings)
	return bestGcd, gcds[bestGcd], quality
}

func (m *Message) Normalise(base int) {
	var s strings.Builder
	bit := 1
	c := '1'
	r := base / 2
	bits := 0
	for i, t := range m.Timings {
		count := (t + r) / base
		m.Timings[i] = count * base
		for i := 0; i < count; i++ {
			s.WriteRune(c)
		}
		bits += count
		bit ^= 1
		c = '0' + rune(bit)
	}
	m.Str = s.String()
}

// Return true if base is close to a factor of v.
func gcd(v, base, t int) (int, bool) {
	d := int(math.Round(float64(v) / float64(base)))
	n := d * base
	return d, n < (v+t) && n > (v-t)
}

// Return a run-length encoded form of 0/1 string.
func Summary(s string) string {
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
