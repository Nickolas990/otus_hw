package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

func main() {
	var timeout time.Duration
	flag.DurationVar(&timeout, "timeout", 10*time.Second, "connection timeout")
	flag.Parse()

	if flag.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "Usage: go-telnet --timeout=10s host port")
		os.Exit(1)
	}

	host := flag.Arg(0)
	port := flag.Arg(1)
	address := net.JoinHostPort(host, port)

	client := NewTelnetClient(address, timeout, os.Stdin, os.Stdout)

	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer func(client TelnetClient) {
		err := client.Close()
		if err != nil {
			return
		}
	}(client)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		<-ctx.Done()
		err := client.Close()
		if err != nil {
			return
		}
	}()

	errCh := make(chan error, 1)

	go func() {
		errCh <- client.Send()
	}()

	go func() {
		errCh <- client.Receive()
	}()

	select {
	case <-ctx.Done():
		fmt.Fprintln(os.Stderr, "...Operation canceled")
	case err := <-errCh:
		if err != nil {
			fmt.Fprintf(os.Stderr, "...Connection error: %v\n", err)
		}
	}
}
