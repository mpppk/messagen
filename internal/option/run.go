package option

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"golang.org/x/xerrors"
)

type RunCmdConfig struct {
	FilePath     string
	RootType     string
	Num          int
	InitialState map[string]string
}

func NewRunCmdConfigFromViper() (*RunCmdConfig, error) {
	rawConfig, err := newRunCmdRawConfig()
	if err != nil {
		return nil, err
	}
	return newRunCmdConfigFromRawConfig(rawConfig)
}

func newRunCmdConfigFromRawConfig(rawConfig *RunCmdRawConfig) (*RunCmdConfig, error) {
	state, err := parseKVStr(rawConfig.State)
	if err != nil {
		return nil, err
	}
	return &RunCmdConfig{
		FilePath:     rawConfig.File,
		RootType:     rawConfig.Root,
		Num:          rawConfig.Num,
		InitialState: state,
	}, nil
}

func parseKVStr(kvListStr string) (map[string]string, error) {
	m := map[string]string{}
	if kvListStr == "" {
		return m, nil
	}

	kvList := strings.Split(kvListStr, ",")
	for _, kv := range kvList {
		keyAndValue := strings.Split(kv, "=")
		if len(keyAndValue) != 2 {
			return nil, fmt.Errorf("invalid key and value string. %s", kvListStr)
		}
		m[keyAndValue[0]] = keyAndValue[1]
	}
	return m, nil
}

func newRunCmdRawConfig() (*RunCmdRawConfig, error) {
	var conf RunCmdRawConfig
	if err := viper.Unmarshal(&conf); err != nil {
		return nil, xerrors.Errorf("failed to unmarshal run command config from viper: %w", err)
	}

	return &conf, nil
}

type RunCmdRawConfig struct {
	File  string
	Root  string
	Num   int
	State string // TODO: viper cannot parse map[string]string correctly. See https://github.com/spf13/viper/issues/608
}
