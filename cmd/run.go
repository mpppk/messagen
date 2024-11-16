package cmd

import (
	crand "crypto/rand"
	"fmt"
	"math"
	"math/big"
	"math/rand"

	"github.com/mpppk/messagen/internal/option"
	"github.com/mpppk/messagen/messagen"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func newRunCmd(fs afero.Fs) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Generate message",
		//Long: ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := option.NewRunCmdConfigFromViper()
			if err != nil {
				return err
			}

			if seed, err := crand.Int(crand.Reader, big.NewInt(math.MaxInt64)); err != nil {
				return err
			} else {
				rand.Seed(seed.Int64())
			}

			msgConfig, err := messagen.ParseYamlFileOrUrl(config.FilePath)
			if err != nil {
				return err
			}

			generator, err := messagen.New(nil)
			if err != nil {
				return err
			}

			if err := generator.AddDefinition(msgConfig.Definitions...); err != nil {
				return err
			}

			if config.Verbose {
				printState(config.InitialState)
			}
			if msgs, err := generator.Generate(config.RootType, config.InitialState, uint(config.Num)); err != nil {
				return err
			} else {
				for _, msg := range msgs {
					cmd.Println(msg)
					cmd.Println()
				}
			}

			return nil
		},
	}
	if err := setRunCmdFlags(cmd, fs); err != nil {
		return nil, err
	}
	return cmd, nil
}

func printState(state map[string]string) {
	if len(state) == 0 {
		return
	}
	fmt.Println("---- state ----")
	for key, value := range state {
		fmt.Println(key+":", value)
	}
	fmt.Println("---------------")
	fmt.Println()
}

func setRunCmdFlags(cmd *cobra.Command, fs afero.Fs) error {
	stringFlags := []*option.StringFlag{
		{
			Flag: &option.Flag{
				Name:         "config",
				IsPersistent: true,
				Usage:        "config file (default is $HOME/.messagen.yaml)",
			},
		},
		{
			Flag: &option.Flag{
				Name:      "file",
				Shorthand: "f",
				Usage:     "target file",
			},
			Value: "./messagen.yaml",
		},
		{
			Flag: &option.Flag{
				Name:  "root",
				Usage: "Name of definition root type",
			},
			Value: "Root",
		},
		{
			Flag: &option.Flag{
				Name:      "state",
				Shorthand: "s",
				Usage:     "initial state",
			},
			Value: "",
		},
	}

	intFlags := []*option.IntFlag{
		{
			Flag: &option.Flag{
				Name:      "num",
				Shorthand: "n",
				Usage:     "number of messages",
			},
			Value: 1,
		},
	}

	boolFlags := []*option.BoolFlag{
		{
			Flag: &option.Flag{
				Name:      "verbose",
				Shorthand: "v",
				Usage:     "verbose",
			},
			Value: false,
		},
	}

	for _, stringFlag := range stringFlags {
		if err := option.RegisterStringFlag(cmd, stringFlag); err != nil {
			return err
		}
	}

	for _, intFlag := range intFlags {
		if err := option.RegisterIntFlag(cmd, intFlag); err != nil {
			return err
		}
	}

	for _, boolFlag := range boolFlags {
		if err := option.RegisterBoolFlag(cmd, boolFlag); err != nil {
			return err
		}
	}

	return nil
}

func init() {
	cmdGenerators = append(cmdGenerators, newRunCmd)
}
