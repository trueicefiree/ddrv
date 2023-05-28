package discord

import "errors"

var (
	ErrInvalidSeek   = errors.New("invalid seek offset")
	ErrAlreadyClosed = errors.New("already closed")
	ErrClosed        = errors.New("is closed")
)

// Chunk represents a part of a stream
type Chunk struct {
	URL   string // URL where the chunk is stored
	Size  int    // Size of the chunk
	Start int
	End   int
}
