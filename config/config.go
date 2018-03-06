package config

import (
	"encoding/json"
	"io/ioutil"
	"sort"
)

// Config loads options
type Config struct {
	InputFiles []string `json:"input_files"`
}

// New is
func New() *Config {
	return &Config{}
}

// MergeFile is a helper to load a file
func (cfg *Config) MergeFile(fileName string) error {
	buf, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(buf, cfg); err != nil {
		return err
	}
	inputs := make(map[string]bool)
	for _, v := range cfg.InputFiles {
		inputs[v] = true
	}
	cfg.InputFiles = nil
	for k := range inputs {
		cfg.InputFiles = append(cfg.InputFiles, k)
	}
	sort.Strings(cfg.InputFiles)
	return nil
}
