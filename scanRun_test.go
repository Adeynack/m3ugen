package m3ugen

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testDirectoryPath = "test_directory"
)

var (
	currentTime = time.Now()
	currentID   = int32(0)
)

// todo: Test with no `output`

func Test_InvalidConfig_MissingOutputFilePath(t *testing.T) {
	config := &Config{}
	_, err := Start(config)
	if assert.Error(t, err) {
		assert.Equal(t, "configuration requires an output file path (OutputPath)", err.Error())
	}
}

func Test_InvalidConfig_MissingFoldersToScan(t *testing.T) {
	config := &Config{OutputPath: "foo.m3u"}
	_, err := Start(config)
	if assert.Error(t, err) {
		assert.Equal(t, "configuration requires at least one folder to scan (ScanFolders)", err.Error())
	}
}

func Test_FullConfigAndScan(t *testing.T) {
	config := NewDefaultConfig()
	config.Extensions = []string{"mpg", "mp4"}
	config.RandomizeList = false
	withTestFolder(t, testStructure01, config, func(t *testing.T, basePath string, entries []string) {
		assert.Len(t, entries, 8)
		et := entriesTest{t, basePath, entries}
		et.containsFile("folder1", "file2.mp4")
		et.containsFile("folder2", "file1.mpg")
		et.containsFile("folder2", "file2.mpg")
		et.containsFile("folder2", "file3.mpg")
		et.containsFile("folder2", "subfolder1", "file3.mpg")
		et.containsFile("folder2", "subfolder1", "file4.mpg")
		et.containsFile("folder2", "subfolder2", "file3.mpg")
		et.containsFile("folder2", "subfolder2", "file4.mpg")
	})
}

func Test_FullConfigAndScan_Maximum3(t *testing.T) {
	config := NewDefaultConfig()
	config.Extensions = []string{"mpg", "mp4"}
	config.RandomizeList = false
	config.MaximumEntries = 3
	withTestFolder(t, testStructure01, config, func(t *testing.T, basePath string, entries []string) {
		assert.Len(t, entries, 3)
		// which one of the 3 entries got chosen is non-deterministic and cannot be asserted.
	})
}

func Test_DeepFolderStructure(t *testing.T) {
	// This test proved the following flaw: If `folderToScanChan` is not of a dynamic size (buffered or unbuffered),
	// pass a certain folder number, a deadlock occurs. Solution was to introduce `dynchan`.
	config := NewDefaultConfig()
	config.ReceiveFilesWorkers = 4
	config.ScanFolderWorkers = 4
	config.Extensions = []string{"mpg", "mp4"}

	filesPerFolder := 5
	foldersPerFolder := 5
	folderDepth := 5

	// Dynamically calculating the expected number of found files
	// in order to be able to play with the 3 parameters above for
	// different testing conditions.
	totalFolderCount := 1
	for b := folderDepth; b > 0; b-- {
		totalFolderCount += int(math.Pow(float64(foldersPerFolder), float64(b)))
	}
	totalFiles := totalFolderCount * filesPerFolder

	structure := generateDynamicStructure("", uint(filesPerFolder), uint(foldersPerFolder), uint(folderDepth))
	withTestFolder(t, structure, config, func(t *testing.T, basePath string, entries []string) {
		assert.Equal(t, totalFiles, int(len(entries)))
	})
}

type entriesTest struct {
	t        *testing.T
	basePath string
	entries  []string
}

func (t *entriesTest) containsFile(expectedPathParts ...string) {
	subPath := filepath.Join(expectedPathParts...)
	expectedEntry := filepath.Join(t.basePath, subPath)
	assert.Contains(t.t, t.entries, expectedEntry)
}

type TestFolderStructure struct {
	Name    string
	Folders []*TestFolderStructure
	Files   []string
}

