package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Ma-Kas/paymostats/internal/api"
	"github.com/Ma-Kas/paymostats/internal/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "paymostats",
	Short: "Check your Paymo stats",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := config.ResolveToken()
		if err == config.ErrNoToken {
			fmt.Println("No API key found, let's log you in.")
			if err := loginCmd.RunE(cmd, args); err != nil {
				if errors.Is(err, api.ErrLoginAborted) {
					return nil
				}
				return err
			}
			// re-resolve after successful login
			token, err = config.ResolveToken()
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		// Validate token to handle 401 case immediately
		client := api.NewClient(token)
		if _, err := client.Me(); err != nil {
			if errors.Is(err, api.ErrUnauthorized) {
				fmt.Println("Your stored API key is invalid or expired.")
				fmt.Println("a) Login with a different key")
				fmt.Println("q) Quit")
				fmt.Print("> ")

				var choice string
				fmt.Scanln(&choice)
				switch strings.ToLower(strings.TrimSpace(choice)) {
				case "a":
					// Drop old token to avoid going "already logged in" path
					_ = config.DeleteToken()

					if err := loginCmd.RunE(cmd, args); err != nil {
						if errors.Is(err, api.ErrLoginAborted) {
							return nil
						}
						return err
					}
					// Success: go to regular menu
				default:
					return nil
				}
			} else {
				return err
			}
		}

		return runMenu()
	},
}

func Execute() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	_ = rootCmd.Execute()
}
