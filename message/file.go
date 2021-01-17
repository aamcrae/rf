package message

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ReadTagFile reads and unpacks a RF message file
// The format is:
//  <tag> message-timings
//
// The message timings are microsecond intervals for 1-0-1-0... transitions.
func ReadTagFile(name string) (map[string][]Raw, error) {
	msgs := make(map[string][]Raw)
	f, err := os.Open(name)
	if err != nil {
		return nil, err
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
		msgs[strs[0]] = append(msgs[strs[0]], Raw(raw))
	}
	return msgs, nil
}
