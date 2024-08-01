package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"strings"
)

// reverseString reverses the input string
func reverseString(input string) string {
	runes := []rune(input)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// handleConnection handles the incoming TCP connection
func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("Client disconnected")
			} else {
				log.Println("Error reading message:", err)
			}
			return
		}

		message = strings.TrimSpace(message)
		log.Println("Received message:", message)
		reversedMessage := reverseString(message)
		log.Println("Sending reversed message:", reversedMessage)

		conn.Write([]byte(reversedMessage + "\n"))
	}
}

func main() {
	listener, err := net.Listen("tcp", ":9101")
	if err != nil {
		log.Fatal("Error starting TCP server:", err)
	}
	defer listener.Close()
	log.Println("TCP server listening on port 9101")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}
