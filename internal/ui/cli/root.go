package cli

import (
	"fmt"

	"github.com/Ma-Kas/paymostats/internal/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "paymostats",
	Short: "Check your Paymo stats",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := config.ResolveToken()
		if err == config.ErrNoToken {
			fmt.Println("No API key found, please log in first.")
			return loginCmd.RunE(cmd, args)
		}
		return runMenu()
	},
}

func Execute() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	_ = rootCmd.Execute()
}
