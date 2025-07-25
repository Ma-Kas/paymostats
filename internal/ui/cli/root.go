package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "paymostats",
	Short: "Check your Paymo stats",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMenu()
	},
}

func Execute() {
	_ = rootCmd.Execute()
}
