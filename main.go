package main

import (
	"log"
	"net"
	"net/http"
	"os"
)

func main() {
	tepraAddress := os.Getenv("TEPRA_ADDRESS")
	if tepraAddress == "" {
		log.Fatal("TEPRA_ADDRESS is not specified")
	}
	if _, _, err := net.SplitHostPort(tepraAddress); err != nil {
		log.Fatalf("TEPRA_ADDRESS is not a valid address: %v", err)
	}

	listenAddress := os.Getenv("LISTEN_ADDRESS")
	if listenAddress != "" {
		log.Printf("LISTEN_ADDRESS is not specified")
	}

	s := &Server{
		Address: tepraAddress,
	}
	log.Fatal(http.ListenAndServe(listenAddress, s))
}
