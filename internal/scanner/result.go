package scanner

// Result holds the result of scanning a single file.
type Result struct {
	Path   string            // File path (relative or absolute based on input)
	Size   int64             // File size in bytes
	Hashes map[string]string // Algorithm name -> hash value
	Error  error             // Error if any (nil on success)
}

// IsError returns true if this result represents an error.
func (r *Result) IsError() bool {
	return r.Error != nil
}
