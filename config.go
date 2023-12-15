package m3ugen

// Config is the configuration a playlist generation needs to be performed.
type Config struct {
	// If the detailed information should be outputted to the
	// console as it is scanning and generating the playlist.
	Verbose bool `json:"verbose"`
	// Debug indicates if detailed debug information should be outputted to the console.
	Debug bool `json:"debug"`
	// The path of the output playlist.
	OutputPath string `json:"output"`
	// The list of folders to scan for files.
	ScanFolders []string `json:"scan"`
	// List of extensions to filter for. If empty, do not filter on extensions.
	Extensions []string `json:"extensions"`
	// If the list should be written in the order the files were
	// scanned (false) or in a randomised way (true).
	RandomizeList bool `json:"randomize"`
	// Maximum entries to output in the playlist -1 means "none".
	MaximumEntries int `json:"maximum_entries"`
	// If the tool should report duplicate entries in the detected files
	// (the configured path could be duplicates or include one another).
	DetectDuplicates bool `json:"detect_duplicates"`
	// Number of workers scanning the folders.
	ScanFolderWorkers int `json:"scan_folder_workers"`
	// Number of workers filtering the files.
	ReceiveFilesWorkers int `json:"receive_files_workers"`
}

// NewDefaultConfig creates a configuration with default values.
func NewDefaultConfig() *Config {
	return &Config{
		Verbose:             false,
		Debug:               false,
		OutputPath:          "",
		ScanFolders:         nil,
		Extensions:          nil,
		RandomizeList:       false,
		MaximumEntries:      -1,
		ScanFolderWorkers:   4,
		ReceiveFilesWorkers: 1,
	}
}
