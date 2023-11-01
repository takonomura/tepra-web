package main

import "image"

type Tape struct {
	image.Image
}

func (t Tape) Width() uint16 {
	return uint16(t.Bounds().Max.Y)
}

func (t Tape) Length() uint32 {
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
