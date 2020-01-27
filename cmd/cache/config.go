package main

import (
	"errors"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type Config struct {
	MinTimeout int `yaml:"MinTimeout"`
	MaxTimeout int `yaml:"MaxTimeout"`
	StreamLen  int `yaml:"StreamLen"`
	dbAddress  string
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
	if !(c.MaxTimeout >= c.MinTimeout && c.MinTimeout > 0) {
		return errors.New("expected that MaxTimeout >= c.MinTimeout AND c.MinTimeout > 0")
	}
	if c.StreamLen <= 0 {
		return errors.New("expected that StreamLen > 0")
	}
	if len(c.dbAddress) == 0 {
		return errors.New("expected Redis db address")
	}

	return nil
}
