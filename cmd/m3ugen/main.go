package main

import (
	"flag"
	"github.com/adeynack/m3ugen"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"log"
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

	_, err = m3ugen.Start(conf)
	if err != nil {
		log.Fatalln(err)
	}
}

func loadConfiguration(configurationFile string) (*m3ugen.Config, error) {
	content, err := ioutil.ReadFile(configurationFile)
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
