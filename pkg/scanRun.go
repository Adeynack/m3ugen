package pkg

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

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
}

var (
	regexGetFileExtension = regexp.MustCompile(`^.*\.(.*)$`)
)

// Start begins the process of scanning and generating the playlist.
func Start(config *m3ugen.Config) (*ScanRun, error) {
	//if config.Verbose {
	//	fmt.Printf("Starting scan & generate process using config %+v\n", config)
	//}

	var err error
	r := &ScanRun{
		Config:          config,
		FoundFilesPaths: make([]string, 0, initialFoundFilesPathCapacity),
	}
	if err != nil { // TODO: Remove useless check. `err` is never initialized.
		return nil, err
	}

	err = r.scan()
	if err != nil {
		return nil, err
	}

	if config.DetectDuplicates {
		r.detectDuplicates()
	}

	err = r.writePlaylist()
	if err != nil {
		return nil, err
	}

	if r.Config.Verbose {
		r.logExcludedExtensions()
	}

	return r, nil
}

// LogVerbose outputs a message only when `verbose` mode is activated.
func (r *ScanRun) LogVerbose(format string, a ...interface{}) {
	if r.Config.Verbose {
		fmt.Println(fmt.Sprintf(format, a...))
	}
}

func (r *ScanRun) writePlaylist() (err error) {
	fileList := make([]string, len(r.FoundFilesPaths))
	copy(fileList, r.FoundFilesPaths)
	if r.Config.RandomizeList {
		r.LogVerbose("Shuffling the found files")
		shuffle(fileList)
	}

	foundFilesPathsCount := int64(len(r.FoundFilesPaths))
	max := r.Config.MaximumEntries
	if max < 1 {
		r.LogVerbose("No maximum entries. Writing all %d files to output.", foundFilesPathsCount)
		max = foundFilesPathsCount
	} else if max > foundFilesPathsCount {
		r.LogVerbose("Limited to %d. Writing all %d found files to output.", max, foundFilesPathsCount)
		max = foundFilesPathsCount
	} else {
		r.LogVerbose("Limited to %d. Writing the first %d found files to output.", max, max)
	}

	r.LogVerbose("Writing playlist to %s", r.Config.OutputPath)
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
	excluded := make([]string, 0, len(r.FoundExtensions))
	for extension, included := range r.FoundExtensions {
		if !included {
			excluded = append(excluded, extension)
		}
	}
	excludedList := strings.Join(excluded, ", ")
	r.LogVerbose("Extensions not considered: %s", excludedList)
}

func (r *ScanRun) detectDuplicates() {
	r.LogVerbose("Detecting duplicates")
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
			r.LogVerbose("File %q is present %d times in the search", f, c)
			duplicatesCount++
		}
	}
	r.LogVerbose("%d files were detected as duplicates", duplicatesCount)
}

func shuffle(a []string) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(a), func(i, j int) {
		a[i], a[j] = a[j], a[i]
	})
}
