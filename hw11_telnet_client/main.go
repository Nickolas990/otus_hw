package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
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
	client := NewTelnetClient(net.JoinHostPort(flag.Arg(0), flag.Arg(1)), timeout, os.Stdin, os.Stdout)
	if err := client.Connect(); err != nil {
		fmt.Fprintln(os.Stderr, "...Connection error:", err)
		os.Exit(1)
	}
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		cancel()
		client.Close()
		fmt.Fprintln(os.Stderr, "...EOF")
		os.Exit(0)
	}()

	go func() {
		if err := client.Send(); err != nil {
			fmt.Fprintln(os.Stderr, "...Send error:", err)
			cancel()
			client.Close()
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, "...EOF")
		cancel()
		client.Close()
		os.Exit(0)
	}()

	done := make(chan error, 1)
	go func() {
		done <- client.Receive()
	}()

	select {
	case <-ctx.Done():
		fmt.Fprintln(os.Stderr, "...Operation canceled")
	case err := <-done:
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Fprintln(os.Stderr, "...Connection was closed by peer")
			} else {
				fmt.Fprintln(os.Stderr, "...Receive error:", err)
			}
		} else {
			fmt.Fprintln(os.Stderr, "...Connection was closed by peer")
		}
	}
}
