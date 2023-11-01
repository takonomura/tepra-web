package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
)

const headerSize = 32

var (
	udpDialTimeout  = 1 * time.Second
	udpWriteTimeout = 1 * time.Second
	udpReadTimeout  = 1 * time.Second
)

func doUDP(addr string, reqPayload []byte) (respHeader []byte, respBody []byte, err error) {
	conn, err := net.DialTimeout("udp4", addr, udpDialTimeout)
	if err != nil {
		return nil, nil, fmt.Errorf("dial udp: %w", err)
	}
	defer conn.Close()

	conn.SetWriteDeadline(time.Now().Add(udpWriteTimeout))
	if _, err := conn.Write(reqPayload); err != nil {
		return nil, nil, fmt.Errorf("send udp request: %w", err)
	}

	resp := make([]byte, 1500)

	conn.SetReadDeadline(time.Now().Add(udpReadTimeout))
	n, err := conn.Read(resp)
	if err != nil {
		return nil, nil, fmt.Errorf("receive udp response: %w", err)
	}

	if n < headerSize {
		return nil, nil, fmt.Errorf("too small response: %x", resp[:n])
	}
	return resp[:headerSize], resp[headerSize:n], nil
}

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

func getStatus(addr string) (Status, error) {
	req := make([]byte, headerSize)
	putUDPHeader(req, 0x01, 0)
	log.Printf("sending status request: %x", req)
	resph, resp, err := doUDP(addr, req)
	if err != nil {
		return Status{}, err
	}
	log.Printf("received status response: %x %x", resph, resp)
	if len(resp) != 20 {
		return Status{}, fmt.Errorf("invalid response length %d: %x", len(resp), resp)
	}
	var s Status
	copy(s[:], resp)
	return s, nil
}

func lockPrinter(addr string) error {
	req := make([]byte, headerSize)
	putUDPHeader(req, 0x02, 0)
	log.Printf("sending lock request: %x", req)
	resph, resp, err := doUDP(addr, req)
	if err != nil {
		return err
	}
	log.Printf("received response for lock request: %x %x", resph, resp)
	if !bytes.Equal(resp, []byte{0x02, 0x00, 0x00}) {
		return fmt.Errorf("invalid response: %x %x", resph, resp)
	}
	return nil
}

func unlockPrinter(addr string) error {
	req := make([]byte, headerSize)
	putUDPHeader(req, 0x03, 0)
	log.Printf("sending unlock request: %x", req)
	resph, resp, err := doUDP(addr, req)
	if err != nil {
		return err
	}
	log.Printf("received response for unlock request: %x %x", resph, resp)
	if !bytes.Equal(resp, []byte{0x03, 0x00, 0x00}) {
		return fmt.Errorf("invalid response: %x %x", resph, resp)
	}
	return nil
}

func startPrint(addr string) error {
	req := make([]byte, headerSize)
	putUDPHeader(req, 0x0101, 0)
	log.Printf("sending start request: %x", req)
	resph, resp, err := doUDP(addr, req)
	if err != nil {
		return err
	}
	log.Printf("received response for start request: %x %x", resph, resp)
	if len(resp) != 0 {
		return fmt.Errorf("invalid response: %x %x", resph, resp)
	}
	return nil
}

//func startPrint2(addr string) error {
//        req := make([]byte, headerSize)
//        putUDPHeader(req, 0x0100, 0)
//        log.Printf("sending start request2: %x", req)
//        resp, err := doUDP(addr, req)
//        if err != nil {
//                return err
//        }
//        log.Printf("received response for start request2: %x", resp)
//        return nil
//}
