package quickxorhash

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestQuickXorHash(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string // base64 encoded
	}{
		{
			name:     "empty",
			input:    []byte{},
			expected: "AAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		},
		{
			name:     "hello",
			input:    []byte("hello"),
			expected: "aCgDG9jwBgAAAAAABQAAAAAAAAA=",
		},
		{
			name:     "hello world",
			input:    []byte("hello world"),
			expected: "aCgDG9jwBhDc4Q1yawMZAAAAAAA=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New()
			h.Write(tt.input)
			result := h.Sum(nil)
			got := base64.StdEncoding.EncodeToString(result)
			if got != tt.expected {
				t.Errorf("QuickXorHash(%q) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestQuickXorHashSum(t *testing.T) {
	input := []byte("test data for quickxor")
	h1 := New()
	h1.Write(input)
	result1 := h1.Sum(nil)

	result2 := Sum(input)

	if !bytes.Equal(result1, result2[:]) {
		t.Errorf("Sum() and Hash.Sum() produced different results")
	}
}

func TestQuickXorHashReset(t *testing.T) {
	h := New()
	h.Write([]byte("some data"))
	h.Reset()
	h.Write([]byte("hello"))

	expected := New()
	expected.Write([]byte("hello"))

	if !bytes.Equal(h.Sum(nil), expected.Sum(nil)) {
		t.Error("Reset() did not properly reset the hash state")
	}
}

func TestQuickXorHashSize(t *testing.T) {
	h := New()
	if h.Size() != Size {
		t.Errorf("Size() = %d, want %d", h.Size(), Size)
	}
}

func TestQuickXorHashBlockSize(t *testing.T) {
	h := New()
	if h.BlockSize() != BlockSize {
		t.Errorf("BlockSize() = %d, want %d", h.BlockSize(), BlockSize)
	}
}

func BenchmarkQuickXorHash(b *testing.B) {
	data := make([]byte, 1024*1024) // 1MB
	for i := range data {
		data[i] = byte(i)
	}

	b.ResetTimer()
	b.SetBytes(int64(len(data)))

	for i := 0; i < b.N; i++ {
		h := New()
		h.Write(data)
		h.Sum(nil)
	}
}
