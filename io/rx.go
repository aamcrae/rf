// Module to read raw signals

package io

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/aamcrae/pru"
)

const rx_event = 18

type Receiver struct {
	pru *pru.PRU
	gpio uint32
}

func NewReceiver(gpio uint) (*Receiver, error) {
	rx := new(Receiver)
	rx.gpio = uint32(gpio)
	pc := pru.NewConfig()
	pc.Event2Channel(rx_event, 2).Channel2Interrupt(2, 2)
	var err error
	rx.pru, err = pru.Open(pc)
	if err != nil {
		return nil, err
	}
	return rx, nil
}

func (rx *Receiver) Close() {
	rx.pru.Close()
}

// Read data from GPIO
func (rx *Receiver) Read(max int) ([]int, error) {
	u := rx.pru.Unit(0)
	e := rx.pru.Event(rx_event)
	r := u.Ram.Open()
	params := []interface{}{
		uint32(rx_event - 16),	// Event to send when complete
		uint32(rx.gpio),	// GPIO to use
		uint32(20),			// Address for data
		uint32(max),		// Maximum count
		uint32(0),			// Running count
	}
	for _, v := range params {
		binary.Write(r, rx.pru.Order, v)
	}
	fmt.Printf("Running capture on GPIO %d, using event %d, max capture %d\n", rx.gpio, rx_event, max)
	start := time.Now()
	err := u.Run(prurx_img)
	if err != nil {
		return nil, err
	}
	ok, err := e.WaitTimeout(time.Second * 10)
	elapsed := time.Now().Sub(start)
	if err != nil {
		return nil, err
	}
	if !ok {
		// If timed out, stop the unit
		u.Disable()
		fmt.Printf("Timed out\n")
	}
	count := rx.pru.Order.Uint32(u.Ram[16:])
	fmt.Printf("count = %d, time taken = %s\n", count, elapsed.String())
	tm := make([]int, count)
	r.Seek(20, io.SeekStart)
	for i := 0; i < int(count); i++ {
		var v uint32
		binary.Read(r, rx.pru.Order, &v)
		tm[i] = int(pru.Duration(int(v)).Microseconds())
	}
	return tm, nil
}
