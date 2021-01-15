package message

import (
	"math"
	"sort"
)

const (
	GCD1    = iota // Min
	GCD2    = iota // Min / 2
	GCD3    = iota // Min / 3
	GCD4    = iota // Min / 4
	MAX_GCD = iota
)

type Raw []int

// Analyse one raw message, trying to estimate the base sync time by
// finding the closest GCD of the message segments.
func (m Raw) Analyse(tolerance int) (int, int, int) {
	st := make([]int, len(m))
	copy(st, m)
	sort.Ints(st)
	total := 0
	count := 0
	// Average the minimum values, up to 20% larger.
	cutoff := st[0] + st[0]*20/100
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
		for _, v := range st {
			if _, match := gcd(v, i, allow); match {
				matches[g]++
			}
		}
		if best < matches[g] {
			best = matches[g]
			bestGcd = g
		}
	}
	quality := matches[bestGcd] * 100 / len(st)
	return bestGcd, gcds[bestGcd], quality
}

// Normalise returns a normalised slice of timings, where
// each timing is rounded to a multiple of base.
func (m Raw) Normalise(base int) []int {
	n := make([]int, len(m))
	r := base / 2
	for i, t := range m {
		// Round timing to the nearest base.
		count := (t + r) / base
		n[i] = count
	}
	return n
}

// Return true if base is close to a factor of v.
func gcd(v, base, t int) (int, bool) {
	d := int(math.Round(float64(v) / float64(base)))
	n := d * base
	return d, n < (v+t) && n > (v-t)
}

// EstimateBase attempts to extract a bit length base from the raw messages
func EstimateBase(msgs []Raw, tolerance int) int {
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
