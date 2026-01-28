package output

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Virace/fast-hasher/internal/scanner"
)

// TextFormatter formats results in text format compatible with md5sum/sha256sum.
type TextFormatter struct {
	// Algorithms is the list of algorithm names in order.
	// If multiple algorithms, each hash is output on a separate line as "algo:hash  path"
	Algorithms []string
}

// NewTextFormatter creates a new text formatter.
func NewTextFormatter(algorithms []string) *TextFormatter {
	return &TextFormatter{Algorithms: algorithms}
}

// Format formats a successful result.
// Single algorithm: "hash  path"
// Multiple algorithms: "algo:hash  path" (one line per algorithm)
func (f *TextFormatter) Format(result *scanner.Result) string {
	if len(f.Algorithms) == 1 {
		algo := f.Algorithms[0]
		hash := result.Hashes[algo]
		return fmt.Sprintf("%s  %s", hash, result.Path)
	}

	// Multiple algorithms: sort for consistent output
	algos := make([]string, len(f.Algorithms))
	copy(algos, f.Algorithms)
	sort.Strings(algos)

	var lines []string
	for _, algo := range algos {
		hash := result.Hashes[algo]
		lines = append(lines, fmt.Sprintf("%s:%s  %s", algo, hash, result.Path))
	}
	return strings.Join(lines, "\n")
}

// FormatError formats an error result.
func (f *TextFormatter) FormatError(result *scanner.Result) string {
	return fmt.Sprintf("# ERROR: %s: %s", result.Path, result.Error)
}
