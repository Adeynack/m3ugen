package m3ugen

import (
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/adeynack/m3ugen/pkg/dynchan"
)

const (
	reportInterval = 5 * time.Second
)

func (r *ScanRun) scan() error {
	folderToScanChanIn, folderToScanChanOut := dynchan.NewBuffered[string](uint(r.Config.ChannelsBufferSize))
	defer close(folderToScanChanIn)

	errChan := make(chan error)
	foundFileChan := make(chan string, r.Config.ChannelsBufferSize)
	excludedExtensionChan := make(chan string, r.Config.ChannelsBufferSize)

	miscWorkersWG := r.startWorkers(foundFileChan, excludedExtensionChan, errChan)
	filesToConsiderChan, filesToConsiderWG := r.startFilesToConsiderWorkers(foundFileChan, excludedExtensionChan)
	r.scanAllFolders(folderToScanChanIn, folderToScanChanOut, filesToConsiderChan, errChan)

	// All folders are scanned. Close `filesToConsiderChan` and wait for the `receiveFilesWorker`s to complete.
	close(filesToConsiderChan)
	filesToConsiderWG.Wait()

	// Close other channels and wait for goroutines to be done.
	close(errChan)
	close(foundFileChan)
	close(excludedExtensionChan)
	miscWorkersWG.Wait()

	r.verbose("scan completed")
	return nil
}

func (r *ScanRun) startWorkers(foundFileChan <-chan string, excludedExtensionChan <-chan string, errChan <-chan error) *sync.WaitGroup {
	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(3)
	go r.appendFoundFileWorker(waitGroup, foundFileChan)
	go r.appendExcludedExtensionWorker(waitGroup, excludedExtensionChan)
	go r.manageErrorsWorker(waitGroup, errChan)
	return waitGroup
}

func (r *ScanRun) startFilesToConsiderWorkers(foundFileChan chan<- string, excludedExtensionChan chan<- string) (chan<- string, *sync.WaitGroup) {
	filesToConsiderChan := make(chan string, r.Config.ChannelsBufferSize)
	filesToConsiderWG := new(sync.WaitGroup)
	filesToConsiderWG.Add(r.Config.ReceiveFilesWorkers)
	for i := 0; i < r.Config.ReceiveFilesWorkers; i++ {
		go r.receiveFilesWorker(i, filesToConsiderWG, filesToConsiderChan, foundFileChan, excludedExtensionChan)
	}
	return filesToConsiderChan, filesToConsiderWG
}

func (r *ScanRun) scanAllFolders(
	folderToScanChanIn chan<- string,
	folderToScanChanOut <-chan string,
	filesToConsiderChan chan<- string,
	errChan chan<- error,
) {
	// Start scan workers
	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(len(r.Config.ScanFolders))
	for i := 0; i < r.Config.ScanFolderWorkers; i++ {
		go r.scanFolderWorker(i, folderToScanChanIn, folderToScanChanOut, filesToConsiderChan, errChan, waitGroup)
	}
	// Feed with folders to scan
	for _, folder := range r.Config.ScanFolders {
		folderToScanChanIn <- folder
	}
	// Wait for recursive completion
	waitGroup.Wait()
}

func (r *ScanRun) scanFolderWorker(
	workerNumber int,
	folderToScanChanIn chan<- string,
	folderToScanChanOut <-chan string,
	filesToConsiderChan chan<- string,
	errChan chan<- error,
	foldersToScanWG *sync.WaitGroup,
) {
	r.verbose("[scanFolderWorker %d] Starting", workerNumber)
	defer r.verbose("[scanFolderWorker %d] Done", workerNumber)
	for {
		folderToScan, ok := <-folderToScanChanOut
		if !ok {
			return
		}
		r.verbose("[scanFolderWorker %d] Scanning %q", workerNumber, folderToScan)
		files, err := os.ReadDir(folderToScan)
		if err != nil {
			errChan <- err
		} else {
			for _, file := range files {
				path := path.Join(folderToScan, file.Name())
				if file.IsDir() {
					foldersToScanWG.Add(1)
					folderToScanChanIn <- path
				} else {
					filesToConsiderChan <- path
				}
			}
		}
		foldersToScanWG.Done()
	}
}

func (r *ScanRun) manageErrorsWorker(waitGroup *sync.WaitGroup, errChan <-chan error) {
	defer waitGroup.Done()
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

func (r *ScanRun) appendFoundFileWorker(waitGroup *sync.WaitGroup, foundFileChan <-chan string) {
	defer waitGroup.Done()
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

func (r *ScanRun) appendExcludedExtensionWorker(waitGroup *sync.WaitGroup, excludedExtensionChan <-chan string) {
	defer waitGroup.Done()
	r.verbose("[appendExcludedExtensionWorker] Start")
	defer r.verbose("[appendExcludedExtensionWorker] Done")
	for ext := range excludedExtensionChan {
		r.FoundExtensions[ext] = false
	}
}
