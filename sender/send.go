// Program that serves images taken from a webcam.
package main

import (
	"flag"
	"log"

	"github.com/aamcrae/rf/io"
)

var file = flag.String("file", "", "Message file database")
var message = flag.String("message", "", "Message name")
var repeats = flag.Int("repeat", 3, "Number of repeats")
var gpio = flag.Int("gpio", 15, "Output GPIO number")

func main() {
	flag.Parse()

	msgs, err := io.ReadMessageFile(*file)
	if err != nil {
		log.Fatalf("%s", err)
	}
	log.Printf("%d messages read", len(msgs))
	tx, err := io.NewTransmitter(uint(*gpio))
	if err != nil {
		log.Fatalf("%s", err)
	}
	defer tx.Close()
	m, ok := msgs[*message]
	if !ok {
		log.Fatalf("%s: message not found", *message)
	}
	err = tx.Send(m, *repeats)
	if err != nil {
		log.Fatalf("%s: %v", *message, err)
	}
}
