package cmd

import "github.com/spf13/cobra"

var px = &cobra.Command{
	Use:   "px",
	Short: "px server",
	Long:  "this is a px server",
	Run: func(cmd *cobra.Command, args []string) {

	},
	PreRun: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	rootCmd.AddCommand(px)
}
