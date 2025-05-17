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
			log.Fatal(err)
		}

		rLine := fmt.Sprintf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n",
			r.RequestLine.Method, r.RequestLine.RequestTarget, r.RequestLine.HttpVersion)
		headers := "Headers:\n"
		for k, v := range r.Headers {
			headers += fmt.Sprintf("- %s: %s\n", k, v)
		}

		fmt.Print(rLine + headers)
	}
}
