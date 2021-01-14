// Program that serves images taken from a webcam.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aamcrae/rf/io"
)

var port = flag.Int("port", 8080, "Web server port number")
var gpio = flag.Int("gpio", 15, "PRU GPIO bit number")
var messages = flag.String("messages","/etc/rf-messages", "File containing messages")
var verbose = flag.Bool("v", false, "Log more information")
var repeats = flag.Int("repeats", 3, "Number of message repeats")

func main() {
	flag.Parse()
	tx, err := io.NewTransmitter(uint(*gpio))
	if err != nil {
		log.Fatalf("NewTransmitter: %v", err)
	}
	msgs, err := io.ReadMessageFile(*messages)
	if err != nil {
		log.Fatalf("%s: %v", *messages, err)
	}
	for k, m := range msgs {
		if *verbose {
			log.Printf("Message %s, length %d", k, len(m))
		}
		http.Handle(fmt.Sprintf("/tx/%s", k), http.HandlerFunc(handler(tx, k, m)))
	}
	url := fmt.Sprintf(":%d", *port)
	if *verbose {
		log.Printf("Starting server on %s", url)
	}
	server := &http.Server{Addr: url}
	log.Fatal(server.ListenAndServe())
}

func handler(tx *io.Transmitter, key string, msg []time.Duration) (func(http.ResponseWriter, *http.Request)) {
	return func(w http.ResponseWriter, r *http.Request) {
		if *verbose {
			log.Printf("Sending message %s", key)
		}
		err := tx.Send(msg, *repeats)
		if err != nil {
			log.Printf("Message %s: %v", key, err)
		}
	}
}
