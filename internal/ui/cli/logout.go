// internal/ui/cli/logout.go
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"

	"github.com/Ma-Kas/paymostats/internal/config"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove your stored Paymo API key from the Keychain",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.DeleteApiKey(); err != nil {
			if err == keyring.ErrNotFound {
				fmt.Println("No API key stored.")
				return nil
			}
			return err
		}
		fmt.Println("API key removed from Keychain")
		return nil
	},
}
