package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"io"
	"log"
	"net"
	"time"
)

type Tape struct {
	image.Image
}

func (t Tape) Height() uint16 {
	return uint16(t.Bounds().Max.Y)
}

func (t Tape) Width() uint32 {
	return uint32(t.Bounds().Max.X)
}

func (t Tape) getPixel(x uint32, y uint16) byte {
	r, g, b, a := t.At(int(x), int(y)).RGBA()
	if r == 0 && g == 0 && b == 0 && a > 0 {
		return 0x01
	} else {
		return 0x00
	}
}

func (t Tape) writeImage(w io.Writer) {
	bytes := (t.Height() + 7) / 8
	for x := t.Width() - 1; x >= 0; x-- {
		w.Write([]byte{0x1b, 0x2e, 0x00, 0x00, 0x00, 0x01})
		binary.Write(w, binary.LittleEndian, t.Height())
		for i := uint16(0); i < bytes; i++ {
			var b byte
			for j := uint16(0); j < 8; j++ {
				y := i*8 + j
				if y < t.Height() {
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
	b[2] = 0x07 // length
	b[3] = 0x4c // 76

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
	b := createSetTapeWidthPacket(t.Width() + 4)
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

func (t Tape) Print(addr string) error {
	s, err := getStatus(addr)
	if err != nil {
		return fmt.Errorf("checking status: %w", err)
	}

	if !s.IsIdle() {
		return fmt.Errorf("not idle: %x", s)
	}

	if err := lockPrinter(addr); err != nil {
		return fmt.Errorf("requesting lock: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	log.Printf("connecting tcp")
	c, err := net.Dial("tcp4", addr)
	if err != nil {
		return fmt.Errorf("connecting tcp: %w", err)
	}
	log.Printf("connected")
	time.Sleep(500 * time.Millisecond)

	if err := startPrint(addr); err != nil {
		return fmt.Errorf("requesting start: %w", err)
	}
	time.Sleep(500 * time.Millisecond)
	log.Printf("sending print: %x", t.buildPrintRequest().Bytes())
	if _, err := t.buildPrintRequest().WriteTo(c); err != nil {
		c.Close()
		return fmt.Errorf("writing print data: %w", err)
	}
	time.Sleep(500 * time.Millisecond)
	log.Printf("closing tcp")
	if err := c.Close(); err != nil {
		return fmt.Errorf("closing tcp: %w", err)
	}
	time.Sleep(500 * time.Millisecond)

	if err := startPrint2(addr); err != nil {
		return fmt.Errorf("requesting start2: %w", err)
	}
	time.Sleep(500 * time.Millisecond)

	for { // TODO: Set timeout
		s, err := getStatus(addr)
		if err != nil {
			return fmt.Errorf("waiting print finish: %w", err)
		}
		time.Sleep(500 * time.Millisecond)
		if !s.IsPrinting() {
			break
		}
	}

	//if err := unlockPrinter(addr); err != nil {
	//        return fmt.Errorf("requesting unlock: %w", err)
	//}

	return nil
}
