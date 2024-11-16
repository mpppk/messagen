package messagen

import (
	"io/ioutil"
	"net/http"
	"strings"

	"golang.org/x/xerrors"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Definitions []*Definition `yaml:"Definitions"`
}

func ReadYamlFromFileOrUrl(filePathOrUrl string) ([]byte, error) {
	if strings.HasPrefix(filePathOrUrl, "http") {
		res, err := http.Get(filePathOrUrl)
		if err != nil {
			return nil, xerrors.Errorf("failed to fetch yaml from %s: %w", filePathOrUrl, err)
		}
		defer func() {
			if e := res.Body.Close(); e != nil {
				panic(e)
			}
		}()

		contents, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, xerrors.Errorf("failed to read yaml from %s: %w", filePathOrUrl, err)
		}
		return contents, nil
	}

	contents, err := ioutil.ReadFile(filePathOrUrl)
	if err != nil {
		return nil, xerrors.Errorf("failed to parse yaml: %w", err)
	}
	return contents, nil
}

func ParseYamlFileOrUrl(filePathOrUrl string) (*Config, error) {
	contents, err := ReadYamlFromFileOrUrl(filePathOrUrl)
	if err != nil {
		return nil, err
	}
	return ParseYaml(contents)
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
