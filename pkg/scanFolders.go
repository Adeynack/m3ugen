package pkg

import (
	"io/ioutil"
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
		go r.receiveFilesWorker(receiveFilesWorkersWG, filesToConsiderChan, foundFileChan, excludedExtensionChan)
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

	r.LogVerbose("scan completed")
	return nil
}

func (r *ScanRun) scanFolderWorker(
	workerNumber int,
	folderToScanChan chan string,
	filesToConsiderChan chan<- string,
	errChan chan<- error,
	foldersToScanWG *sync.WaitGroup,
) {
	r.LogVerbose("[scanFolderWorker %d] Starting", workerNumber)
	defer r.LogVerbose("[scanFolderWorker %d] Done", workerNumber)
	for {
		f, ok := <-folderToScanChan
		if !ok {
			return
		}
		r.LogVerbose("[scanFolderWorker %d] Scanning %q", workerNumber, f)
		files, err := ioutil.ReadDir(f)
		if err != nil {
			errChan <- err
		} else {
			// foldersInDir := make([]string, 0)
			for _, file := range files {
				path := path.Join(f, file.Name())
				switch {
				case file.IsDir():
					// foldersInDir = append(foldersInDir, path)
					foldersToScanWG.Add(1)
					go func(p string) {
						folderToScanChan <- p
					}(path)
				default:
					filesToConsiderChan <- path
				}
			}
		}
		foldersToScanWG.Done()
	}
}

func (r *ScanRun) manageErrorsWorker(errChan <-chan error) {
	r.LogVerbose("[manageErrorsWorker] Start")
	defer r.LogVerbose("[manageErrorsWorker] Done")
	for {
		select {
		case err, ok := <-errChan:
			if !ok {
				return
			}
			r.LogVerbose("ERROR: %v", err)
		}
	}
}

func (r *ScanRun) receiveFilesWorker(
	receiveFilesWorkersWG *sync.WaitGroup,
	filesToConsiderChan <-chan string,
	foundFileChan chan<- string,
	excludedExtensionChan chan<- string,
) {
	defer receiveFilesWorkersWG.Done()

	if len(r.Config.Extensions) == 0 {
		r.receiveFilesWorkerPlain(filesToConsiderChan, foundFileChan)
	} else {
		r.receiveFilesWorkerWithExtensionFilter(filesToConsiderChan, foundFileChan, excludedExtensionChan)
	}
}

func (r *ScanRun) receiveFilesWorkerPlain(
	filesToConsiderChan <-chan string,
	foundFileChan chan<- string,
) {
	r.LogVerbose("[receiveFilesWorkerPlain] Start")
	defer r.LogVerbose("[receiveFilesWorkerPlain] Done")
	for {
		select {
		case f, ok := <-filesToConsiderChan:
			if !ok {
				return
			}
			foundFileChan <- f
		}
	}
}

func (r *ScanRun) receiveFilesWorkerWithExtensionFilter(
	filesToConsiderChan <-chan string,
	foundFileChan chan<- string,
	excludedExtensionChan chan<- string,
) {
	r.LogVerbose("[receiveFilesWorkerWithExtensionFilter] Start")
	defer r.LogVerbose("[receiveFilesWorkerWithExtensionFilter] Done")
	r.FoundExtensions = make(map[string]bool)
	for {
		select {
		case fullPath, ok := <-filesToConsiderChan:
			if !ok {
				return
			}
			matches := regexGetFileExtension.FindStringSubmatch(fullPath)
			currentFileExtension := ""
			if len(matches) > 1 {
				currentFileExtension = matches[len(matches)-1]
			}
			extensionExcluded := true
			for _, configuredExtension := range r.Config.Extensions {
				if strings.EqualFold(configuredExtension, currentFileExtension) {
					//r.LogVerbose(
					//	"File %q matches configured extension %q and is being considered",
					//	fullPath, configuredExtension)
					foundFileChan <- fullPath
					extensionExcluded = false
					break
				}
			}
			if extensionExcluded {
				//r.LogVerbose(
				//	"File %q does not match any configured extension and is being ignored",
				//	fullPath)
				excludedExtensionChan <- currentFileExtension
			}
		}
	}
}

func (r *ScanRun) appendFoundFileWorker(foundFileChan <-chan string) {
	r.LogVerbose("[appendFoundFileWorker] Start")
	defer r.LogVerbose("[appendFoundFileWorker] Done")

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
			r.LogVerbose("... %d files found", len(r.FoundFilesPaths))
		}
	}
}

func (r *ScanRun) appendExcludedExtensionWorker(excludedExtensionChan <-chan string) {
	r.LogVerbose("[appendExcludedExtensionWorker] Start")
	defer r.LogVerbose("[appendExcludedExtensionWorker] Done")
	for {
		select {
		case ext, ok := <-excludedExtensionChan:
			if !ok {
				return
			}
			r.FoundExtensions[ext] = false
		}
	}
}
