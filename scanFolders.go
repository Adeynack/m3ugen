package m3ugen

import (
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

const (
	reportInterval = 5 * time.Second
)

func (r *ScanRun) scan() error {
	folderToScanChan := make(chan string, 32)
	defer close(folderToScanChan)
	filesToConsiderChan := make(chan string, 1024)
	errChan := make(chan error)
	defer close(errChan)
	foundFileChan := make(chan string)
	defer close(foundFileChan)
	excludedExtensionChan := make(chan string)
	defer close(excludedExtensionChan)

	go r.appendFoundFileWorker(foundFileChan)
	go r.appendExcludedExtensionWorker(excludedExtensionChan)
	go r.manageErrorsWorker(errChan)

	receiveFilesWorkersWG := new(sync.WaitGroup)
	receiveFilesWorkersWG.Add(r.Config.ReceiveFilesWorkers)
	for i := 0; i < r.Config.ReceiveFilesWorkers; i++ {
		go r.receiveFilesWorker(i, receiveFilesWorkersWG, filesToConsiderChan, foundFileChan, excludedExtensionChan)
	}

	// Start scan workers + Feed with folders to scan + Wait for their completion
	foldersToScanWG := new(sync.WaitGroup)
	foldersToScanWG.Add(len(r.Config.ScanFolders))
	for i := 0; i < r.Config.ScanFolderWorkers; i++ {
		go r.scanFolderWorker(i, folderToScanChan, filesToConsiderChan, errChan, foldersToScanWG)
	}
	for _, folder := range r.Config.ScanFolders {
		folderToScanChan <- folder
	}
	foldersToScanWG.Wait()

	// All folders are scanned. Close `filesToConsiderChan` and wait for the `receiveFilesWorker`s to complete.
	close(filesToConsiderChan)
	receiveFilesWorkersWG.Wait()

	r.verbose("scan completed")
	return nil
}

func (r *ScanRun) scanFolderWorker(
	workerNumber int,
	folderToScanChan chan string,
	filesToConsiderChan chan<- string,
	errChan chan<- error,
	foldersToScanWG *sync.WaitGroup,
) {
	r.verbose("[scanFolderWorker %d] Starting", workerNumber)
	defer r.verbose("[scanFolderWorker %d] Done", workerNumber)
	for {
		f, ok := <-folderToScanChan
		if !ok {
			return
		}
		r.verbose("[scanFolderWorker %d] Scanning %q", workerNumber, f)
		files, err := os.ReadDir(f)
		if err != nil {
			errChan <- err
		} else {
			for _, file := range files {
				path := path.Join(f, file.Name())
				if file.IsDir() {
					foldersToScanWG.Add(1)
					folderToScanChan <- path
				} else {
					filesToConsiderChan <- path
				}
			}
		}
		foldersToScanWG.Done()
	}
}

func (r *ScanRun) manageErrorsWorker(errChan <-chan error) {
	r.verbose("[manageErrorsWorker] Start")
	defer r.verbose("[manageErrorsWorker] Done")
	for err := range errChan {
		r.verbose("ERROR: %v", err)
	}
}

func (r *ScanRun) receiveFilesWorker(
	workerNumber int,
	receiveFilesWorkersWG *sync.WaitGroup,
	filesToConsiderChan <-chan string,
	foundFileChan chan<- string,
	excludedExtensionChan chan<- string,
) {
	defer receiveFilesWorkersWG.Done()
	if len(r.Config.Extensions) == 0 {
		r.receiveFilesWorkerPlain(workerNumber, filesToConsiderChan, foundFileChan)
	} else {
		r.receiveFilesWorkerWithExtensionFilter(workerNumber, filesToConsiderChan, foundFileChan, excludedExtensionChan)
	}
}

func (r *ScanRun) receiveFilesWorkerPlain(
	workerNumber int,
	filesToConsiderChan <-chan string,
	foundFileChan chan<- string,
) {
	r.verbose("[receiveFilesWorkerPlain %d] Start", workerNumber)
	defer r.verbose("[receiveFilesWorkerPlain %d] Done", workerNumber)
	for f := range filesToConsiderChan {
		foundFileChan <- f
	}
}

func (r *ScanRun) receiveFilesWorkerWithExtensionFilter(
	workerNumber int,
	filesToConsiderChan <-chan string,
	foundFileChan chan<- string,
	excludedExtensionChan chan<- string,
) {
	r.verbose("[receiveFilesWorkerWithExtensionFilter %d] Start", workerNumber)
	defer r.verbose("[receiveFilesWorkerWithExtensionFilter %d] Done", workerNumber)
	r.FoundExtensions = make(map[string]bool)
	for fullPath := range filesToConsiderChan {
		r.debug("[receiveFilesWorkerWithExtensionFilter %d] Considering file: %s", workerNumber, fullPath)
		matches := regexGetFileExtension.FindStringSubmatch(fullPath)
		var currentFileExtension string
		if len(matches) > 1 {
			currentFileExtension = matches[len(matches)-1]
		}
		extensionExcluded := true
		for _, configuredExtension := range r.Config.Extensions {
			if strings.EqualFold(configuredExtension, currentFileExtension) {
				r.debug("[receiveFilesWorkerWithExtensionFilter %d] File matches configured extension %q and is being considered: %s",
					workerNumber, configuredExtension, fullPath)
				foundFileChan <- fullPath
				extensionExcluded = false // TODO: Inline `if` and get rid of this variable ==> lower-case the configured extensions (once at init) and use `slices.Contains` instead of this range/if trick.
				break
			}
		}
		if extensionExcluded {
			r.debug("[receiveFilesWorkerWithExtensionFilter %d] File does not match any configured extension and is being ignored: %s",
				workerNumber, fullPath)
			excludedExtensionChan <- currentFileExtension
		}
	}
}

func (r *ScanRun) appendFoundFileWorker(foundFileChan <-chan string) {
	r.verbose("[appendFoundFileWorker] Start")
	defer r.verbose("[appendFoundFileWorker] Done")

	reportTicker := time.NewTicker(reportInterval)
	defer reportTicker.Stop()

	for {
		select {

		case f, ok := <-foundFileChan:
			if !ok {
				return
			}
			r.FoundFilesPaths = append(r.FoundFilesPaths, f)

		case <-reportTicker.C:
			r.verbose("... %d files found", len(r.FoundFilesPaths))
		}
	}
}

func (r *ScanRun) appendExcludedExtensionWorker(excludedExtensionChan <-chan string) {
	r.verbose("[appendExcludedExtensionWorker] Start")
	defer r.verbose("[appendExcludedExtensionWorker] Done")
	for ext := range excludedExtensionChan {
		r.FoundExtensions[ext] = false
	}
}
