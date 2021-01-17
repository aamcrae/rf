package message

type Listener struct {
	timings       Raw
	bit           int
	Gap           int
	MinLen        int
	MaxLen        int
	MinPulse      int
	ShortestPulse int

	Noise    int
	Overflow int
	Runt     int
}

func NewListener() *Listener {
	l := new(Listener)
	l.Gap = 4000
	l.MinLen = 10
	l.MaxLen = 200
	l.MinPulse = 20
	l.Clear()
	return l
}

func (l *Listener) Clear() {
	l.bit = 0
	l.timings = nil
	l.Noise = 0
	l.Overflow = 0
	l.Runt = 0
	l.ShortestPulse = l.Gap + 1
}

func (l *Listener) Next(tv int) Raw {
	b := l.bit
	l.bit ^= 1 // flip bit
	if tv < l.MinPulse {
		// Pulse length is too short, likely noise, discard message
		if l.timings != nil {
			l.Noise++
		}
		l.timings = nil
		return nil
	}
	// Check for end of message gap.
	if b == 0 && tv > l.Gap {
		if len(l.timings) >= l.MinLen && len(l.timings) < l.MaxLen {
			t := l.timings
			l.timings = make([]int, 0)
			return t // Return message
		} else {
			// Discard out-of-range message.
			l.Runt++
			l.timings = make([]int, 0)
			return nil
		}
	}
	// Ignore values until an intermessage gap is seen.
	if l.timings != nil {
		if tv < l.ShortestPulse {
			l.ShortestPulse = tv
		}
		l.timings = append(l.timings, tv)
		if len(l.timings) >= l.MaxLen {
			l.Overflow++
			l.timings = nil
		}
	}
	return nil
}

// Given a slice of raw timings, extract all the messages.
func (l *Listener) Decode(rawInput []int) []Raw {
	l.Clear()
	var msgs []Raw
	for _, tv := range rawInput {
		t := l.Next(tv)
		if t != nil {
			msgs = append(msgs, t)
		}
	}
	return msgs
}
