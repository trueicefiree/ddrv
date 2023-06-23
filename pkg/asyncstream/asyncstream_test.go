package asyncstream

import (
	"bytes"
	"errors"
	"testing"
)

func TestProcess(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name        string
		concurrency int
		chunkSize   int
		input       string
		processor   Processor
		expectError bool
	}{
		{
			name:        "successful processing",
			concurrency: 2,
			chunkSize:   4,
			input:       "This is a test string",
			processor:   func(b []byte, start int) error { return nil },
			expectError: false,
		},
		{
			name:        "processing error",
			concurrency: 2,
			chunkSize:   4,
			input:       "This string will cause an error",
			processor:   func(b []byte, start int) error { return errors.New("processing error") },
			expectError: true,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stream := New(tc.concurrency, tc.chunkSize)
			err := stream.Process(bytes.NewBufferString(tc.input), tc.processor)
			if tc.expectError && err == nil {
				t.Errorf("expected an error, but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("got unexpected error: %v", err)
			}
		})
	}
}
