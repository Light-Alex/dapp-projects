package types

import (
	"github.com/AnchoredLabs/rwa-backend/libs/core/bootstrap"
	"github.com/urfave/cli/v3"
)

const (
	ConfigFile        = bootstrap.FlagConfigFile
	ConfigFileDefault = bootstrap.FlagConfigFileDefault
	APP               = bootstrap.FlagApp
	AppAlpacaWs       = "alpaca-stream"
)

type Cli struct {
	ConfigFilePath string `json:"config_file_path"`
	App            string `json:"app"`
}

func NewCli(c *cli.Command) *Cli {
	return &Cli{
		ConfigFilePath: c.String(ConfigFile),
		App:            c.String(APP),
	}
}
