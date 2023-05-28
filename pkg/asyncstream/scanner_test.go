package asyncstream

import (
    "io"
    "strings"
    "testing"
)

func TestScanner(t *testing.T) {
    // Define the test cases
    testCases := []struct {
        name     string
        input    io.Reader
        bufSize  int
        expected []string
    }{
        {
            name:     "successful read",
            input:    strings.NewReader("Test data"),
            bufSize:  4,
            expected: []string{"Test", " dat", "a"},
        },
        {
            name:     "buffer size greater than data",
            input:    strings.NewReader("Test"),
            bufSize:  10,
            expected: []string{"Test"},
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            scanner := NewScanner(tc.input, tc.bufSize)

            for _, expected := range tc.expected {
                // Check that Scan() returns the correct value.
                if got := scanner.Scan(); !got {
                    t.Errorf("Expected Scan to return true, but it returned false")
                }

                // Check that Bytes() returns the correct value.
                if got := string(scanner.Bytes()); got != expected {
                    t.Errorf("Expected Bytes to return %v, but it returned %v", expected, got)
                }

                // Check that String() returns the correct value.
                if got := scanner.String(); got != expected {
                    t.Errorf("Expected String to return %v, but it returned %v", expected, got)
                }
            }

            // After all data is scanned, Scan should return false
            if scanner.Scan() {
                t.Errorf("Expected Scan to return false, but it returned true")
            }
        })
    }
}
