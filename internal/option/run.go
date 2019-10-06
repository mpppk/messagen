package option

import (
	"github.com/spf13/viper"
	"golang.org/x/xerrors"
)

type RunCmdConfig struct {
	FilePath string
	RootType string
}

func NewRunCmdConfigFromViper() (*RunCmdConfig, error) {
	rawConfig, err := newRunCmdRawConfig()
	return newRunCmdConfigFromRawConfig(rawConfig), err
}

func newRunCmdConfigFromRawConfig(rawConfig *RunCmdRawConfig) *RunCmdConfig {
	return &RunCmdConfig{
		FilePath: rawConfig.File,
		RootType: rawConfig.Root,
	}
}

func newRunCmdRawConfig() (*RunCmdRawConfig, error) {
	var conf RunCmdRawConfig
	if err := viper.Unmarshal(&conf); err != nil {
		return nil, xerrors.Errorf("failed to unmarshal run command config from viper: %w", err)
	}

	return &conf, nil
}

type RunCmdRawConfig struct {
	File string
	Root string
}
