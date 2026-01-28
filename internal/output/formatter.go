// Package output provides formatting for scan results.
package output

import (
	"github.com/Virace/fast-hasher/internal/scanner"
)

// Formatter formats scan results for output.
type Formatter interface {
	// Format formats a successful result.
	Format(result *scanner.Result) string
	// FormatError formats an error result.
	FormatError(result *scanner.Result) string
}
