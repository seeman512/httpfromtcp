package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		str, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Read error: %v\n", err)
			continue
		}

		_, err = conn.Write([]byte(str))
		if err != nil {
			fmt.Printf("Write error: %v\n", err)
			continue
		}
	}
}
