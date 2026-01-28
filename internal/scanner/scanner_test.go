package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Virace/fast-hasher/internal/hasher"
)

func createTestFiles(t *testing.T, dir string) {
	t.Helper()

	files := map[string]string{
		"file1.txt":             "content of file1",
		"file2.txt":             "content of file2",
		"file3.log":             "log content",
		"subdir/file4.txt":      "subdir file",
		"subdir/file5.md":       "markdown content",
		"subdir/deep/file6.txt": "deep file",
	}

	for path, content := range files {
		fullPath := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}
}

func TestScanner_ScanFile(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello world"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	hashers, _ := hasher.Parse("md5,sha256")
	s := NewScanner(hashers)

	result := s.ScanFile(testFile)
	if result == nil {
		t.Fatal("ScanFile returned nil")
	}
	if result.Error != nil {
		t.Fatalf("ScanFile error: %v", result.Error)
	}
	if result.Hashes["md5"] == "" {
		t.Error("Missing md5 hash")
	}
	if result.Hashes["sha256"] == "" {
		t.Error("Missing sha256 hash")
	}
}

func TestScanner_ScanFile_WithFilter(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello world"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	hashers, _ := hasher.Parse("md5")
	s := NewScanner(hashers)
	s.Filter = &FilterOptions{MaxSize: 5} // File is 11 bytes

	result := s.ScanFile(testFile)
	if result != nil {
		t.Error("Expected file to be filtered out")
	}
}

func TestScanner_ScanDir(t *testing.T) {
	dir := t.TempDir()
	createTestFiles(t, dir)

	hashers, _ := hasher.Parse("md5")
	s := NewScanner(hashers)
	s.Recursive = true

	var results []*Result
	for result := range s.ScanDir(dir) {
		results = append(results, result)
	}

	// Should have 6 files
	successCount := 0
	for _, r := range results {
		if r.Error == nil {
			successCount++
		}
	}
	if successCount != 6 {
		t.Errorf("Expected 6 successful results, got %d", successCount)
	}
}

func TestScanner_ScanDir_NonRecursive(t *testing.T) {
	dir := t.TempDir()
	createTestFiles(t, dir)

	hashers, _ := hasher.Parse("md5")
	s := NewScanner(hashers)
	s.Recursive = false

	var results []*Result
	for result := range s.ScanDir(dir) {
		results = append(results, result)
	}

	// Should only have root level files (3)
	successCount := 0
	for _, r := range results {
		if r.Error == nil {
			successCount++
		}
	}
	if successCount != 3 {
		t.Errorf("Expected 3 successful results, got %d", successCount)
	}
}

func TestScanner_ScanDir_WithFilter(t *testing.T) {
	dir := t.TempDir()
	createTestFiles(t, dir)

	hashers, _ := hasher.Parse("md5")
	s := NewScanner(hashers)
	s.Recursive = true
	s.Filter = &FilterOptions{IncludeExts: []string{".txt"}}

	var results []*Result
	for result := range s.ScanDir(dir) {
		results = append(results, result)
	}

	// Should only have .txt files (4)
	successCount := 0
	for _, r := range results {
		if r.Error == nil {
			successCount++
		}
	}
	if successCount != 4 {
		t.Errorf("Expected 4 successful results, got %d", successCount)
	}
}

func TestScanner_ScanFiles(t *testing.T) {
	dir := t.TempDir()
	createTestFiles(t, dir)

	paths := []string{
		filepath.Join(dir, "file1.txt"),
		filepath.Join(dir, "file2.txt"),
	}

	hashers, _ := hasher.Parse("md5")
	s := NewScanner(hashers)

	var results []*Result
	for result := range s.ScanFiles(paths) {
		results = append(results, result)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestScanner_ScanFromReader(t *testing.T) {
	dir := t.TempDir()
	createTestFiles(t, dir)

	input := strings.Join([]string{
		filepath.Join(dir, "file1.txt"),
		filepath.Join(dir, "file2.txt"),
		"# comment line",
		"",
	}, "\n")

	hashers, _ := hasher.Parse("md5")
	s := NewScanner(hashers)

	var results []*Result
	for result := range s.ScanFromReader(strings.NewReader(input)) {
		results = append(results, result)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestScanner_AbsolutePath(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	hashers, _ := hasher.Parse("md5")
	s := NewScanner(hashers)
	s.AbsolutePath = true

	result := s.ScanFile(testFile)
	if result == nil {
		t.Fatal("ScanFile returned nil")
	}

	if !filepath.IsAbs(result.Path) {
		t.Errorf("Expected absolute path, got %s", result.Path)
	}
}
