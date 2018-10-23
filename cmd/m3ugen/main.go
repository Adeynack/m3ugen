package main

import (
	"flag"
	"github.com/adeynack/m3ugen"
	"github.com/olebedev/config"
	"log"
	"strconv"
)

func main() {

	configurationFile := flag.String("conf", "", "Path to the configuration YAML file.")
	flag.Parse()

	if configurationFile == nil {
		log.Fatalln("No configuration file provided")
	}

	conf, err := config.ParseYamlFile(*configurationFile)
	if err != nil {
		log.Fatalln(err)
	}

	c := &m3ugen.Config{
		Verbose:        getVerbose(conf),
		OutputPath:     getOutputPath(conf),
		ScanFolders:    getScanFolders(conf),
		Extensions:     getExtensions(conf),
		RandomizeList:  getRandomizeList(conf),
		MaximumEntries: getMaximumEntries(conf),
	}

	m3ugen.Start(c)
}

func getVerbose(conf *config.Config) bool {
	return conf.UBool("verbose", false)
}

func getOutputPath(conf *config.Config) string {
	outputPath, err := conf.String("output")
	if err != nil {
		log.Fatalf("must have an output path configured: %s\n", err)
	}
	if outputPath == "" {
		log.Fatalln("must have an output path configured")
	}
	return outputPath
}

func getScanFolders(conf *config.Config) []string {
	list, err := conf.List("scan")
	if err != nil {
		log.Fatalf("could not get list of folder to scan: %s\n", err)
	}
	scan := make([]string, 0, len(list))
	for _, elem := range list {
		strElem, ok := elem.(string)
		if !ok {
			log.Fatalln("list of folders to scan must contain only strings")
		}
		scan = append(scan, strElem)
	}
	return scan
}

func getExtensions(conf *config.Config) []string {
	list := conf.UList("extensions")
	extensions := make([]string, 0, len(list))
	for _, elem := range list {
		strElem, ok := elem.(string)
		if !ok {
			log.Fatalln("list of extensions to scan for must contain only strings")
		}
		extensions = append(extensions, strElem)
	}
	return extensions
}

func getRandomizeList(conf *config.Config) bool {
	return conf.UBool("randomize", false)
}

func getMaximumEntries(conf *config.Config) int64 {
	rawMax := conf.UString("maximum_entries", "")
	if rawMax == "" {
		return -1
	}
	max, err := strconv.ParseInt(rawMax, 10, 64)
	if err != nil {
		log.Fatalf("cannot parse maximum entries: %s", err)
	}
	return max
}
