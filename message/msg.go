// Program to read a file of captured timings and create an output message
package message

import (
	"strings"
)

type Message struct {
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
