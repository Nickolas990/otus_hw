package main

import (
	"bufio"
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
	scanner := bufio.NewScanner(c.in)
	for scanner.Scan() {
		if _, err := c.conn.Write(append(scanner.Bytes(), '\n')); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func (c *telnetClient) Receive() error {
	_, err := io.Copy(c.out, c.conn)
	if err != nil {
		return err
	}
	return nil
}

// Place your code here.
// P.S. Author's solution takes no more than 50 lines.
