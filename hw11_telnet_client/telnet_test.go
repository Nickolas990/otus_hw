package main

import (
	"bytes"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	//nolint:depguard
	"github.com/stretchr/testify/require"
)

func TestTelnetClient(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")

		require.NoError(t, err)

		defer func() { require.NoError(t, l.Close()) }()

		var wg sync.WaitGroup

		wg.Add(2)

		go func() {
			defer wg.Done()

			in := &bytes.Buffer{}

			out := &bytes.Buffer{}

			timeout, err := time.ParseDuration("10s")

			require.NoError(t, err)

			client := NewTelnetClient(l.Addr().String(), timeout, io.NopCloser(in), out)

			require.NoError(t, client.Connect())

			defer func() { require.NoError(t, client.Close()) }()

			in.WriteString("hello\n")

			err = client.Send()

			require.NoError(t, err)

			err = client.Receive()

			require.NoError(t, err)

			require.Equal(t, "world\n", out.String())
		}()

		go func() {
			defer wg.Done()

			conn, err := l.Accept()

			require.NoError(t, err)

			require.NotNil(t, conn)

			defer func() { require.NoError(t, conn.Close()) }()

			request := make([]byte, 1024)

			n, err := conn.Read(request)

			require.NoError(t, err)

			require.Equal(t, "hello\n", string(request)[:n])

			n, err = conn.Write([]byte("world\n"))

			require.NoError(t, err)

			require.NotEqual(t, 0, n)
		}()

		wg.Wait()
	})
}

func TestTelnetClientAdditionalScenarios(t *testing.T) {
	t.Run("connection timeout", func(t *testing.T) {
		timeout, err := time.ParseDuration("1ms")
		require.NoError(t, err)

		client := NewTelnetClient("192.0.2.0:1234", timeout, nil, nil)
		err = client.Connect()
		require.Error(t, err)
	})

	t.Run("server disconnect", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer l.Close()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()

			in := &bytes.Buffer{}
			out := &bytes.Buffer{}

			timeout, err := time.ParseDuration("10s")
			require.NoError(t, err)

			client := NewTelnetClient(l.Addr().String(), timeout, io.NopCloser(in), out)
			require.NoError(t, client.Connect())
			defer client.Close()

			in.WriteString("hello\n")
			err = client.Send()
			require.NoError(t, err)

			// Attempt to receive data
			for i := 0; i < 5; i++ {
				err = client.Receive()
				t.Logf("Receive attempt %d error: %v", i+1, err)
				if err != nil {
					require.Equal(t, io.EOF, err)
					return
				}
				time.Sleep(100 * time.Millisecond)
			}
			t.Error("Expected error but got nil after multiple attempts")
		}()

		go func() {
			defer wg.Done()

			conn, err := l.Accept()
			require.NoError(t, err)
			defer conn.Close()

			request := make([]byte, 1024)
			n, err := conn.Read(request)
			require.NoError(t, err)
			require.Equal(t, "hello\n", string(request)[:n])

			// Simulate server disconnect with a slight delay to ensure client attempts to read
			time.Sleep(100 * time.Millisecond)
			err = conn.Close()
			require.NoError(t, err)
		}()

		wg.Wait()
	})
}
