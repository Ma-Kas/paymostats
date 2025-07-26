package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Ma-Kas/paymostats/internal/api"
	"github.com/Ma-Kas/paymostats/internal/config"
)

var rootCmd = &cobra.Command{
	Use:   "paymostats",
	Short: "Paymo time tracker stats",
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)

		// Resolve apiKey (env/keychain handled in config.ResolveApiKey())
		apiKey, err := config.ResolveApiKey()
		switch {
		case err == config.ErrNoApiKey:
			// Offer to login now, or quit quietly
			fmt.Println("No API key found")
			fmt.Println(strings.Repeat("=", 40))
			fmt.Println("a) Login now")
			fmt.Println("q) Quit")
			fmt.Println(strings.Repeat("=", 40))
			fmt.Print(">> ")
			choice, _ := readChoice(reader)
			if choice != "a" {
				return nil
			}

			// Call subcommand once; user may cancel (returns nil)
			if err := loginCmd.RunE(cmd, args); err != nil {
				return err
			}

			// Re-resolve after login; proceed only if present
			var rerr error
			apiKey, rerr = config.ResolveApiKey()
			if rerr != nil {
				return nil
			}

		case err != nil:
			return err
		}

		// Validate apiKey - handle 401 with re-login prompt
		client := api.NewClient(apiKey)
		if _, err := client.Me(); err != nil {
			if errors.Is(err, api.ErrUnauthorized) {
				fmt.Println("Your stored API key is invalid or expired")
				fmt.Println(strings.Repeat("=", 40))
				fmt.Println("a) Login with different API key")
				fmt.Println("q) Quit")
				fmt.Println(strings.Repeat("=", 40))
				fmt.Print(">> ")
				choice, _ := readChoice(reader)
				if choice != "a" {
					return nil
				}

				_ = config.DeleteApiKey()
				if err := loginCmd.RunE(cmd, args); err != nil {
					return err
				}

				// Re-resolve again after login - continue only if valid now
				var rerr error
				apiKey, rerr = config.ResolveApiKey()
				if rerr != nil {
					return nil
				}
				client = api.NewClient(apiKey)
				if _, verr := client.Me(); verr != nil {
					fmt.Println("Login didn't complete; try `paymostats login` again later.")
					return nil
				}
			} else {
				return err
			}
		}

		// Valid api key - run the main menu
		return runMenu()
	},
}

func Execute() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)

	if err := rootCmd.Execute(); err != nil {
		// Only unexpected errors reach here.
		fmt.Fprintln(os.Stderr, "Error:", err)
	}
}
