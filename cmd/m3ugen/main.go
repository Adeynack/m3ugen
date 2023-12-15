package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/adeynack/m3ugen"
	"github.com/ghodss/yaml"
)

func main() {
	flag.Parse()
	configurationFile := flag.Arg(0)
	if configurationFile == "" {
		fmt.Fprintln(os.Stderr, "no configuration file provided")
		os.Exit(1)
	}

	conf, err := loadConfiguration(configurationFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading configuration: %v\n", err)
		os.Exit(1)
	}

	if _, err = m3ugen.Start(conf); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func loadConfiguration(configurationFile string) (*m3ugen.Config, error) {
	content, err := os.ReadFile(configurationFile)
	if err != nil {
		return nil, err
	}
	conf := m3ugen.NewDefaultConfig()
	if err = yaml.Unmarshal(content, conf); err != nil {
		return nil, err
	}
	return conf, nil
}
