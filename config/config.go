package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Model struct {
	MappingFile         string `yaml:"migration-file"`
	VisualExportFile    string `yaml:"visual-rss-file"`
	CollectionsDir      string `yaml:"collections-dir"`
	BatchSize           int    `yaml:"batch-size"`
	NationalArchivesURL string `yaml:"national-archives-url"`
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

	return &cfg, nil
}
