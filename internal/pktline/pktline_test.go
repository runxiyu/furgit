package pktline

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestWriteReadLineRoundtrip(t *testing.T) {
	var buf bytes.Buffer
	payload := []byte("hello\n")
	if err := WriteLine(&buf, payload); err != nil {
		t.Fatalf("WriteLine: %v", err)
	}

	dst := make([]byte, 64)
	line, n, status, err := ReadLine(&buf, dst)
	if err != nil {
		t.Fatalf("ReadLine: %v", err)
	}
	if status != StatusData {
		t.Fatalf("status: got %v, want %v", status, StatusData)
	}
	if n != len(payload) {
		t.Fatalf("n: got %d, want %d", n, len(payload))
	}
	if !bytes.Equal(line, payload) {
		t.Fatalf("payload: got %q, want %q", line, payload)
	}
}

func TestReadLineSpecialPackets(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		status Status
	}{
		{"flush", "0000", StatusFlush},
		{"delim", "0001", StatusDelim},
		{"response_end", "0002", StatusResponseEnd},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewBufferString(tt.input)
			dst := make([]byte, 16)
			line, n, status, err := ReadLine(r, dst)
			if err != nil {
				t.Fatalf("ReadLine: %v", err)
			}
			if status != tt.status {
				t.Fatalf("status: got %v, want %v", status, tt.status)
			}
			if n != 0 || len(line) != 0 {
				t.Fatalf("expected empty payload, got %d bytes", n)
			}
		})
	}
}

func TestReadLineInvalidHeader(t *testing.T) {
	r := bytes.NewBufferString("zzzz")
	dst := make([]byte, 16)
	_, _, _, err := ReadLine(r, dst)
	if !errors.Is(err, ErrInvalidHeader) {
		t.Fatalf("expected ErrInvalidHeader, got %v", err)
	}
}

func TestReadLineBufferTooSmall(t *testing.T) {
	var buf bytes.Buffer
	payload := []byte("abcd")
	if err := WriteLine(&buf, payload); err != nil {
		t.Fatalf("WriteLine: %v", err)
	}
	dst := make([]byte, 2)
	_, _, _, err := ReadLine(&buf, dst)
	if !errors.Is(err, ErrBufferTooSmall) {
		t.Fatalf("expected ErrBufferTooSmall, got %v", err)
	}
}

func TestWriteLineTooLarge(t *testing.T) {
	payload := make([]byte, maxPacketDataLen+1)
	if err := WriteLine(io.Discard, payload); !errors.Is(err, ErrPacketTooLarge) {
		t.Fatalf("expected ErrPacketTooLarge, got %v", err)
	}
}
