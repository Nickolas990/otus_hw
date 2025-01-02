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

	t.Run("large data transfer", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer l.Close()

		data := make([]byte, 20*20)
		for i := range data {
			data[i] = 'a'
		}

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()

			in := &bytes.Buffer{}
			out := &bytes.Buffer{}

			timeout, err := time.ParseDuration("30s") // увеличен тайм-аут
			require.NoError(t, err)

			client := NewTelnetClient(l.Addr().String(), timeout, io.NopCloser(in), out)
			require.NoError(t, err)

			require.NoError(t, client.Connect())
			defer client.Close()

			in.Write(data)
			err = client.Send()
			require.NoError(t, err)

			err = client.Receive()
			if err != nil {
				t.Errorf("client.Receive() error: %v", err)
				return
			}

			require.Equal(t, string(data), out.String())
		}()

		go func() {
			defer wg.Done()

			conn, err := l.Accept()
			require.NoError(t, err)
			defer conn.Close()

			request := make([]byte, 1024*1024)
			n, err := conn.Read(request)
			require.NoError(t, err)
			require.Equal(t, string(data), string(request)[:n])

			t.Log("Server received data successfully")

			n, err = conn.Write(data)
			require.NoError(t, err)
			require.NotEqual(t, 0, n)

			t.Log("Server sent data back successfully")

			// Добавим задержку перед закрытием соединения
			time.Sleep(1 * time.Second)
		}()

		wg.Wait()
	})
}
