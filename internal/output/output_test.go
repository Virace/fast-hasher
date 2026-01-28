package output

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/Virace/fast-hasher/internal/scanner"
)

func TestTextFormatter_Format_SingleAlgorithm(t *testing.T) {
	f := NewTextFormatter([]string{"sha256"})
	result := &scanner.Result{
		Path: "test.txt",
		Size: 100,
		Hashes: map[string]string{
			"sha256": "abc123def456",
		},
	}

	got := f.Format(result)
	expected := "abc123def456  test.txt"

	if got != expected {
		t.Errorf("Format() = %q, want %q", got, expected)
	}
}

func TestTextFormatter_Format_MultipleAlgorithms(t *testing.T) {
	f := NewTextFormatter([]string{"md5", "sha256"})
	result := &scanner.Result{
		Path: "test.txt",
		Size: 100,
		Hashes: map[string]string{
			"md5":    "aabbccdd",
			"sha256": "11223344",
		},
	}

	got := f.Format(result)
	// Should have both lines (sorted by algorithm name)
	if !strings.Contains(got, "md5:aabbccdd  test.txt") {
		t.Errorf("Missing md5 line in output: %q", got)
	}
	if !strings.Contains(got, "sha256:11223344  test.txt") {
		t.Errorf("Missing sha256 line in output: %q", got)
	}
}

func TestTextFormatter_FormatError(t *testing.T) {
	f := NewTextFormatter([]string{"sha256"})
	result := &scanner.Result{
		Path:  "missing.txt",
		Error: errors.New("file not found"),
	}

	got := f.FormatError(result)
	if !strings.Contains(got, "ERROR") {
		t.Errorf("Error format should contain ERROR: %q", got)
	}
	if !strings.Contains(got, "missing.txt") {
		t.Errorf("Error format should contain path: %q", got)
	}
}

func TestJSONFormatter_Format(t *testing.T) {
	f := NewJSONFormatter()
	result := &scanner.Result{
		Path: "test.txt",
		Size: 100,
		Hashes: map[string]string{
			"md5":    "aabbccdd",
			"sha256": "11223344",
		},
	}

	got := f.Format(result)

	// Parse JSON to verify
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(got), &data); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	if data["path"] != "test.txt" {
		t.Errorf("path = %v, want test.txt", data["path"])
	}
	if data["size"].(float64) != 100 {
		t.Errorf("size = %v, want 100", data["size"])
	}
	if data["md5"] != "aabbccdd" {
		t.Errorf("md5 = %v, want aabbccdd", data["md5"])
	}
	if data["sha256"] != "11223344" {
		t.Errorf("sha256 = %v, want 11223344", data["sha256"])
	}
}

func TestJSONFormatter_FormatError(t *testing.T) {
	f := NewJSONFormatter()
	result := &scanner.Result{
		Path:  "missing.txt",
		Error: errors.New("file not found"),
	}

	got := f.FormatError(result)

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(got), &data); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	if data["path"] != "missing.txt" {
		t.Errorf("path = %v, want missing.txt", data["path"])
	}
	if data["error"] != "file not found" {
		t.Errorf("error = %v, want 'file not found'", data["error"])
	}
}
