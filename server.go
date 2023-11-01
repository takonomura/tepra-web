package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Server struct {
	Address string

	mu sync.Mutex
}

func (s *Server) Print(img image.Image) error {
	if !s.mu.TryLock() {
		return fmt.Errorf("another printing is in progress; please try again later")
	}
	defer s.mu.Unlock()

	t := Tape{Image: img}
	return t.Print(s.Address)
}

func writeJSONMessage(w http.ResponseWriter, code int, msg string) {
	body := struct {
		Message string `json:"message"`
	}{
		Message: msg,
	}
	b, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))
	w.WriteHeader(code)
	w.Write(b)
}

func (s *Server) handleNotFound(w http.ResponseWriter, r *http.Request) {
	writeJSONMessage(w, http.StatusNotFound, "not found")
}

//go:embed index.html
var indexHTML []byte

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		writeJSONMessage(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(indexHTML)))
	w.WriteHeader(http.StatusOK)
	w.Write(indexHTML)
}

func (s *Server) handlePrint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONMessage(w, http.StatusMethodNotAllowed, "unexpected http method")
		return
	}

	f, _, err := r.FormFile("tape")
	if err != nil {
		log.Printf("failed to parse request body: %+v", err)
		writeJSONMessage(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		log.Printf("failed to parse png: %+v", err)
		writeJSONMessage(w, http.StatusBadRequest, "invalid image")
		return
	}

	err = s.Print(img)
	if err != nil {
		log.Printf("failed to print: %+v", err)
		writeJSONMessage(w, http.StatusInternalServerError, fmt.Sprintf("failed to print: %v", err))
		return
	}

	writeJSONMessage(w, http.StatusOK, "printed successfully")
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		s.handleIndex(w, r)
	case "/print":
		s.handlePrint(w, r)
	default:
		s.handleNotFound(w, r)
	}
}
