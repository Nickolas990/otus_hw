package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

func main() {
	// Пример использования telnet клиента
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <address>\n", os.Args[0])
		os.Exit(1)
	}

	address := os.Args[1]
	client := NewTelnetClient(address, 10*time.Second, os.Stdin, os.Stdout)

	err := client.Connect()
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	go func() {
		for {
			err := client.Receive()
			if err != nil {
				if err == io.EOF {
					fmt.Println("Connection closed by server")
					return
				}
				log.Fatalf("Receive error: %v", err)
			}
		}
	}()

	err = client.Send()
	if err != nil {
		log.Fatalf("Send error: %v", err)
	}
}
