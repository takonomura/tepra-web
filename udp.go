package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

const headerSize = 32

func putUDPHeader(b []byte, cmd uint32, length uint32) {
	copy(b[0:16], []byte{
		0x54, 0x50, 0x52, 0x54,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x20,
	})
	binary.BigEndian.PutUint32(b[16:20], cmd)
	binary.BigEndian.PutUint32(b[20:24], length)
}

func doUDP(addr string, reqPayload []byte) ([]byte, error) {
	conn, err := net.Dial("udp4", addr)
	if err != nil {
		return nil, fmt.Errorf("dial udp: %w", err)
	}
	defer conn.Close()

	if _, err := conn.Write(reqPayload); err != nil {
		return nil, fmt.Errorf("send udp request: %w", err)
	}

	resp := make([]byte, 1500)
	n, err := conn.Read(resp)
	if err != nil {
		return nil, fmt.Errorf("receive udp response: %w", err)
	}
	return resp[:n], nil
}

type Status [20]byte

type TapeType uint8

const (
	TapeTypeUnknown TapeType = iota
	TapeTypeWidth6mm
	TapeTypeWidth9mm
	TapeTypeWidth12mm
	TapeTypeWidth24mm
	TapeTypeWidth36mm
)

func (t TapeType) String() string {
	switch t {
	case TapeTypeUnknown:
		return "TapeTypeUnknown"
	case TapeTypeWidth6mm:
		return "TapeTypeWidth6mm"
	case TapeTypeWidth9mm:
		return "TapeTypeWidth9mm"
	case TapeTypeWidth12mm:
		return "TapeTypeWidth12mm"
	case TapeTypeWidth24mm:
		return "TapeTypeWidth24mm"
	case TapeTypeWidth36mm:
		return "TapeTypeWidth36mm"
	default:
		return fmt.Sprintf("TapeType(%d)", t)
	}
}

func (s Status) TapeType() TapeType {
	return TapeType(s[3])
}

func (s Status) IsPrinting() bool {
	return s[1] == 0x02
}

func (s Status) IsIdle() bool {
	return s[0] == 0x14 && s[1] == 0x00 && s[2] == 0x00 && s[4] == 0x00 && s[5] == 0x00 && s[6] == 0x00 && s[7] == 0x00 && s[8] == 0x40 && s[9] == 0x00 && s[10] == 0x00 && s[11] == 0x00 && s[12] == 0x00 && s[14] == 0x00 && s[15] == 0x00 && s[16] == 0x00 && s[17] == 0x00 && s[18] == 0x00 && s[19] == 0x00
}

func getStatus(addr string) (Status, error) {
	req := make([]byte, headerSize)
	putUDPHeader(req, 0x01, 0)
	log.Printf("sending status request: %x", req)
	resp, err := doUDP(addr, req)
	if err != nil {
		return Status{}, err
	}
	log.Printf("received status response: %x", resp)
	if len(resp) != (headerSize + 20) {
		return Status{}, fmt.Errorf("invalid response length %d: %x", len(resp), resp)
	}
	var s Status
	copy(s[:], resp[headerSize:headerSize+20])
	return s, nil
}

func lockPrinter(addr string) error {
	req := make([]byte, headerSize)
	putUDPHeader(req, 0x02, 0)
	log.Printf("sending lock request: %x", req)
	resp, err := doUDP(addr, req)
	if err != nil {
		return err
	}
	log.Printf("received response for lock request: %x", resp)
	return nil
}

func unlockPrinter(addr string) error {
	req := make([]byte, headerSize)
	putUDPHeader(req, 0x03, 0)
	log.Printf("sending unlock request: %x", req)
	resp, err := doUDP(addr, req)
	if err != nil {
		return err
	}
	log.Printf("received response for unlock request: %x", resp)
	return nil
}

func startPrint(addr string) error {
	req := make([]byte, headerSize)
	putUDPHeader(req, 0x0101, 0)
	log.Printf("sending start request: %x", req)
	resp, err := doUDP(addr, req)
	if err != nil {
		return err
	}
	log.Printf("received response for start request: %x", resp)
	return nil
}

func startPrint2(addr string) error {
	req := make([]byte, headerSize)
	putUDPHeader(req, 0x0100, 0)
	log.Printf("sending start request2: %x", req)
	resp, err := doUDP(addr, req)
	if err != nil {
		return err
	}
	log.Printf("received response for start request2: %x", resp)
	return nil
}
