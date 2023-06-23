package ddrv

import (
	"bytes"
	"io"
	"sort"
	"sync"

	"github.com/forscht/ddrv/pkg/asyncstream"
)

// AsyncWriter implements io.WriteCloser.
// It streams data in chunks to Discord server channels using webhook
// AsyncWriter buffers bytes into memory and writes data to discord in parallel
type AsyncWriter struct {
	mgr       *Manager // Manager where Writer writes data
	chunkSize int      // The maximum Size of a chunk
	onChunk   func(chunk *Attachment)
	cCounter  int64
	mu        sync.Mutex
	wg        sync.WaitGroup
	closed    bool // Whether the Writer has been closed
	err       error
	chunks    []*attachmentWithIdx
	pwriter   *io.PipeWriter // PipeWriter for writing the current chunk
	errCh     chan error
}

type attachmentWithIdx struct {
	idx        int
	attachment *Attachment
}

func NewAsyncWriter(onChunk func(chunk *Attachment), chunkSize int, mgr *Manager) io.WriteCloser {
	sw := &AsyncWriter{
		mgr:       mgr,
		onChunk:   onChunk,
		chunkSize: chunkSize,
		errCh:     make(chan error),
	}
	return sw
}

func (w *AsyncWriter) Write(p []byte) (int, error) {
	if w.closed {
		return 0, ErrClosed
	}
	w.mu.Lock()
	if w.pwriter == nil {
		reader, writer := io.Pipe()
		w.pwriter = writer
		go w.startWorkers(&BlockingReader{reader})
	}
	w.mu.Unlock()
	if w.err != nil {
		return 0, w.err
	}
	return w.pwriter.Write(p)
}

func (w *AsyncWriter) Close() error {
	if w.closed {
		return ErrAlreadyClosed
	}
	w.closed = true
	if w.pwriter != nil {
		if err := w.pwriter.Close(); err != nil {
			return err
		}
	}
	<-w.errCh
	if w.onChunk != nil {
		sort.SliceStable(w.chunks, func(i, j int) bool {
			return w.chunks[i].idx < w.chunks[j].idx
		})
		for _, chunk := range w.chunks {
			w.onChunk(chunk.attachment)
		}
	}
	return nil
}

func (w *AsyncWriter) startWorkers(reader io.Reader) {
	stream := asyncstream.New(4, w.chunkSize)
	w.errCh <- stream.Process(&BlockingReader{reader}, func(data []byte, idx int) error {
		a, err := w.mgr.write(bytes.NewReader(data))
		if err != nil {
			return err
		}
		w.mu.Lock()
		w.chunks = append(w.chunks, &attachmentWithIdx{
			idx:        idx,
			attachment: a,
		})
		w.mu.Unlock()
		return nil
	})
}

type BlockingReader struct {
	reader io.Reader
}

func (br *BlockingReader) Read(p []byte) (int, error) {
	currReadIdx := 0
	for currReadIdx < len(p) {
		n, err := br.reader.Read(p[currReadIdx:])
		currReadIdx += n
		if err != nil {
			return currReadIdx, err
		}
	}
	return currReadIdx, nil
}
