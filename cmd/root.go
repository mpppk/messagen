package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/afero"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

func NewRootCmd(fs afero.Fs) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:          "messagen",
		Short:        "messagen",
		SilenceUsage: true,
	}

	if err := setFlags(cmd, fs); err != nil {
		return nil, err
	}

	return cmd, nil
}

func setFlags(cmd *cobra.Command, fs afero.Fs) error {
	var subCmds []*cobra.Command
	for _, cmdGen := range cmdGenerators {
		subCmd, err := cmdGen(fs)
		if err != nil {
			return err
		}
		subCmds = append(subCmds, subCmd)
	}
	cmd.AddCommand(subCmds...)
	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd, err := NewRootCmd(afero.NewOsFs())
	if err != nil {
		panic(err)
	}
	rootCmd.SetOut(os.Stdout)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".messagen" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".messagen")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
