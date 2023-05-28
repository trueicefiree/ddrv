package discord

import (
	"errors"
	"io"
)

type StreamReader struct {
	pos    int64
	chunks []Chunk
	curIdx int
	closed bool
	reader io.ReadCloser
}

func NewStreamReader(chunks []Chunk, pos int64) *StreamReader {
	sr := &StreamReader{chunks: chunks, pos: pos}

	// Calculate Start and End for each part
	var offset int
	for i := range sr.chunks {
		sr.chunks[i].Start = offset
		sr.chunks[i].End = offset + sr.chunks[i].Size
		offset = sr.chunks[i].End
	}

	return sr
}

// ReadAt simulates reading from parts using part's URL, with offset
func (sr *StreamReader) Read(p []byte) (n int, err error) {
	// find the chunk containing off
	return 0, errors.New("offset not found in any chunk")
}

func (sr *StreamReader) Close() error {
	return nil
}

func (sr *StreamReader) next() error {
	// find the chunk containing off
	var start int
	for i, chunk := range sr.chunks {
		start += chunk.Size
		if start > int(sr.pos) {
			sr.curIdx = i
			break
		}
	}

	return nil
}
