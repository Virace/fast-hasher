package scanner

import (
	"testing"
)

func TestFilterOptions_Match_Size(t *testing.T) {
	tests := []struct {
		name   string
		filter FilterOptions
		size   int64
		want   bool
	}{
		{
			name:   "no size limit",
			filter: FilterOptions{},
			size:   1000000,
			want:   true,
		},
		{
			name:   "within max size",
			filter: FilterOptions{MaxSize: 1000},
			size:   500,
			want:   true,
		},
		{
			name:   "exceeds max size",
			filter: FilterOptions{MaxSize: 1000},
			size:   1500,
			want:   false,
		},
		{
			name:   "within min size",
			filter: FilterOptions{MinSize: 100},
			size:   200,
			want:   true,
		},
		{
			name:   "below min size",
			filter: FilterOptions{MinSize: 100},
			size:   50,
			want:   false,
		},
		{
			name:   "within range",
			filter: FilterOptions{MinSize: 100, MaxSize: 1000},
			size:   500,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.Match("test.txt", tt.size)
			if got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterOptions_Match_Extension(t *testing.T) {
	tests := []struct {
		name   string
		filter FilterOptions
		path   string
		want   bool
	}{
		{
			name:   "no extension filter",
			filter: FilterOptions{},
			path:   "test.txt",
			want:   true,
		},
		{
			name:   "include ext matches",
			filter: FilterOptions{IncludeExts: []string{".txt", ".md"}},
			path:   "test.txt",
			want:   true,
		},
		{
			name:   "include ext no match",
			filter: FilterOptions{IncludeExts: []string{".txt", ".md"}},
			path:   "test.exe",
			want:   false,
		},
		{
			name:   "include ext without dot",
			filter: FilterOptions{IncludeExts: []string{"txt", "md"}},
			path:   "test.txt",
			want:   true,
		},
		{
			name:   "exclude ext matches",
			filter: FilterOptions{ExcludeExts: []string{".log", ".tmp"}},
			path:   "test.log",
			want:   false,
		},
		{
			name:   "exclude ext no match",
			filter: FilterOptions{ExcludeExts: []string{".log", ".tmp"}},
			path:   "test.txt",
			want:   true,
		},
		{
			name:   "case insensitive",
			filter: FilterOptions{IncludeExts: []string{".TXT"}},
			path:   "test.txt",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.Match(tt.path, 100)
			if got != tt.want {
				t.Errorf("Match(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestFilterOptions_Match_Glob(t *testing.T) {
	tests := []struct {
		name   string
		filter FilterOptions
		path   string
		want   bool
	}{
		{
			name:   "no glob filter",
			filter: FilterOptions{},
			path:   "test.txt",
			want:   true,
		},
		{
			name:   "include glob matches",
			filter: FilterOptions{IncludeGlobs: []string{"*.txt"}},
			path:   "test.txt",
			want:   true,
		},
		{
			name:   "include glob no match",
			filter: FilterOptions{IncludeGlobs: []string{"*.txt"}},
			path:   "test.exe",
			want:   false,
		},
		{
			name:   "exclude glob matches",
			filter: FilterOptions{ExcludeGlobs: []string{"*.log"}},
			path:   "test.log",
			want:   false,
		},
		{
			name:   "exclude glob no match",
			filter: FilterOptions{ExcludeGlobs: []string{"*.log"}},
			path:   "test.txt",
			want:   true,
		},
		{
			name:   "multiple include patterns",
			filter: FilterOptions{IncludeGlobs: []string{"*.txt", "*.md"}},
			path:   "readme.md",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.Match(tt.path, 100)
			if got != tt.want {
				t.Errorf("Match(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
