package discord

import "errors"

// ErrClosed is returned when writer or reader is
// closed and caller is trying to Read or Write
var ErrClosed = errors.New("is closed")

// ErrAlreadyClosed is returned when stream is already closed
var ErrAlreadyClosed = errors.New("already closed")
