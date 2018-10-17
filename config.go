package m3ugen

const (
	initialScanSliceCap = 10000
)

type Config struct {
	// If the detailed information should be outputted to the console as it is scanning
	// and generating the playlist.
	Verbose bool
	// The path of the output playlist.
	OutputPath string
	// The list of folders to scan for files.
	ScanFolders []string
	// List of extensions to filter for. If empty, do not filter on extensions.
	Extensions []string
	// If the list should be written in the order the files were scanned (false) or
	// in a randomised way (true).
	RandomizeList bool
	// Maximum entries to output in the playlist
	// -1 means "none"
	MaximumEntries int64
}
