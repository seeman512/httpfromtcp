package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:42069")
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Connection Accepted")

		r, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal()
		}

		fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n",
			r.RequestLine.Method, r.RequestLine.RequestTarget, r.RequestLine.HttpVersion)
	}
}
