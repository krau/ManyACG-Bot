package cmd

import (
	. "ManyACG-Bot/logger"
)

const (
	Version string = "v0.1.5"
)

func ShowVersion() {
	Logger.Infof("Version: %s", Version)
}
