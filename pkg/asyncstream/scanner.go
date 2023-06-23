package asyncstream

import (
	"io"
)

// Scanner is a structure that provides a buffered read functionality
// from an io.Reader. It also keeps track of the first non-EOF error
// that occurred during reading.
type Scanner struct {
	r   io.Reader // The io.Reader from which the Scanner reads.
	buf []byte    // The buffer that holds the data read from the Reader.
	err error     // The first non-EOF error that was encountered by the Scanner.
}

// NewScanner initializes and returns a new Scanner that reads from 'r' and
// uses a buffer of 'size' bytes.
func NewScanner(r io.Reader, size int) *Scanner {
	return &Scanner{
		r:   r,
		buf: make([]byte, size),
	}
}

// Scan reads the next chunk of data from the io.Reader into the buffer.
// It returns false if an error occurred during reading, otherwise true.
func (s *Scanner) Scan() bool {
	n, err := s.r.Read(s.buf) // Read data from the reader into the buffer.
	s.buf = s.buf[:n]         // Slice the buffer to the actual number of bytes read.
	if err == io.EOF && n > 0 {
		s.setErr(err) // record the error
		return true
	}
	if err != nil { // If an error occurred,
		s.setErr(err) // record the error
		return false  // and return false.
	}
	return true // Return true, indicating the read was successful.
}

// Bytes returns the most recent data read by the Scanner.
func (s *Scanner) Bytes() []byte {
	return s.buf
}

// String returns the most recent data read by the Scanner as a string.
func (s *Scanner) String() string {
	return string(s.buf)
}

// Err returns the first non-EOF error that was encountered by the Scanner.
// If the only error encountered was io.EOF, it returns nil, as io.EOF is
// expected when reaching the end of the input.
func (s *Scanner) Err() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}

// setErr records the first non-EOF error encountered by the Scanner.
// If no error has been recorded yet or the only error encountered was
// io.EOF, it sets the error field to 'err'.
func (s *Scanner) setErr(err error) {
	if s.err == nil || s.err == io.EOF {
		s.err = err
	}
}
