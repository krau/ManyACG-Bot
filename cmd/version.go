package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	Version string = "0.2.3"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of ManyACG-Bot",
	Run: func(cmd *cobra.Command, args []string) {
		ShowVersion()
	},
}

func init() {
	rootCmd.AddCommand(VersionCmd)
}

func ShowVersion() {
	fmt.Println(Version)
}
