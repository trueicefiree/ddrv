package asyncstream

import (
	"context"
	"io"
	"sync"
)

// Processor is a function type that processes a chunk of data and
// returns an error if processing failed.
type Processor func([]byte, int, int) error

// chunk is a struct that holds a chunk of data and its start and end
// positions in the original data stream.
type chunk struct {
	buf   []byte // The chunk of data.
	start int    // The start position in the original data stream.
	end   int    // The end position in the original data stream.
}

// AsyncStream is a struct that reads data from a stream, divides
// it into chunks, and processes the chunks concurrently.
type AsyncStream struct {
	mu    sync.Mutex // Mutex to synchronize access to shared state.
	conc  int        // The number of worker goroutines to use for processing.
	csize int        // The size of each data chunk to read.
}

// New creates a new AsyncStream with the specified
// number of workers and chunk size.
func New(concurrency int, chunkSize int) *AsyncStream {
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
	chunkCh := make(chan chunk, ar.conc)

	// Start a goroutine to read data from the stream, divide it into chunks,
	// and send the chunks to the workers.
	go func() {
		defer close(chunkCh)
		byteIdx := 0
		scanner := NewScanner(stream, ar.csize)
		for scanner.Scan() {
			select {
			case <-ctx.Done(): // Stop if the context is canceled.
				return
			case chunkCh <- chunk{buf: scanner.Bytes(), start: byteIdx, end: byteIdx + len(scanner.Bytes())}:
				byteIdx += len(scanner.Bytes())
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
					if err := processor(chunk.buf, chunk.start, chunk.end); err != nil {
						errCh <- err
						cancel() // Cancel the context, stopping all goroutines.
						return
					}
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
