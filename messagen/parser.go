package messagen

import (
	"io/ioutil"

	"golang.org/x/xerrors"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Definitions []*Definition `yaml:"Definitions"`
}

func ParseYamlFile(filePath string) (*Config, error) {
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, xerrors.Errorf("failed to parse yaml: %w", err)
	}
	return ParseYaml(contents)
}

func ParseYaml(contents []byte) (*Config, error) {
	config := Config{Definitions: []*Definition{}}
	if err := yaml.Unmarshal(contents, &config); err != nil {
		return nil, xerrors.Errorf("failed to parse yaml: %w", err)
	}
	return &config, nil
}
