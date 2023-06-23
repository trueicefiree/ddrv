package asyncstream

import (
	"context"
	"io"
	"sync"
)

// Processor is a function type that processes a chunk of data and
// returns an error if processing failed.
type Processor func([]byte, int) error

// chunk is a struct that holds a chunk of data and its start and end
// positions in the original data stream.
type chunk struct {
	buf []byte // The chunk of data.
	idx int
}

var bufferPool sync.Pool

// AsyncStream is a struct that reads data from a stream, divides
// it into chunks, and processes the chunks concurrently.
type AsyncStream struct {
	mu           sync.Mutex // Mutex to synchronize access to shared state.
	conc         int        // The number of worker goroutines to use for processing.
	csize        int        // The size of each data chunk to read.
	chunkCounter int32      // Atomic counter for chunk index.
}

// New creates a new AsyncStream with the specified
// number of workers and chunk size.
func New(concurrency int, chunkSize int) *AsyncStream {
	bufferPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, chunkSize)
		},
	}

	return &AsyncStream{
		conc:  concurrency,
		csize: chunkSize,
	}
}

// Process reads data from the provided stream and processes each chunk using
// the provided Processor function. It processes chunks concurrently using the
// specified number of workers and stops all workers if any of them encounters
// an error.
func (ar *AsyncStream) Process(stream io.Reader, processor Processor) error {
	// Create a cancelable context. It will be used to stop all goroutines if
	// an error is encountered.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create channels for passing chunks to workers and errors back to the caller.
	errCh := make(chan error)
	chunkCh := make(chan chunk)

	// Start a goroutine to read data from the stream, divide it into chunks,
	// and send the chunks to the workers.
	go func() {
		defer close(chunkCh)
		chunkIdx := 0
		scanner := NewScanner(stream, ar.csize)
		for scanner.Scan() {
			data := scanner.Bytes()
			buf := bufferPool.Get().([]byte)
			copy(buf, data)
			select {
			case <-ctx.Done(): // Stop if the context is canceled.
				bufferPool.Put(buf) // Return the buffer to the pool.
				return
			case chunkCh <- chunk{buf: buf, idx: chunkIdx}:
				chunkIdx++
			}
		}
		// If reading failed, report the error.
		if scanner.Err() != nil {
			errCh <- scanner.Err()
		}
	}()

	// Start workers.
	var wg sync.WaitGroup
	wg.Add(ar.conc)
	for i := 0; i < ar.conc; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done(): // Stop if the context is canceled.
					return
				case chunk, ok := <-chunkCh:
					if !ok { // Stop if there are no more chunks.
						return
					}
					// Process the chunk and report any errors.
					if err := processor(chunk.buf, chunk.idx); err != nil {
						errCh <- err
						cancel() // Cancel the context, stopping all goroutines.
						return
					}
					bufferPool.Put(chunk.buf) // Return the buffer to the pool.
				}
			}
		}()
	}

	// Close the error channel after all workers have finished.
	go func() {
		wg.Wait()
		close(errCh)
	}()

	// Return the first error encountered by the workers, if any.
	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}
