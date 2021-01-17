package message

const (
	GCD1    = iota // Min
	GCD2    = iota // Min / 2
	GCD3    = iota // Min / 3
	GCD4    = iota // Min / 4
	MAX_GCD = iota
)

type Base struct {
	Tolerance int
	Count     int
	gcds      [MAX_GCD]int
	avg       [MAX_GCD]int64
	quality   [MAX_GCD]int64
}

// Analyse the raw message and save the data.
func (b *Base) Add(msg Raw) {
	gcd, est, q := msg.Analyse(b.Tolerance)
	b.gcds[gcd]++
	b.avg[gcd] += int64(est)
	b.quality[gcd] += int64(q)
	b.Count++
}

// EstimateBase attempts to extract a bit length base.
func (b *Base) EstimateBase(round int) (int, int) {
	count := -1
	best := -1
	for i, v := range b.gcds {
		if v != 0 {
			if count < v {
				count = v
				best = i
			}
		}
	}
	final := (b.avg[best]/int64(count) + int64(round/2)) / int64(round)
	return int(final) * round, int(b.quality[best] / int64(count))
}
