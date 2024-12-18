package gofofa

import (
	"fmt"
	"github.com/FofaInfo/GoFOFA/pkg/readformats"
)

type Cate struct {
	Name    string   `yaml:"name"`
	Filters []string `yaml:"filters"`
}

type CusFields struct {
	Name   string `yaml:"name"`
	Fields string `yaml:"fields"`
}

type Config struct {
	Categories   []Cate      `yaml:"categories"`
	CustomFields []CusFields `yaml:"custom_fields"`
}

func LoadConfig(configFile string) (*Config, error) {
	reader := readformats.NewYAMLReader(configFile)
	var config Config
	err := reader.UnmarshalFile(&config)
	if err != nil {
		return nil, fmt.Errorf("read Config file failed: %v", err)
	}
	return &config, nil
}
