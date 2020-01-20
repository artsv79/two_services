package main

import (
	"errors"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type Config struct {
	URLs             []string `yaml:"URLs"`
	MinTimeout       int      `yaml:"MinTimeout"`
	MaxTimeout       int      `yaml:"MaxTimeout"`
	NumberOfRequests int      `yaml:"NumberOfRequests"`
	dbAddress        string
}

func ReadConfig() Config {
	var c Config
	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
}

func (c *Config) Validate() error {
	if c.URLs == nil || len(c.URLs) == 0 {
		return errors.New("non-empty URL list should be provided")
	}
	for _, u := range c.URLs {
		if len(u) == 0 {
			return errors.New("ULR is empty")
		}
	}
	if !(c.MaxTimeout >= c.MinTimeout && c.MinTimeout > 0) {
		return errors.New("expected that MaxTimeout >= c.MinTimeout AND c.MinTimeout > 0")
	}
	if c.NumberOfRequests <= 0 {
		return errors.New("expected that numberOfRequests >= 1")
	}
	if len(c.dbAddress) == 0 {
		return errors.New("expected Redis db address")
	}

	return nil
}
