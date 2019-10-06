package option

import (
	"github.com/spf13/viper"
	"golang.org/x/xerrors"
)

type RootCmdConfig struct{}

func NewRootCmdConfigFromViper() (*RootCmdConfig, error) {
	rawConfig, err := newRootCmdRawConfig()
	return newRootCmdConfigFromRawConfig(rawConfig), err
}

func newRootCmdRawConfig() (*RootCmdRawConfig, error) {
	var conf RootCmdRawConfig
	if err := viper.Unmarshal(&conf); err != nil {
		return nil, xerrors.Errorf("failed to unmarshal config from viper: %w", err)
	}

	if err := conf.validate(); err != nil {
		return nil, xerrors.Errorf("failed to create root cmd config: %w", err)
	}
	return &conf, nil
}

func newRootCmdConfigFromRawConfig(rawConfig *RootCmdRawConfig) *RootCmdConfig {
	return &RootCmdConfig{}
}

type RootCmdRawConfig struct {
	File     string
	RootType string
}

func (c *RootCmdRawConfig) validate() error {
	return nil
}
