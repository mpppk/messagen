package messagen

import (
	"fmt"
	"io/ioutil"

	"golang.org/x/xerrors"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Definitions []Definition `yaml:"Definitions"`
}

func ParseYaml(filePath string) (*Config, error) {
	config := Config{Definitions: []Definition{}}
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, xerrors.Errorf("failed to parse yaml: %w", err)
	}
	fmt.Println(string(contents))
	if err := yaml.Unmarshal(contents, &config); err != nil {
		return nil, xerrors.Errorf("failed to parse yaml: %w", err)
	}
	return &config, nil
}
