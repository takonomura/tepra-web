package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

var (
	tcpDialTimeout  = 1 * time.Second
	tcpWriteTimeout = 5 * time.Second
	printTimeout    = 30 * time.Second
)

func (t Tape) writeImage(w io.Writer) {
	bytes := (t.Width() + 7) / 8
	for x := t.Length() - 1; x >= 0; x-- {
		w.Write([]byte{0x1b, 0x2e, 0x00, 0x00, 0x00, 0x01})
		binary.Write(w, binary.LittleEndian, t.Width())
		for i := uint16(0); i < bytes; i++ {
			var b byte
			for j := uint16(0); j < 8; j++ {
				y := i*8 + j
				if y < t.Width() {
					b |= t.getPixel(x, y)
				}
				if j < 7 {
					b <<= 1
				}
			}
			w.Write([]byte{b})
		}
		if x == 0 {
			break
		}
	}
	w.Write([]byte{0x0c})
}

func createSetTapeWidthPacket(width uint32) (b [10]byte) {
	b[0], b[1] = 0x1b, 0x7b
	b[2] = 0x07
	b[3] = 0x4c

	binary.LittleEndian.PutUint32(b[4:8], width)

	// Checksum
	for _, v := range b[3:8] {
		b[8] += v
	}

	b[9] = 0x7d

	return b
}

func (t Tape) writePrintRequest(w io.Writer) {
	// https://github.com/hikalium/sr5900p/blob/aa1f3d42a5fbcf0a1ad2917e3bbb005196bf1bb9/src/print.rs#L82
	w.Write([]byte{0x1b, 0x7b, 0x03, 0x40, 0x40, 0x7d})
	w.Write([]byte{0x1b, 0x7b, 0x07, 0x7b, 0x00, 0x00, 0x53, 0x54, 0x22, 0x7d})
	w.Write([]byte{0x1b, 0x7b, 0x07, 0x43, 0x02, 0x02, 0x01, 0x01, 0x49, 0x7d})
	w.Write([]byte{0x1b, 0x7b, 0x04, 0x44, 0x05, 0x49, 0x7d})
	w.Write([]byte{0x1b, 0x7b, 0x03, 0x47, 0x47, 0x7d})
	b := createSetTapeWidthPacket(t.Length() + 4)
	w.Write(b[:])
	w.Write([]byte{0x1b, 0x7b, 0x05, 0x54, 0x2a, 0x00, 0x7e, 0x7d})
	w.Write([]byte{0x1b, 0x7b, 0x04, 0x48, 0x05, 0x4d, 0x7d})
	w.Write([]byte{0x1b, 0x7b, 0x04, 0x73, 0x00, 0x73, 0x7d})
	t.writeImage(w)
	w.Write([]byte{0x1b, 0x7b, 0x03, 0x40, 0x40, 0x7d})
}

func (t Tape) buildPrintRequest() *bytes.Buffer {
	b := new(bytes.Buffer)
	t.writePrintRequest(b)
	return b
}

func (t Tape) Print(addr string) (err error) {
	s, err := getStatus(addr)
	if err != nil {
		return fmt.Errorf("checking status: %w", err)
	}

	if !s.IsIdle() {
		return fmt.Errorf("not idle: %x", s)
	}
	if s.TapeType().Width() < t.Width() {
		return fmt.Errorf("print width is longer than %d (got %d)", s.TapeType().Width(), t.Width())
	}

	if err := lockPrinter(addr); err != nil {
		return fmt.Errorf("requesting lock: %w", err)
	}

	log.Printf("connecting tcp")
	c, err := net.DialTimeout("tcp4", addr, tcpDialTimeout)
	if err != nil {
		return fmt.Errorf("connecting tcp: %w", err)
	}
	log.Printf("connected")

	if err := startPrint(addr); err != nil {
		return fmt.Errorf("requesting start: %w", err)
	}

	log.Printf("sending print: %x", t.buildPrintRequest().Bytes())
	c.SetWriteDeadline(time.Now().Add(tcpWriteTimeout))
	if _, err := t.buildPrintRequest().WriteTo(c); err != nil {
		c.Close()
		return fmt.Errorf("writing print data: %w", err)
	}
	log.Printf("sended print data")

	log.Printf("closing tcp")
	if err := c.Close(); err != nil {
		return fmt.Errorf("closing tcp: %w", err)
	}
	log.Printf("closed tcp")

	time.Sleep(500 * time.Millisecond)

	printDeadline := time.Now().Add(printTimeout)
	for {
		if time.Now().After(printDeadline) {
			return fmt.Errorf("print take too long time")
		}
		time.Sleep(100 * time.Millisecond)
		s, err := getStatus(addr)
		if err != nil {
			return fmt.Errorf("waiting print finish: %w", err)
		}
		if !s.IsPrinting() {
			break
		}
	}

	if err := unlockPrinter(addr); err != nil {
		return fmt.Errorf("requesting unlock: %w", err)
	}

	return nil
}
