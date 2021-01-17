// Module to send raw messages to a transmitter

package io

import (
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/aamcrae/pru"
)

const defaultGap = 4000
const tx_event = 18

type Transmitter struct {
	pru  *pru.PRU
	gpio uint32
	Gap  int
	lock sync.Mutex
}

func NewTransmitter(gpio uint) (*Transmitter, error) {
	tx := new(Transmitter)
	tx.Gap = defaultGap
	tx.gpio = uint32(gpio)
	pc := pru.NewConfig()
	pc.Event2Channel(tx_event, 2).Channel2Interrupt(2, 2)
	var err error
	tx.pru, err = pru.Open(pc)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (tx *Transmitter) Close() {
	tx.pru.Close()
}

// Send a message.
func (tx *Transmitter) Send(msg []int, repeats int) error {
	// Only one message can be sent at a time.
	tx.lock.Lock()
	defer tx.lock.Unlock()
	u := tx.pru.Unit(0)
	e := tx.pru.Event(tx_event)
	r := u.Ram.Open()
	params := []interface{}{
		uint32(tx_event - 16),                  // Event to send when complete
		uint32(tx.gpio),                        // GPIO to use
		uint32(repeats),                        // Number of message repeats
		uint32(len(msg) + 1),                   // Length of message
		uint32(20),                             // Address of data
		uint32(pru.MicroSeconds2Ticks(tx.Gap)), // Inter-message gap as first time
	}
	for _, v := range params {
		binary.Write(r, tx.pru.Order, v)
	}
	tout := tx.Gap
	for _, t := range msg {
		binary.Write(r, tx.pru.Order, uint32(pru.MicroSeconds2Ticks(t)))
		tout += t
	}
	tout *= repeats
	start := time.Now()
	err := u.Run(prutx_img)
	if err != nil {
		return err
	}
	// Timeout is twice the expected transmission time
	timeout := time.Duration(tout*2) * time.Microsecond
	ok, err := e.WaitTimeout(timeout)
	if err != nil {
		return err
	}
	if !ok {
		u.Disable()
		return fmt.Errorf("transmit timeout (%s)", time.Duration(timeout))
	}
	fmt.Printf("Transmission took %s\n", time.Now().Sub(start).String())
	return nil
}
