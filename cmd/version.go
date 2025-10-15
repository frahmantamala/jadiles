package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version    string
	versionCmd = &cobra.Command{
		RunE:  printVersion,
		Use:   "version",
		Short: "to print the app version",
	}
)

func printVersion(_ *cobra.Command, _ []string) error {
	fmt.Println(version)
	return nil
}
