// Program that serves images taken from a webcam.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/aamcrae/rf/io"
	"github.com/aamcrae/rf/message"
)

var port = flag.Int("port", 8080, "Web server port number")
var gpio = flag.Int("gpio", 15, "PRU GPIO bit number")
var messages = flag.String("messages", "/etc/rf-messages", "File containing messages")
var verbose = flag.Bool("v", false, "Log more information")
var repeats = flag.Int("repeats", 3, "Number of message repeats")

func main() {
	flag.Parse()
	tx, err := io.NewTransmitter(uint(*gpio))
	if err != nil {
		log.Fatalf("NewTransmitter: %v", err)
	}
	mList, err := message.ReadMessageFile(*messages)
	if err != nil {
		log.Fatalf("%s: %v", *messages, err)
	}
	for _, m := range mList {
		if *verbose {
			log.Printf("Message %s, length %d", m.Name, len(m.Raw))
		}
		http.Handle(fmt.Sprintf("/tx/%s", m.Name), http.HandlerFunc(handler(tx, m)))
	}
	url := fmt.Sprintf(":%d", *port)
	if *verbose {
		log.Printf("Starting server on %s", url)
	}
	server := &http.Server{Addr: url}
	log.Fatal(server.ListenAndServe())
}

func handler(tx *io.Transmitter, msg *message.Message) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if *verbose {
			log.Printf("Sending message %s", msg.Name)
		}
		err := tx.Send(msg.Raw, *repeats)
		if err != nil {
			log.Printf("Message %s: %v", msg.Name, err)
		}
	}
}
