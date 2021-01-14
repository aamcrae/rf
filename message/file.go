package message

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const MinMessageLength = 5

// ReadMessageFile reads and unpacks a RF message file
// The format is:
//  <text-key> message-timings
//
// The message timings are microsecond intervals for 1-0-1-0... transitions.
func ReadMessageFile(name string) (map[string][]time.Duration, error) {
	msgs := make(map[string][]time.Duration)
	f, err := os.Open(name)
	if err != nil {
		return msgs, err
	}
	defer f.Close()
	scan := bufio.NewScanner(f)
	lineno := 0
	for scan.Scan() {
		lineno++
		strs := strings.Split(scan.Text(), " ")
		if len(strs) != 2 {
			return msgs, fmt.Errorf("%s: line %d: unknown format", name, lineno)
		}
		ts := strings.Split(strs[1], ",")
		if len(ts) < MinMessageLength {
			return msgs, fmt.Errorf("%s: line %d: Bad message length", name, lineno)
		}
		var msg []time.Duration
		for i, t := range ts {
			v, err := strconv.ParseInt(t, 10, 32)
			if err != nil {
				return msgs, fmt.Errorf("%s: line %d, timing %d (%s) bad format", name, lineno, i, t)
			}
			msg = append(msg, time.Duration(v)*time.Microsecond)
		}
		msgs[strs[0]] = msg
	}
	return msgs, nil
}
