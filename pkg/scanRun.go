package pkg

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strings"

	"github.com/adeynack/m3ugen"
)

const (
	initialFoundFilesPathCapacity = 1 * (1024 ^ 2) // 1 Mi
)

// ScanRun represents a scan & playlist generation process.
type ScanRun struct {
	Config *m3ugen.Config

	FoundFilesPaths []string

	// FoundExtensions is a list of observed extensions. Value is true when
	// the extension was considered and false when excluded.
	FoundExtensions map[string]bool

	verbose func(f string, args ...any)
	debug   func(f string, args ...any)
}

var (
	regexGetFileExtension = regexp.MustCompile(`^.*\.(.*)$`)
)

// Start begins the process of scanning and generating the playlist.
func Start(config *m3ugen.Config) (*ScanRun, error) {
	r := &ScanRun{
		Config:          config,
		FoundFilesPaths: make([]string, 0, initialFoundFilesPathCapacity),
	}
	r.initializeVerboseAndDebugOutputs()
	r.debug("Starting scan & generate process using config %+v", config)

	if err := r.scan(); err != nil {
		return nil, err
	}

	if config.DetectDuplicates {
		r.detectDuplicates()
	}

	if err := r.writePlaylist(); err != nil {
		return nil, err
	}

	r.logExcludedExtensions()

	return r, nil
}

func (r *ScanRun) initializeVerboseAndDebugOutputs() {
	// VERBOSE (implicit if 'Debug' activated)
	if r.Config.Verbose || r.Config.Debug {
		r.verbose = func(format string, a ...any) {
			fmt.Fprintln(os.Stderr, fmt.Sprintf(format, a...))
		}
	} else {
		r.verbose = func(format string, a ...any) {}
	}

	// DEBUG
	if r.Config.Debug {
		r.debug = func(format string, a ...any) {
			line := fmt.Sprintf(format, a...)
			log.Default().Println(line)
		}
	} else {
		r.debug = func(format string, a ...any) {}
	}
}

func (r *ScanRun) writePlaylist() (err error) {
	fileList := make([]string, len(r.FoundFilesPaths))
	copy(fileList, r.FoundFilesPaths)
	if r.Config.RandomizeList {
		r.verbose("Shuffling the found files")
		shuffle(fileList)
	}

	foundFilesPathsCount := len(r.FoundFilesPaths)
	max := r.Config.MaximumEntries
	if max < 1 {
		r.verbose("No maximum entries. Writing all %d files to output.", foundFilesPathsCount)
		max = foundFilesPathsCount
	} else if max > foundFilesPathsCount {
		r.verbose("Limited to %d. Writing all %d found files to output.", max, foundFilesPathsCount)
		max = foundFilesPathsCount
	} else {
		r.verbose("Limited to %d. Writing the first %d found files to output.", max, max)
	}

	r.verbose("Writing playlist to %s", r.Config.OutputPath)
	f, err := os.OpenFile(r.Config.OutputPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return
	}
	defer func() {
		err = FirstErr(err, f.Close())
	}()

	w := bufio.NewWriter(f)
	defer func() {
		err = FirstErr(err, w.Flush())
	}()
	for _, entry := range fileList[:max] {
		_, err = fmt.Fprintln(w, entry)
		if err != nil {
			return
		}
	}
	return
}

func (r *ScanRun) logExcludedExtensions() {
	if !r.Config.Verbose {
		return
	}

	excluded := make([]string, 0, len(r.FoundExtensions))
	for extension, included := range r.FoundExtensions {
		if !included {
			excluded = append(excluded, extension)
		}
	}
	excludedList := strings.Join(excluded, ", ")
	r.verbose("Extensions not considered: %s", excludedList)
}

func (r *ScanRun) detectDuplicates() {
	r.verbose("Detecting duplicates")
	fileCounter := make(map[string]int)
	for _, f := range r.FoundFilesPaths {
		c, ok := fileCounter[f]
		if !ok {
			c = 0
		}
		fileCounter[f] = c + 1
	}
	duplicatesCount := 0
	for f, c := range fileCounter {
		if c > 1 {
			r.verbose("File %q is present %d times in the search", f, c)
			duplicatesCount++
		}
	}
	r.verbose("%d files were detected as duplicates", duplicatesCount)
}

func shuffle[T any](a []T) {
	rand.Shuffle(len(a), func(i, j int) { a[i], a[j] = a[j], a[i] })
}
