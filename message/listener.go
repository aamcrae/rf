package message

type Listener struct {
	timings  Raw
	bit      int
	Sync     int
	Gap      int
	Min      int
	Max      int
	Debounce int
}

func NewListener() *Listener {
	l := new(Listener)
	l.Gap = 4000
	l.Min = 10
	l.Max = 200
	l.Debounce = 20
	return l
}

func (l *Listener) Clear() {
	l.bit = 0
	l.timings = nil
}

func (l *Listener) Next(tv int) Raw {
	b := l.bit
	l.bit ^= 1 // flip bit
	if tv < l.Debounce {
		// Pulse length is too short, likely noise, discard message
		l.timings = nil
		return nil
	}
	// Check for end of message gap.
	if b == 0 && tv > l.Gap {
		if len(l.timings) >= l.Min && len(l.timings) < l.Max {
			t := l.timings
			l.timings = make([]int, 0)
			return t // Return message
		} else {
			// Discard out-of-range message.
			l.timings = make([]int, 0)
			return nil
		}
	}
	// Ignore values until an intermessage gap is seen.
	if l.timings != nil {
		l.timings = append(l.timings, tv)
		if len(l.timings) >= l.Max {
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
