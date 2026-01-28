// Package hasher provides a unified interface for various hash algorithms.
package hasher

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
)

// Hasher represents a hash algorithm that can be used to compute checksums.
type Hasher interface {
	// Name returns the algorithm name (e.g., "md5", "sha256")
	Name() string
	// New creates a new hash.Hash instance
	New() hash.Hash
	// OutputSize returns the size of the hash output in bytes
	OutputSize() int
	// IsBase64 returns true if the hash should be encoded as base64 (e.g., quickxor)
	IsBase64() bool
}

// HashResult holds the hash result for a single algorithm.
type HashResult struct {
	Algorithm string
	Hash      string
	Error     error
}

// HashReader computes hashes from an io.Reader using multiple hashers simultaneously.
// This reads the data only once, computing all hashes in parallel.
func HashReader(r io.Reader, hashers []Hasher) (map[string]string, error) {
	if len(hashers) == 0 {
		return nil, fmt.Errorf("no hashers provided")
	}

	// Create hash instances
	hashes := make([]hash.Hash, len(hashers))
	writers := make([]io.Writer, len(hashers))
	for i, h := range hashers {
		hashes[i] = h.New()
		writers[i] = hashes[i]
	}

	// Create a multi-writer to write to all hashes simultaneously
	mw := io.MultiWriter(writers...)

	// Copy data to all hashes
	if _, err := io.Copy(mw, r); err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	// Collect results
	results := make(map[string]string, len(hashers))
	for i, h := range hashers {
		sum := hashes[i].Sum(nil)
		if h.IsBase64() {
			results[h.Name()] = base64.StdEncoding.EncodeToString(sum)
		} else {
			results[h.Name()] = hex.EncodeToString(sum)
		}
	}

	return results, nil
}

// HashFile computes hashes for a file using multiple hashers.
func HashFile(path string, hashers []Hasher) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	return HashReader(f, hashers)
}
