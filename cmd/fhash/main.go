// fhash is a fast, concurrent file hashing CLI tool.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Virace/fast-hasher/internal/hasher"
	"github.com/Virace/fast-hasher/internal/output"
	"github.com/Virace/fast-hasher/internal/scanner"
)

// Version information (set by build flags)
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

// Config holds the CLI configuration.
type Config struct {
	// Algorithms
	Algo string

	// Input mode
	Paths     []string
	FromFile  string
	FromStdin bool
	Recursive bool

	// Output mode
	Machine      bool
	JSON         bool
	AbsolutePath bool

	// Error handling
	OnError string

	// Filter options
	MaxSize    string
	MinSize    string
	IncludeExt string
	ExcludeExt string
	Include    string
	Exclude    string

	// Concurrency
	Workers int

	// Other
	ListAlgos bool
	Version   bool
}

func main() {
	cfg := parseFlags()

	if cfg.Version {
		fmt.Printf("fhash %s\n", Version)
		fmt.Printf("- commit: %s\n", Commit)
		fmt.Printf("- built: %s\n", BuildTime)
		fmt.Printf("- os/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("- go: %s\n", runtime.Version())
		os.Exit(0)
	}

	if cfg.ListAlgos {
		fmt.Println("Supported algorithms:")
		for _, name := range hasher.List() {
			fmt.Printf("  %s\n", name)
		}
		os.Exit(0)
	}

	if cfg.Algo == "" {
		fmt.Fprintln(os.Stderr, "Error: --algo is required")
		fmt.Fprintln(os.Stderr, "Use --list to see available algorithms")
		os.Exit(1)
	}

	// Parse algorithms
	hashers, err := hasher.Parse(cfg.Algo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create scanner
	s := scanner.NewScanner(hashers)
	s.Workers = cfg.Workers
	s.Recursive = cfg.Recursive
	s.AbsolutePath = cfg.AbsolutePath

	// Set error strategy
	if cfg.OnError == "fail" {
		s.OnError = scanner.FailOnError
	} else {
		s.OnError = scanner.SkipOnError
	}

	// Set filter options
	filter, err := parseFilterOptions(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	s.Filter = filter

	// Create formatter
	var formatter output.Formatter
	if cfg.JSON {
		formatter = output.NewJSONFormatter()
	} else {
		algoNames := make([]string, len(hashers))
		for i, h := range hashers {
			algoNames[i] = h.Name()
		}
		formatter = output.NewTextFormatter(algoNames)
	}

	// Determine input source and process
	var results <-chan *scanner.Result

	if cfg.FromStdin {
		results = s.ScanFromReader(os.Stdin)
	} else if cfg.FromFile != "" {
		f, err := os.Open(cfg.FromFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		results = s.ScanFromReader(f)
	} else if len(cfg.Paths) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no input files or directories specified")
		flag.Usage()
		os.Exit(1)
	} else {
		// Determine if paths are files or directories
		var files []string
		var dirs []string

		for _, p := range cfg.Paths {
			info, err := os.Stat(p)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				if s.OnError == scanner.FailOnError {
					os.Exit(1)
				}
				continue
			}
			if info.IsDir() {
				dirs = append(dirs, p)
			} else {
				files = append(files, p)
			}
		}

		// Create a combined results channel
		resultChans := make([]<-chan *scanner.Result, 0)

		// Files
		if len(files) > 0 {
			resultChans = append(resultChans, s.ScanFiles(files))
		}

		// Directories
		for _, dir := range dirs {
			resultChans = append(resultChans, s.ScanDir(dir))
		}

		// Merge channels
		results = mergeResultChannels(resultChans)
	}

	// Output results
	hasError := false
	for result := range results {
		if result.IsError() {
			hasError = true
			if !cfg.Machine {
				fmt.Fprintln(os.Stderr, formatter.FormatError(result))
			} else if cfg.JSON {
				fmt.Println(formatter.FormatError(result))
			}
		} else {
			fmt.Println(formatter.Format(result))
		}
	}

	if hasError && s.OnError == scanner.FailOnError {
		os.Exit(1)
	}
}

func parseFlags() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Algo, "algo", "", "Hash algorithm(s), comma-separated (required)")
	flag.StringVar(&cfg.Algo, "a", "", "Hash algorithm(s) (shorthand)")

	flag.StringVar(&cfg.FromFile, "from-file", "", "Read file paths from file (one per line)")
	flag.StringVar(&cfg.FromFile, "f", "", "Read file paths from file (shorthand)")
	flag.BoolVar(&cfg.FromStdin, "from-stdin", false, "Read file paths from stdin")

	flag.BoolVar(&cfg.Recursive, "recursive", true, "Scan directories recursively")
	flag.BoolVar(&cfg.Recursive, "r", true, "Scan directories recursively (shorthand)")

	flag.BoolVar(&cfg.Machine, "machine", false, "Machine-readable output (no progress)")
	flag.BoolVar(&cfg.Machine, "m", false, "Machine-readable output (shorthand)")
	flag.BoolVar(&cfg.JSON, "json", false, "Output as JSON Lines")
	flag.BoolVar(&cfg.JSON, "j", false, "Output as JSON Lines (shorthand)")
	flag.BoolVar(&cfg.AbsolutePath, "absolute", false, "Output absolute paths")

	flag.StringVar(&cfg.OnError, "on-error", "skip", "Error handling: skip or fail")

	flag.StringVar(&cfg.MaxSize, "max-size", "", "Skip files larger than this size (e.g., 100MB)")
	flag.StringVar(&cfg.MinSize, "min-size", "", "Skip files smaller than this size")
	flag.StringVar(&cfg.IncludeExt, "include-ext", "", "Only process files with these extensions (comma-separated)")
	flag.StringVar(&cfg.IncludeExt, "I", "", "Only process files with these extensions (shorthand)")
	flag.StringVar(&cfg.ExcludeExt, "exclude-ext", "", "Skip files with these extensions (comma-separated)")
	flag.StringVar(&cfg.ExcludeExt, "E", "", "Skip files with these extensions (shorthand)")
	flag.StringVar(&cfg.Include, "include", "", "Include glob patterns (comma-separated)")
	flag.StringVar(&cfg.Include, "i", "", "Include glob patterns (shorthand)")
	flag.StringVar(&cfg.Exclude, "exclude", "", "Exclude glob patterns (comma-separated)")
	flag.StringVar(&cfg.Exclude, "e", "", "Exclude glob patterns (shorthand)")

	flag.IntVar(&cfg.Workers, "workers", runtime.NumCPU(), "Number of concurrent workers")
	flag.IntVar(&cfg.Workers, "w", runtime.NumCPU(), "Number of concurrent workers (shorthand)")

	flag.BoolVar(&cfg.ListAlgos, "list", false, "List supported algorithms")
	flag.BoolVar(&cfg.ListAlgos, "l", false, "List supported algorithms (shorthand)")
	flag.BoolVar(&cfg.Version, "version", false, "Show version")
	flag.BoolVar(&cfg.Version, "v", false, "Show version (shorthand)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <path>...\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintln(os.Stderr, "A fast, concurrent file hashing tool.")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Examples:")
		fmt.Fprintln(os.Stderr, "  fhash -a sha256 file.txt")
		fmt.Fprintln(os.Stderr, "  fhash -a md5,sha256 ./dist")
		fmt.Fprintln(os.Stderr, "  fhash -a sha256 -m -j ./dist")
		fmt.Fprintln(os.Stderr, "  fhash -a xxh3 --max-size 100MB -E .log,.tmp ./project")
		fmt.Fprintln(os.Stderr, "  cat files.txt | fhash -a sha256 --from-stdin -m -j")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
	}

	flag.Parse()
	cfg.Paths = flag.Args()

	return cfg
}

