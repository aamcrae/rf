// Program that serves images taken from a webcam.
package main

import (
	"flag"
	"log"
	"time"

	"github.com/aamcrae/rf/io"
	"github.com/aamcrae/rf/message"
)

var file = flag.String("file", "", "Message file database")
var msg = flag.String("message", "", "Message name")
var repeats = flag.Int("repeat", 3, "Number of repeats")
var gap = flag.Int("gap", 10, "Inter-message gap (milliseconds)")
var gpio = flag.Int("gpio", 15, "Output GPIO number") // PRU unit 0 P8_11

func main() {
	flag.Parse()

	msgs, err := message.ReadTagFile(*file)
	if err != nil {
		log.Fatalf("%s", err)
	}
	log.Printf("%d messages read", len(msgs))
	tx, err := io.NewTransmitter(uint(*gpio))
	if err != nil {
		log.Fatalf("%s", err)
	}
	defer tx.Close()
	ml, ok := msgs[*msg]
	if !ok {
		log.Fatalf("%s: message not found", *msg)
	}
	for rep := 0; rep < *repeats; rep++ {
		for i, m := range ml {
			err = tx.Send(m, 1)
			if err != nil {
				log.Fatalf("%s (%d) repeat %d: %v", *msg, i+1, rep+1, err)
			}
			time.Sleep(time.Duration(*gap) * time.Millisecond)
		}
	}
}
