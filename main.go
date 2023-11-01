package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
)

func loadImage() (image.Image, error) {
	f, err := os.Open("test.png")
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	return png.Decode(f)
}
func main() {
	img, err := loadImage()
	if err != nil {
		panic(err)
	}
	tape := Tape{img}

	f, err := os.Create("tcp-data.bin")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	tape.writePrintRequest(f)

	addr := os.Args[1]

	s, err := getStatus(addr)
	if err != nil {
		panic(err)
	}
	log.Printf("Printing: %v", s.IsPrinting())
	log.Printf("Idle: %v", s.IsIdle())
	log.Printf("Type: %v", s.TapeType())

	err = tape.Print(addr)
	if err != nil {
		panic(err)
	}
}