func withTestFolder(
	t *testing.T,
	testStructure *TestFolderStructure,
	testConfiguration *Config,
	testFunc func(t *testing.T, basePath string, entries []string),
) {
	// CREATE FOLDERS AND FILES FOR TESTING
	uid := atomic.AddInt32(&currentID, 1)
	testFolderName := filepath.Join(
		os.TempDir(),
		fmt.Sprintf("%s_%d_%d", testDirectoryPath, currentTime.Unix(), uid))
	os.RemoveAll(testFolderName)

	err := os.Mkdir(testFolderName, os.ModePerm)
	if !assert.NoErrorf(t, err, "error creating test folder %q", testFolderName) {
		return
	}
	defer func() {
		err = os.RemoveAll(testFolderName)
		assert.NoError(t, err, "error removing test folder %q", testFolderName)
	}()

	err = createStructure(testFolderName, testStructure)
	if !assert.NoErrorf(t, err, "error creating test structure: %s", err) {
		return
	}

	// SCAN AND GENERATE M3U
	testConfiguration.ScanFolders = []string{testFolderName}
	testConfiguration.OutputPath = filepath.Join(testFolderName, "playlist.m3u")
	Start(testConfiguration)

	// PARSE THE GENERATED M3U
	entries, err := parseGeneratedPlaylist(testConfiguration.OutputPath)
	if !assert.NoError(t, err, "error reading generated playlist file") {
		return
	}

	// ASSERT
	testFunc(t, testFolderName, entries)
}

func parseGeneratedPlaylist(outputPath string) ([]string, error) {
	f, err := os.Open(outputPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	entries := make([]string, 0)

	reader := bufio.NewScanner(f)
	for reader.Scan() {
		entries = append(entries, reader.Text())
	}
	if err := reader.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

func createStructure(
	basePath string,
	currentFolder *TestFolderStructure,
) error {
	for _, testFile := range currentFolder.Files {
		filePath := filepath.Join(basePath, testFile)
		file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating test file %q: %s", filePath, err)
		}
		file.Close()
	}
	for _, testFolder := range currentFolder.Folders {
		folderPath := filepath.Join(basePath, testFolder.Name)
		err := os.Mkdir(folderPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating test folder %q: %s", folderPath, err)
		}
		err = createStructure(folderPath, testFolder)
		if err != nil {
			return err
		}
	}
	return nil
}

func generateDynamicStructure(name string, filesPerFolder uint, foldersPerFolder uint, folderDepth uint) *TestFolderStructure {
	folder := &TestFolderStructure{Name: name}

	folder.Files = make([]string, filesPerFolder)
	for i := uint(0); i < filesPerFolder; i++ {
		folder.Files[i] = fmt.Sprintf("file_%04d.mpg", i)
	}

	if folderDepth > 0 {
		folder.Folders = make([]*TestFolderStructure, foldersPerFolder)
		for i := uint(0); i < foldersPerFolder; i++ {
			folder.Folders[i] = generateDynamicStructure(
				fmt.Sprintf("folder_%04d", i),
				filesPerFolder, foldersPerFolder, folderDepth-1)
		}
	} else {
		folder.Folders = make([]*TestFolderStructure, 0)
	}

	return folder
}

var (
	testStructure01 = &TestFolderStructure{
		Folders: []*TestFolderStructure{
			{
				Name: "folder1",
				Files: []string{
					"file1.mp3",
					"file2.mp4",
					"file3.txt",
				},
			},
			{
				Name: "folder2",
				Files: []string{
					"file1.mpg",
					"file2.mpg",
					"file3.mpg",
					"file4.jpg",
				},
				Folders: []*TestFolderStructure{
					{
						Name: "subfolder1",
						Files: []string{
							"file1.txt",
							"file2.txt",
							"file3.mpg",
							"file4.mpg",
							"file5.mp3",
						},
					},
					{
						Name: "subfolder2",
						Files: []string{
							"file1.txt",
							"file2.txt",
							"file3.mpg",
							"file4.mpg",
							"file5.mp3",
						},
					},
				},
			},
		},
	}
)
