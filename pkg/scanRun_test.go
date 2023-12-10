package pkg

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/adeynack/m3ugen"
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

func Test_FullConfigAndScan(t *testing.T) {
	config := m3ugen.NewDefaultConfig()
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
	config := m3ugen.NewDefaultConfig()
	config.Extensions = []string{"mpg", "mp4"}
	config.RandomizeList = false
	config.MaximumEntries = 3
	withTestFolder(t, testStructure01, config, func(t *testing.T, basePath string, entries []string) {
		assert.Len(t, entries, 3)
		et := entriesTest{t, basePath, entries}
		et.containsFile("folder1", "file2.mp4")
		et.containsFile("folder2", "file1.mpg")
		et.containsFile("folder2", "file2.mpg")
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
	testConfiguration *m3ugen.Config,
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
