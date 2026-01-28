package output

import (
	"encoding/json"

	"github.com/Virace/fast-hasher/internal/scanner"
)

// JSONFormatter formats results as JSON Lines (NDJSON).
type JSONFormatter struct{}

// NewJSONFormatter creates a new JSON formatter.
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

// jsonResult is the JSON representation of a successful result.
type jsonResult struct {
	Path   string            `json:"path"`
	Size   int64             `json:"size"`
	Hashes map[string]string `json:"-"` // Will be flattened
}

// jsonErrorResult is the JSON representation of an error result.
type jsonErrorResult struct {
	Path  string `json:"path"`
	Error string `json:"error"`
}

// Format formats a successful result as JSON.
func (f *JSONFormatter) Format(result *scanner.Result) string {
	// Create a map with path and size, then add hashes
	data := make(map[string]interface{})
	data["path"] = result.Path
	data["size"] = result.Size

	// Flatten hashes into the top level
	for algo, hash := range result.Hashes {
		data[algo] = hash
	}

	b, _ := json.Marshal(data)
	return string(b)
}

// FormatError formats an error result as JSON.
func (f *JSONFormatter) FormatError(result *scanner.Result) string {
	data := jsonErrorResult{
		Path:  result.Path,
		Error: result.Error.Error(),
	}
	b, _ := json.Marshal(data)
	return string(b)
}
