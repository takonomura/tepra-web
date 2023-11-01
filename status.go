package main

import (
	"bytes"
	"fmt"
)

type TapeType uint8

const (
	TapeTypeUnknown TapeType = iota
	TapeTypeWidth6mm
	TapeTypeWidth9mm
	TapeTypeWidth12mm
	TapeTypeWidth18mm
	TapeTypeWidth24mm
	TapeTypeWidth36mm
	TapeTypeWidth4mm
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
	case TapeTypeWidth18mm:
		return "TapeTypeWidth18mm"
	case TapeTypeWidth24mm:
		return "TapeTypeWidth24mm"
	case TapeTypeWidth36mm:
		return "TapeTypeWidth36mm"
	case TapeTypeWidth4mm:
		return "TapeTypeWidth4mm"
	default:
		return fmt.Sprintf("TapeType(%d)", t)
	}
}

func (t TapeType) Width() uint16 {
	switch t {
	case TapeTypeWidth6mm:
		return 72
	case TapeTypeWidth9mm:
		return 108
	case TapeTypeWidth12mm:
		return 144
	case TapeTypeWidth18mm:
		return 216
	case TapeTypeWidth24mm:
		return 288
	case TapeTypeWidth36mm:
		return 384
	case TapeTypeWidth4mm:
		return 384
	default:
		return 0
	}
}

type Status [20]byte

func (s Status) TapeType() TapeType {
	return TapeType(s[3])
}

func (s Status) IsPrinting() bool {
	return s[1] == 0x02
}

func (s Status) IsIdle() bool {
	masked := s
	masked[3] = 0xFF
	masked[13] |= 0x01
	return bytes.Equal(masked[:], []byte{
		0x14, 0x00, 0x00, 0xff,
		0x00, 0x00, 0x00, 0x00,
		0x40, 0x00, 0x00, 0x00,
		0x00, 0x01, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	})
}
