package scanner

import (
	"bufio"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/Virace/fast-hasher/internal/hasher"
)

// ErrorStrategy defines how to handle errors during scanning.
type ErrorStrategy int

const (
	// SkipOnError skips files that cause errors and continues scanning.
	SkipOnError ErrorStrategy = iota
	// FailOnError stops scanning immediately on the first error.
	FailOnError
)

// Scanner scans files and computes their hashes.
type Scanner struct {
	Workers      int            // Number of concurrent workers (default: runtime.NumCPU())
	Filter       *FilterOptions // File filter options
	Hashers      []hasher.Hasher
	OnError      ErrorStrategy
	Recursive    bool // Whether to scan directories recursively
	AbsolutePath bool // Whether to output absolute paths
}

// NewScanner creates a new scanner with default settings.
func NewScanner(hashers []hasher.Hasher) *Scanner {
	return &Scanner{
		Workers:   runtime.NumCPU(),
		Hashers:   hashers,
		OnError:   SkipOnError,
		Recursive: true,
	}
}

// ScanFile scans a single file and returns its hash result.
func (s *Scanner) ScanFile(path string) *Result {
	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		return &Result{Path: path, Error: err}
	}

	if info.IsDir() {
		return &Result{Path: path, Error: fs.ErrInvalid}
	}

	// Apply filter
	if s.Filter != nil && !s.Filter.Match(path, info.Size()) {
		return nil // Filtered out
	}

	// Compute hashes
	hashes, err := hasher.HashFile(path, s.Hashers)

	outputPath := path
	if s.AbsolutePath {
		if abs, err := filepath.Abs(path); err == nil {
			outputPath = abs
		}
	}

	return &Result{
		Path:   outputPath,
		Size:   info.Size(),
		Hashes: hashes,
		Error:  err,
	}
}

// ScanFiles scans multiple files concurrently and returns results through a channel.
func (s *Scanner) ScanFiles(paths []string) <-chan *Result {
	results := make(chan *Result, s.Workers*2)

	go func() {
		defer close(results)

		sem := make(chan struct{}, s.Workers)
		var wg sync.WaitGroup

		for _, path := range paths {
			path := path // capture loop variable

			sem <- struct{}{} // acquire
			wg.Add(1)

			go func() {
				defer func() {
					<-sem // release
					wg.Done()
				}()

				result := s.ScanFile(path)
				if result != nil {
					results <- result
					if result.Error != nil && s.OnError == FailOnError {
						return
					}
				}
			}()
		}

		wg.Wait()
	}()

	return results
}

// ScanDir scans a directory and returns results through a channel.
func (s *Scanner) ScanDir(dir string) <-chan *Result {
	results := make(chan *Result, s.Workers*2)

	go func() {
		defer close(results)

		// Collect all files first
		var files []string
		walkFn := func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				if s.OnError == FailOnError {
					return err
				}
				// Send error result
				results <- &Result{Path: path, Error: err}
				return nil
			}

			if d.IsDir() {
				if !s.Recursive && path != dir {
					return fs.SkipDir
				}
				return nil
			}

			// Get file info for filtering
			info, err := d.Info()
			if err != nil {
				if s.OnError == FailOnError {
					return err
				}
				results <- &Result{Path: path, Error: err}
				return nil
			}

			// Apply filter
			if s.Filter != nil && !s.Filter.Match(path, info.Size()) {
				return nil
			}

			files = append(files, path)
			return nil
		}

		if err := filepath.WalkDir(dir, walkFn); err != nil {
			results <- &Result{Path: dir, Error: err}
			return
		}

		// Process files concurrently
		sem := make(chan struct{}, s.Workers)
		var wg sync.WaitGroup

		for _, path := range files {
			path := path

			sem <- struct{}{}
			wg.Add(1)

			go func() {
				defer func() {
					<-sem
					wg.Done()
				}()

				result := s.processFile(path)
				if result != nil {
					results <- result
				}
			}()
		}

		wg.Wait()
	}()

	return results
}

// processFile processes a single file (used internally, assumes filtering is done).
func (s *Scanner) processFile(path string) *Result {
	info, err := os.Stat(path)
	if err != nil {
		return &Result{Path: path, Error: err}
	}

	hashes, err := hasher.HashFile(path, s.Hashers)

	outputPath := path
	if s.AbsolutePath {
		if abs, err := filepath.Abs(path); err == nil {
			outputPath = abs
		}
	}

	return &Result{
		Path:   outputPath,
		Size:   info.Size(),
		Hashes: hashes,
		Error:  err,
	}
}

// ScanFromReader reads file paths from a reader (one per line) and scans them.
func (s *Scanner) ScanFromReader(r io.Reader) <-chan *Result {
	var paths []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			paths = append(paths, line)
		}
	}

	return s.ScanFiles(paths)
}
