package hasher

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestRegisteredHashers(t *testing.T) {
	expected := []string{"blake3", "crc32", "md5", "quickxor", "sha1", "sha256", "sha512", "xxh128", "xxh3"}
	registered := List()

	if len(registered) != len(expected) {
		t.Errorf("Expected %d hashers, got %d: %v", len(expected), len(registered), registered)
	}

	for _, name := range expected {
		if _, ok := Get(name); !ok {
			t.Errorf("Hasher %s not found in registry", name)
		}
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:  "single algorithm",
			input: "sha256",
			want:  []string{"sha256"},
		},
		{
			name:  "multiple algorithms",
			input: "md5,sha256,blake3",
			want:  []string{"md5", "sha256", "blake3"},
		},
		{
			name:  "with spaces",
			input: "md5, sha256 , blake3",
			want:  []string{"md5", "sha256", "blake3"},
		},
		{
			name:  "case insensitive",
			input: "MD5,SHA256,BLAKE3",
			want:  []string{"md5", "sha256", "blake3"},
		},
		{
			name:  "duplicates removed",
			input: "md5,sha256,md5",
			want:  []string{"md5", "sha256"},
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "unknown algorithm",
			input:   "unknown",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashers, err := Parse(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("Parse(%q) unexpected error: %v", tt.input, err)
				return
			}
			if len(hashers) != len(tt.want) {
				t.Errorf("Parse(%q) got %d hashers, want %d", tt.input, len(hashers), len(tt.want))
				return
			}
			for i, h := range hashers {
				if h.Name() != tt.want[i] {
					t.Errorf("Parse(%q)[%d] = %s, want %s", tt.input, i, h.Name(), tt.want[i])
				}
			}
		})
	}
}

func TestHashReader(t *testing.T) {
	data := []byte("hello world")
	expectedMD5 := md5.Sum(data)
	expectedSHA1 := sha1.Sum(data)
	expectedSHA256 := sha256.Sum256(data)

	hashers, err := Parse("md5,sha1,sha256")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	results, err := HashReader(bytes.NewReader(data), hashers)
	if err != nil {
		t.Fatalf("HashReader failed: %v", err)
	}

	if got := results["md5"]; got != hex.EncodeToString(expectedMD5[:]) {
		t.Errorf("md5: got %s, want %s", got, hex.EncodeToString(expectedMD5[:]))
	}
	if got := results["sha1"]; got != hex.EncodeToString(expectedSHA1[:]) {
		t.Errorf("sha1: got %s, want %s", got, hex.EncodeToString(expectedSHA1[:]))
	}
	if got := results["sha256"]; got != hex.EncodeToString(expectedSHA256[:]) {
		t.Errorf("sha256: got %s, want %s", got, hex.EncodeToString(expectedSHA256[:]))
	}
}

func TestHashFile(t *testing.T) {
	// Create a temporary file
	dir := t.TempDir()
	path := filepath.Join(dir, "testfile.txt")
	data := []byte("test file content")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	expectedMD5 := md5.Sum(data)

	hashers, err := Parse("md5")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	results, err := HashFile(path, hashers)
	if err != nil {
		t.Fatalf("HashFile failed: %v", err)
	}

	if got := results["md5"]; got != hex.EncodeToString(expectedMD5[:]) {
		t.Errorf("md5: got %s, want %s", got, hex.EncodeToString(expectedMD5[:]))
	}
}

func TestAllHashers(t *testing.T) {
	data := []byte("test data for all hashers")

	for _, name := range List() {
		t.Run(name, func(t *testing.T) {
			h, ok := Get(name)
			if !ok {
				t.Fatalf("Hasher %s not found", name)
			}

			hashers := []Hasher{h}
			results, err := HashReader(bytes.NewReader(data), hashers)
			if err != nil {
				t.Fatalf("HashReader failed: %v", err)
			}

			result, ok := results[name]
			if !ok {
				t.Fatalf("Result for %s not found", name)
			}

			if result == "" {
				t.Errorf("Empty result for %s", name)
			}

			// Verify consistent results
			results2, _ := HashReader(bytes.NewReader(data), hashers)
			if results[name] != results2[name] {
				t.Errorf("Inconsistent results for %s: %s vs %s", name, results[name], results2[name])
			}
		})
	}
}

func BenchmarkHashers(b *testing.B) {
	data := make([]byte, 1024*1024) // 1MB
	for i := range data {
		data[i] = byte(i)
	}

	for _, name := range List() {
		h, _ := Get(name)
		b.Run(name, func(b *testing.B) {
			b.SetBytes(int64(len(data)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				HashReader(bytes.NewReader(data), []Hasher{h})
			}
		})
	}
}
