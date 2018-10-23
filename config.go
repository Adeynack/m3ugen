package m3ugen

const (
	initialScanSliceCap = 10000
)

// Config is the configuration a playlist generation needs to be
// performed.
//
// Verbose			If the detailed information should be outputted to the
// 					console as it is scanning and generating the playlist.
// OutputPath 		The path of the output playlist.
// ScanFolders		The list of folders to scan for files.
// Extensions		List of extensions to filter for. If empty, do not filter
// 					on extensions.
// RandomizeList	If the list should be written in the order the files were
// 					scanned (false) or in a randomised way (true).
// MaximumEntries	Maximum entries to output in the playlist
//					-1 means "none".
type Config struct {
	Verbose        bool     `json:"verbose"`
	OutputPath     string   `json:"output"`
	ScanFolders    []string `json:"scan"`
	Extensions     []string `json:"extensions"`
	RandomizeList  bool     `json:"randomize"`
	MaximumEntries int64    `json:"maximum_entries"`
}

func NewDefaultConfig() *Config {
	return &Config{
		Verbose:        false,
		OutputPath:     "",
		ScanFolders:    nil,
		Extensions:     nil,
		RandomizeList:  false,
		MaximumEntries: -1,
	}
}
