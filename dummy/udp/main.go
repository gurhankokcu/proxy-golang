package main

import (
	"log"
	"net"
)

// reverseString reverses the input string
func reverseString(input string) string {
	runes := []rune(input)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func main() {
	// Listen on UDP port 9102
	addr := net.UDPAddr{
		Port: 9102,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatal("Error starting UDP server:", err)
	}
	defer conn.Close()
	log.Println("UDP server listening on port 9102")

	buffer := make([]byte, 1024)

	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Println("Error reading from UDP:", err)
			continue
		}

		message := string(buffer[:n])
		log.Println("Received message:", message)

		reversedMessage := reverseString(message)
		log.Println("Sending reversed message:", reversedMessage)

		_, err = conn.WriteToUDP([]byte(reversedMessage), clientAddr)
		if err != nil {
			log.Println("Error sending to UDP:", err)
		}
	}
}
