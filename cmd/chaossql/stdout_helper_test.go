package main

import (
	"bytes"
	"io"
	"net"
	"os"
	"testing"
)

// captureStdout runs fn while redirecting os.Stdout, for commands that print
// with fmt.Print* instead of the cobra writer.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = w
	done := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()
	defer func() { os.Stdout = original }()
	fn()
	w.Close()
	os.Stdout = original
	return <-done
}

// freeLocalAddr reserves an ephemeral loopback port and releases it.
func freeLocalAddr(t *testing.T) (string, int) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	return ln.Addr().String(), ln.Addr().(*net.TCPAddr).Port
}
