// Package scanner provides file scanning and filtering capabilities.
package scanner

import (
	"path/filepath"
	"strings"
)

// FilterOptions defines criteria for filtering files during scanning.
type FilterOptions struct {
	MaxSize      int64    // Skip files larger than this size (0 = no limit)
	MinSize      int64    // Skip files smaller than this size
	IncludeExts  []string // Only process files with these extensions (whitelist, takes priority)
	ExcludeExts  []string // Skip files with these extensions (blacklist)
	IncludeGlobs []string // Include patterns (glob)
	ExcludeGlobs []string // Exclude patterns (glob)
}

// Match returns true if the file matches the filter criteria.
// A file matches if:
// 1. Its size is within the min/max range
// 2. It passes extension filters (include takes priority over exclude)
// 3. It passes glob pattern filters (include takes priority over exclude)
func (f *FilterOptions) Match(path string, size int64) bool {
	// Size filter
	if f.MaxSize > 0 && size > f.MaxSize {
		return false
	}
	if f.MinSize > 0 && size < f.MinSize {
		return false
	}

	// Extension filter
	if !f.matchExtension(path) {
		return false
	}

	// Glob filter
	if !f.matchGlob(path) {
		return false
	}

	return true
}

// matchExtension checks extension filters.
// If IncludeExts is set, the file must have one of those extensions.
// Otherwise, if ExcludeExts is set, the file must not have those extensions.
func (f *FilterOptions) matchExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))

	// If include list is set, file must match
	if len(f.IncludeExts) > 0 {
		for _, e := range f.IncludeExts {
			if strings.ToLower(normalizeExt(e)) == ext {
				return true
			}
		}
		return false
	}

	// If exclude list is set, file must not match
	if len(f.ExcludeExts) > 0 {
		for _, e := range f.ExcludeExts {
			if strings.ToLower(normalizeExt(e)) == ext {
				return false
			}
		}
	}

	return true
}

// matchGlob checks glob pattern filters.
func (f *FilterOptions) matchGlob(path string) bool {
	// Normalize path separators for cross-platform matching
	normalizedPath := filepath.ToSlash(path)
	baseName := filepath.Base(path)

	// If include patterns are set, file must match at least one
	if len(f.IncludeGlobs) > 0 {
		matched := false
		for _, pattern := range f.IncludeGlobs {
			// Try matching against full path and basename
			if m, _ := filepath.Match(pattern, normalizedPath); m {
				matched = true
				break
			}
			if m, _ := filepath.Match(pattern, baseName); m {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// If exclude patterns are set, file must not match any
	if len(f.ExcludeGlobs) > 0 {
		for _, pattern := range f.ExcludeGlobs {
			if m, _ := filepath.Match(pattern, normalizedPath); m {
				return false
			}
			if m, _ := filepath.Match(pattern, baseName); m {
				return false
			}
		}
	}

	return true
}

// normalizeExt ensures the extension starts with a dot.
func normalizeExt(ext string) string {
	if ext == "" {
		return ""
	}
	if !strings.HasPrefix(ext, ".") {
		return "." + ext
	}
	return ext
}
