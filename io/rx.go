// Module to read raw signals

package io

import (
	"encoding/binary"
	"sync"
	"time"

	"github.com/aamcrae/pru"
)

const rx_event = 19
const rxBufferSize = 128 // 512 bytes each buffer
const bufCount = 8       // 4K of buffers
const rxUnit = 1

type Receiver struct {
	pru      *pru.PRU
	gpio     uint32
	bufReady chan uint32
	send     chan time.Duration
	unit     *pru.Unit
	event    *pru.Event
	buffer   int
	wg       sync.WaitGroup
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

func (rx *Receiver) Start() (<-chan time.Duration, error) {
	rx.unit = rx.pru.Unit(rxUnit)
	rx.event = rx.pru.Event(rx_event)
	r := rx.unit.Ram.Open()
	params := []interface{}{
		uint32(rx_event - 16), // Event to send when buffer full
		uint32(rx.gpio),       // GPIO to use
		uint32(bufCount),      // Count of buffers
		uint32(rxBufferSize),  // Buffer size
		// ... buf addresses
	}
	for _, v := range params {
		binary.Write(r, rx.pru.Order, v)
	}
	bufStart := 0x100
	rx.buffer = 0
	for i := 0; i < bufCount; i++ {
		binary.Write(r, rx.pru.Order, uint32(bufStart))
		binary.Write(r, rx.pru.Order, uint32(0))
		bufStart += rxBufferSize * 4
	}
	rx.bufReady = make(chan uint32, 10)
	send := make(chan time.Duration, 200)
	rx.wg.Add(1)
	rx.event.SetHandler(rx.fullBuffer)
	go rx.readBuffer(send)
	err := rx.unit.Run(prurx_img)
	if err != nil {
		rx.Stop()
		return nil, err
	}
	return send, nil
}

func (rx *Receiver) Stop() {
	rx.unit.Disable()
	rx.event.ClearHandler()
	close(rx.bufReady)
	rx.wg.Wait()
}

// Read data from PRU
func (rx *Receiver) Read(max int) ([]time.Duration, error) {
	var tm []time.Duration
	c, err := rx.Start()
	if err != nil {
		return nil, err
	}
	defer rx.Stop()
	for i := 0; i < max; i++ {
		tm = append(tm, <-c)
	}
	return tm, nil
}

// Event handler
func (rx *Receiver) fullBuffer() {
	select {
	case rx.bufReady <- uint32(0x100 + rx.buffer*4*rxBufferSize):
	default:
		panic("Buffer overflow")
	}
	rx.buffer = (rx.buffer + 1) % bufCount
}

func (rx *Receiver) readBuffer(send chan time.Duration) {
	for {
		b := <-rx.bufReady
		if b == 0 {
			close(send)
			rx.wg.Done()
			return
		}
		for i := 0; i < rxBufferSize; i++ {
			send <- pru.Duration(int(rx.pru.Order.Uint32(rx.unit.Ram[b:])))
			b += 4
		}
	}
}
