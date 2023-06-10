// Package bufcp provides functions for copying data from a reader to a writer
// with explicit control over buffer size and flushing of buffered data after each write.
// This package is particularly useful when working with the CreateWriteStream function of FastHTTP's Response,
// which returns a bufio.Writer. By flushing the writer after each write, data is sent to the client immediately,
// allowing the server to start sending large responses without having to buffer them entirely in memory first.
// This can help manage memory usage when working with large files or streams of data.
package bufcp

import (
	"bufio"
	"io"
)

// Copy copies from src to dst until either EOF is reached on src or an error occurs.
// It uses a buffer of size bufSize and flushes the buffer after each writing.
func Copy(dst *bufio.Writer, src io.Reader, bufSize int) (int64, error) {
	buf := make([]byte, bufSize)
	var written int64
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				return written, ew
			}
			if nr != nw {
				return written, io.ErrShortWrite
			}
			if err := dst.Flush(); err != nil {
				return written, err
			}
		}
		if er != nil {
			if er == io.EOF {
				return written, nil
			}
			return written, er
		}
	}
}

// CopyN copies n bytes (or until an error occurs) from src to dst.
// It uses a buffer of size bufSize and flushes the buffer after each writing.
func CopyN(dst *bufio.Writer, src io.Reader, n int64, bufSize int) (int64, error) {
	buf := make([]byte, bufSize)
	var written int64
	for written < n {
		toRead := int64(len(buf))
		if remaining := n - written; remaining < toRead {
			toRead = remaining
		}
		nr, er := src.Read(buf[:toRead])
		if nr > 0 {
			nw, ew := dst.Write(buf[:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				return written, ew
			}
			if nr != nw {
				return written, io.ErrShortWrite
			}
			if err := dst.Flush(); err != nil {
				return written, err
			}
		}
		if er != nil {
			if er == io.EOF {
				return written, nil
			}
			return written, er
		}
	}
	return written, nil
}
