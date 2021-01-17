// Program that serves images taken from a webcam.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aamcrae/rf/io"
	"github.com/aamcrae/rf/message"
)

var port = flag.Int("port", 8080, "Web server port number")
var gpio = flag.Int("gpio", 15, "PRU GPIO bit number")
var messages = flag.String("messages", "/etc/rf-messages", "File containing messages")
var verbose = flag.Bool("v", false, "Log more information")
var repeats = flag.Int("repeats", 1, "Number of message repeats")
var gap = flag.Int("gap", 10, "Inter-message gap")

func main() {
	flag.Parse()
	tx, err := io.NewTransmitter(uint(*gpio))
	if err != nil {
		log.Fatalf("NewTransmitter: %v", err)
	}
	msgs, err := message.ReadTagFile(*messages)
	if err != nil {
		log.Fatalf("%s: %v", *messages, err)
	}
	for tag, m := range msgs {
		if *verbose {
			log.Printf("Message %s, count %d", tag, len(m))
		}
		http.Handle(fmt.Sprintf("/tx/%s", tag), http.HandlerFunc(handler(tx, tag, m)))
	}
	url := fmt.Sprintf(":%d", *port)
	if *verbose {
		log.Printf("Starting server on %s", url)
	}
	server := &http.Server{Addr: url}
	log.Fatal(server.ListenAndServe())
}

func handler(tx *io.Transmitter, tag string, msg []message.Raw) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if *verbose {
			log.Printf("Sending tag %s %d messages", tag, len(msg))
		}
		for i, m := range msg {
			err := tx.Send(m, *repeats)
			if err != nil {
				log.Printf("Message %d: %v", tag, i, err)
			}
			time.Sleep(time.Duration(*gap) * time.Millisecond)
		}
	}
}
