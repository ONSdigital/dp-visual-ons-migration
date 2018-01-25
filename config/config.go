package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"encoding/json"
	"github.com/ONSdigital/go-ns/log"
)

type Model struct {
	MigrationFile  string `yaml:"migration-file"`
	VisualRSSFile  string `yaml:"visual-rss-file"`
	CollectionsDir string `yaml:"collections-dir"`
}

func Load(filename string) (*Model, error) {
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg Model
	if err := yaml.Unmarshal(source, &cfg); err != nil {
		return nil, err
	}

	b, _ := json.MarshalIndent(cfg, "", " ")
	log.Info("successfully loaded configuration", log.Data{"config": string(b)})

	return &cfg, nil
}
