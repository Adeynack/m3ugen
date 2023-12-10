package main

import (
	"flag"
	"log"
	"os"

	"github.com/adeynack/m3ugen"
	"github.com/adeynack/m3ugen/pkg"
	"github.com/ghodss/yaml"
)

func main() {
	flag.Parse()
	configurationFile := flag.Arg(0)

	if configurationFile == "" {
		log.Fatalln("No configuration file provided")
	}

	conf, err := loadConfiguration(configurationFile)
	if err != nil {
		log.Fatalln(err)
	}

	if _, err = pkg.Start(conf); err != nil {
		log.Fatalln(err)
	}
}

func loadConfiguration(configurationFile string) (*m3ugen.Config, error) {
	content, err := os.ReadFile(configurationFile)
	if err != nil {
		return nil, err
	}
	conf := m3ugen.NewDefaultConfig()
	err = yaml.Unmarshal(content, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
