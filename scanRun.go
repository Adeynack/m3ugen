package m3ugen

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	initialFoundFilesPathCapacity = 1 * (1024 ^ 2) // 1 Mi
)

type ScanRun struct {
	Config *Config

	FoundFilesPaths []string

	// FoundExtensions is a list of observed extensions. Value is true when
	// the extension was considered and false when excluded.
	FoundExtensions map[string]bool

	considerFile func(fullPath string) error
}

var (
	regexGetFileExtension = regexp.MustCompile("^.*\\.(.*)$")
)

func Start(config *Config) (*ScanRun, error) {
	//if config.Verbose {
	//	fmt.Printf("Starting scan & generate process using config %+v\n", config)
	//}

	var err error
	r := &ScanRun{
		Config:          config,
		FoundFilesPaths: make([]string, 0, initialFoundFilesPathCapacity),
	}
	if len(config.Extensions) == 0 {
		r.considerFile = r.considerFileWithoutExtensionFilter
	} else {
		r.considerFile = r.considerFileWithExtensionFilter
		r.FoundExtensions = make(map[string]bool)
	}
	if err != nil {
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

func (r *ScanRun) LogVerbose(format string, a ...interface{}) {
	if r.Config.Verbose {
		fmt.Println(fmt.Sprintf(format, a...))
	}
}

func (r *ScanRun) scan() error {
	for _, folder := range r.Config.ScanFolders {
		err := filepath.Walk(folder, r.walkFolder)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ScanRun) walkFolder(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}
	return r.considerFile(path)
}

func (r *ScanRun) considerFileWithoutExtensionFilter(fullPath string) error {
	r.LogVerbose("File %q is considered", fullPath)
	r.FoundFilesPaths = append(r.FoundFilesPaths, fullPath)
	return nil
}

func (r *ScanRun) considerFileWithExtensionFilter(fullPath string) error {
	matches := regexGetFileExtension.FindStringSubmatch(fullPath)
	currentFileExtension := ""
	if len(matches) > 1 {
		currentFileExtension = matches[len(matches)-1]
	}
	for _, configuredExtension := range r.Config.Extensions {
		if strings.EqualFold(configuredExtension, currentFileExtension) {
			//r.LogVerbose(
			//	"File %q matches configured extension %q and is being considered",
			//	fullPath, configuredExtension)
			r.FoundFilesPaths = append(r.FoundFilesPaths, fullPath)
			return nil
		}
	}
	//r.LogVerbose(
	//	"File %q does not match any configured extension and is being ignored",
	//	fullPath)
	r.FoundExtensions[currentFileExtension] = false
	return nil
}

func (r *ScanRun) writePlaylist() (err error) {
	fileList := make([]string, len(r.FoundFilesPaths))
	for i, file := range r.FoundFilesPaths {
		fileList[i] = file
		i++
	}
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
		err = f.Close()
	}()

	w := bufio.NewWriter(f)
	for _, entry := range fileList[:max] {
		_, err = fmt.Fprintln(w, entry)
		if err != nil {
			return
		}
	}
	err = w.Flush()
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
