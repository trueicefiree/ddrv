package bufcp

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestCopy(t *testing.T) {
	srcStr := "Hello, World!"
	src := strings.NewReader(srcStr)
	dst := bufio.NewWriter(bytes.NewBuffer(nil))
	bufSize := 8

	expected := int64(len(srcStr))
	written, err := Copy(dst, src, bufSize)
	if err != nil {
		t.Fatal(err)
	}

	if written != expected {
		t.Fatalf("expected to write %d bytes, but wrote %d bytes", expected, written)
	}
}

func TestCopyN(t *testing.T) {
	srcStr := "Hello, World!"
	src := strings.NewReader(srcStr)
	dst := bufio.NewWriter(bytes.NewBuffer(nil))
	bufSize := 8
	n := int64(5)

	written, err := CopyN(dst, src, n, bufSize)
	if err != nil {
		t.Fatal(err)
	}

	if written != n {
		t.Fatalf("expected to write %d bytes, but wrote %d bytes", n, written)
	}
}
