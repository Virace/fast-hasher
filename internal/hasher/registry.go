package hasher

import (
	"fmt"
	"sort"
	"strings"
)

// registry holds all registered hashers
var registry = make(map[string]Hasher)

// Register adds a hasher to the registry.
func Register(h Hasher) {
	registry[strings.ToLower(h.Name())] = h
}

// Get returns a hasher by name.
func Get(name string) (Hasher, bool) {
	h, ok := registry[strings.ToLower(name)]
	return h, ok
}

// List returns all registered algorithm names in sorted order.
func List() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Parse parses a comma-separated list of algorithm names and returns the corresponding hashers.
// Example: "md5,sha256,blake3"
func Parse(names string) ([]Hasher, error) {
	if names == "" {
		return nil, fmt.Errorf("no algorithms specified")
	}

	parts := strings.Split(names, ",")
	hashers := make([]Hasher, 0, len(parts))
	seen := make(map[string]bool)

	for _, part := range parts {
		name := strings.TrimSpace(strings.ToLower(part))
		if name == "" {
			continue
		}

		if seen[name] {
			continue // skip duplicates
		}
		seen[name] = true

		h, ok := Get(name)
		if !ok {
			return nil, fmt.Errorf("unknown algorithm: %s (available: %s)", name, strings.Join(List(), ", "))
		}
		hashers = append(hashers, h)
	}

	if len(hashers) == 0 {
		return nil, fmt.Errorf("no valid algorithms specified")
	}

	return hashers, nil
}