func parseFilterOptions(cfg *Config) (*scanner.FilterOptions, error) {
	filter := &scanner.FilterOptions{}

	// Parse size limits
	if cfg.MaxSize != "" {
		size, err := parseSize(cfg.MaxSize)
		if err != nil {
			return nil, fmt.Errorf("invalid max-size: %w", err)
		}
		filter.MaxSize = size
	}
	if cfg.MinSize != "" {
		size, err := parseSize(cfg.MinSize)
		if err != nil {
			return nil, fmt.Errorf("invalid min-size: %w", err)
		}
		filter.MinSize = size
	}

	// Parse extensions
	if cfg.IncludeExt != "" {
		filter.IncludeExts = splitAndTrim(cfg.IncludeExt)
	}
	if cfg.ExcludeExt != "" {
		filter.ExcludeExts = splitAndTrim(cfg.ExcludeExt)
	}

	// Parse globs
	if cfg.Include != "" {
		filter.IncludeGlobs = splitAndTrim(cfg.Include)
	}
	if cfg.Exclude != "" {
		filter.ExcludeGlobs = splitAndTrim(cfg.Exclude)
	}

	return filter, nil
}

func parseSize(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" {
		return 0, nil
	}

	multiplier := int64(1)
	if strings.HasSuffix(s, "KB") || strings.HasSuffix(s, "K") {
		multiplier = 1024
		s = strings.TrimSuffix(strings.TrimSuffix(s, "KB"), "K")
	} else if strings.HasSuffix(s, "MB") || strings.HasSuffix(s, "M") {
		multiplier = 1024 * 1024
		s = strings.TrimSuffix(strings.TrimSuffix(s, "MB"), "M")
	} else if strings.HasSuffix(s, "GB") || strings.HasSuffix(s, "G") {
		multiplier = 1024 * 1024 * 1024
		s = strings.TrimSuffix(strings.TrimSuffix(s, "GB"), "G")
	} else if strings.HasSuffix(s, "TB") || strings.HasSuffix(s, "T") {
		multiplier = 1024 * 1024 * 1024 * 1024
		s = strings.TrimSuffix(strings.TrimSuffix(s, "TB"), "T")
	} else if strings.HasSuffix(s, "B") {
		s = strings.TrimSuffix(s, "B")
	}

	var value int64
	_, err := fmt.Sscanf(s, "%d", &value)
	if err != nil {
		return 0, fmt.Errorf("invalid size format: %s", s)
	}

	return value * multiplier, nil
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// mergeResultChannels merges multiple result channels into one.
func mergeResultChannels(chans []<-chan *scanner.Result) <-chan *scanner.Result {
	out := make(chan *scanner.Result)

	go func() {
		defer close(out)
		for _, ch := range chans {
			for r := range ch {
				out <- r
			}
		}
	}()

	return out
}

// readLines reads lines from a file, ignoring empty lines and comments.
func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}
	return lines, scanner.Err()
}
