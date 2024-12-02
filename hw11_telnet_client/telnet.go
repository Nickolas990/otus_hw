package main

import (
	"fmt"
	"io"
	"net"
	"time"
)

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

type telnetClient struct {
	conn    net.Conn
	address string
	timeout time.Duration
	in      io.Reader
	out     io.Writer
}

func NewTelnetClient(address string, timeout time.Duration, in io.Reader, out io.Writer) TelnetClient {
	return &telnetClient{
		address: address,
		timeout: timeout,
		in:      in,
		out:     out,
	}
}

func (c *telnetClient) Connect() error {
	conn, err := net.DialTimeout("tcp", c.address, c.timeout)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *telnetClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *telnetClient) Send() error {
	_, err := io.Copy(c.conn, c.in)
	if err != nil {
		return fmt.Errorf("failed to send data: %w", err)
	}
	return nil
}

func (c *telnetClient) Receive() error {
	_, err := io.Copy(c.out, c.conn)
	if err != nil {
		return fmt.Errorf("failed to receive data: %w", err)
	}
	return nil
}
