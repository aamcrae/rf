// Program to read a file of captured timings and create an output message
package message

import (
	"strings"
)

type Message struct {
	Name  string
	Raw   Raw
	Count []int
	Base  int
	Str   string
	RLE   string
}

func NewMessage(raw Raw, base int) *Message {
	m := new(Message)
	m.Raw = raw
	m.Base = base
	m.Count = raw.Normalise(base)
	var s strings.Builder
	var rle strings.Builder
	bit := 1
	c := '1'
	for _, t := range m.Count {
		for i := 0; i < t; i++ {
			s.WriteRune(c)
		}
		for t > 9 {
			rle.WriteString("90")
			t -= 9
		}
		rle.WriteRune('0' + rune(t))
		bit ^= 1
		c = '0' + rune(bit)
	}
	m.Str = s.String()
	m.RLE = rle.String()
	return m
}

func (m *Message) MatchRaw(raw Raw, start int, tolerance int) int {
	if len(raw) != len(m.Raw) {
		return 0
	}
	count := 0
	// Per bit tolerance
	slop := (m.Base * tolerance) / 100
	for i := start; i < len(raw); i++ {
		allow := slop * m.Count[i]
		mid := m.Count[i] * m.Base
		r := raw[i]
		if r >= (mid - allow) && r < (mid + allow) {
			count++
		}
	}
	return count
}
