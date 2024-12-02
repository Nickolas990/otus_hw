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
	address string
	timeout time.Duration
	conn    net.Conn
	in      io.ReadCloser
	out     io.Writer
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &telnetClient{address: address, timeout: timeout, in: in, out: out}
}

func (c *telnetClient) Connect() error {
	conn, err := net.DialTimeout("tcp", c.address, c.timeout)
	if err != nil {
		return err
	}
	c.conn = conn
	// _, err = fmt.Fprintln(c.out, "...Connected to", c.address)
	// if err != nil {
	//	return err
	//}
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
	return err
}

func (c *telnetClient) Receive() error {
	buffer := make([]byte, 1024)
	for {
		n, err := c.conn.Read(buffer)
		if n > 0 {
			if _, writeErr := c.out.Write(buffer[:n]); writeErr != nil {
				return fmt.Errorf("failed to write data to output: %w", writeErr)
			}
		}
		if err != nil {
			if err == io.EOF {
				fmt.Println("EOF encountered in Receive")
				return nil
			}
			return fmt.Errorf("failed to receive data: %w", err)
		}
	}
}

// Place your code here.
// P.S. Author's solution takes no more than 50 lines.
