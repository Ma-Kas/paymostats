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

const maxLoginAttempts = 3

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Store your Paymo API key securely in the macOS Keychain",
	RunE: func(cmd *cobra.Command, args []string) error {
		// If a token already exists and the user explicitly ran `paymostats login`,
		// offer to replace it.
		if tok, err := config.ResolveToken(); err == nil && tok != "" {
			fmt.Println("You're already logged in.")
			fmt.Println(strings.Repeat("=", 40))
			fmt.Println("a) Login with different token")
			fmt.Println("q) Quit")
			fmt.Println(strings.Repeat("=", 40))
			fmt.Print("> ")

			reader := bufio.NewReader(os.Stdin)
			line, _ := reader.ReadString('\n')
			choice := strings.ToLower(strings.TrimSpace(line))
			if choice != "a" {
				return api.ErrLoginAborted
			}
			// user wants to replace the token → delete it so we don't loop back
			_ = config.DeleteToken()
		}

		reader := bufio.NewReader(os.Stdin)

		for attempts := 1; attempts <= maxLoginAttempts; attempts++ {
			fmt.Print("Paste your Paymo API key (press ENTER to cancel): ")
			line, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			token := strings.TrimSpace(line)

			// allow a quick way out
			if token == "" {
				return api.ErrLoginAborted
			}

			client := api.NewClient(token)
			if _, err := client.Me(); err != nil {
				if errors.Is(err, api.ErrUnauthorized) {
					fmt.Printf("That key is invalid (401). Try again (%d/%d), or press ENTER to abort.\n", attempts, maxLoginAttempts)
					continue
				}
				return fmt.Errorf("could not validate key: %w", err)
			}

			if err := config.SaveToken(token); err != nil {
				return err
			}
			fmt.Println("Saved in your Keychain")
			return nil
		}

		fmt.Println("Too many failed attempts. Aborting.")
		return api.ErrLoginAborted
	},
}
