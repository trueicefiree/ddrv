package ddrv

import (
	"io"
)

// Writer implements io.WriteCloser.
// It streams data in chunks to Discord server channels using webhook
type Writer struct {
	mgr       *Manager // Manager where Writer writes data
	chunkSize int      // The maximum Size of a chunk
	onChunk   func(chunk *Attachment)

	idx     int              // Current position in the current chunk
	closed  bool             // Whether the Writer has been closed
	errCh   chan error       // Channel to send any errors that occur during writing
	chunkCh chan *Attachment // Channel to send chunks after they're written
	pwriter *io.PipeWriter   // PipeWriter for writing the current chunk
}

// NewWriter creates a new Writer with the given chunk Size and manager.
func NewWriter(onChunk func(chunk *Attachment), chunkSize int, mgr *Manager) io.WriteCloser {
	w := &Writer{
		mgr:       mgr,
		errCh:     make(chan error, 0),
		chunkCh:   make(chan *Attachment, 0),
		onChunk:   onChunk,
		chunkSize: chunkSize,
	}
	return w
}

// Write implements the Write method of io.Writer. It writes p to the Writer.
// If p is larger than the remaining space in the current chunk, Write splits p across
// multiple chunks as needed. Returns the total number of bytes from p that were written.
func (w *Writer) Write(p []byte) (int, error) {
	if w.closed {
		return 0, ErrClosed
	}
	if w.pwriter == nil {
		w.next()
	}
	total := len(p)
	for len(p) > 0 {
		if w.idx+len(p) > w.chunkSize {
			n, err := w.pwriter.Write(p[:w.chunkSize-w.idx])
			if err != nil {
				return total, err
			}
			if err := w.flush(true); err != nil {
				return total, err
			}
			p = p[n:]
		} else {
			n, err := w.pwriter.Write(p)
			if err != nil {
				return total, err
			}
			w.idx += n
			p = p[n:]
		}
	}
	return total, nil
}

// Close implements the Close method of io.Closer. It closes the Writer.
// If the Writer is already closed, Close returns ErrAlreadyClosed.
func (w *Writer) Close() error {
	if w.closed {
		return ErrAlreadyClosed
	}
	w.closed = true
	return w.flush(false)
}

// flush closes the current chunk, waits for it to be written to storage,
// and starts a new chunk if next is true.
func (w *Writer) flush(next bool) error {
	if err := w.pwriter.Close(); err != nil {
		return err
	}
	select {
	case err := <-w.errCh:
		return err
	case chunk := <-w.chunkCh:
		if w.onChunk != nil {
			w.onChunk(chunk)
		}
	}
	if next {
		w.next()
	}
	return nil
}

// next starts a new chunk for writing.
func (w *Writer) next() {
	if !w.closed {
		reader, writer := io.Pipe()
		w.pwriter = writer
		go func() {
			chunk, err := w.mgr.write(reader)
			if err != nil {
				w.errCh <- err
			} else {
				w.idx = 0
				w.chunkCh <- chunk
			}
		}()
	}
}
