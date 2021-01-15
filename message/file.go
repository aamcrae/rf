package message

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ReadMessageFile reads and unpacks a RF message file
// The format is:
//  <text-key> base message-timings
//
// The message timings are microsecond intervals for 1-0-1-0... transitions.
func ReadMessageFile(name string) (map[string]*Message, error) {
	msgs := make(map[string]*Message)
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
		if len(strs) != 3 {
			return msgs, fmt.Errorf("%s: line %d: unknown format", name, lineno)
		}
		base, err := strconv.ParseInt(strs[1], 10, 32)
		if err != nil {
			return msgs, fmt.Errorf("%s: line %d, bad base (%s)", name, lineno, strs[1])
		}
		ts := strings.Split(strs[2], ",")
		if len(ts) < 5 {
			return msgs, fmt.Errorf("%s: line %d: Bad message length", name, lineno)
		}
		var raw []int
		for i, t := range ts {
			v, err := strconv.ParseInt(t, 10, 32)
			if err != nil {
				return msgs, fmt.Errorf("%s: line %d, timing %d (%s) bad format", name, lineno, i, t)
			}
			raw = append(raw, int(v))
		}
		msgs[strs[0]] = NewMessage(Raw(raw), int(base))
	}
	return msgs, nil
}
