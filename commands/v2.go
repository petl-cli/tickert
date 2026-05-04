package commands

import "github.com/spf13/cobra"

var v2Cmd = &cobra.Command{
	Use:   "v2",
	Short: "",
}

func init() {
	rootCmd.AddCommand(v2Cmd)
}
