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
		reader := bufio.NewReader(os.Stdin)

		// If an api key already exists and the user explicitly ran `paymostats login`,
		// offer to replace it
		if key, err := config.ResolveApiKey(); err == nil && key != "" {
			fmt.Println("You're already logged in")
			fmt.Println(strings.Repeat("=", 40))
			fmt.Println("a) Login with different API key")
			fmt.Println("q) Quit")
			fmt.Println(strings.Repeat("=", 40))
			fmt.Print(">> ")

			choice, _ := readChoice(reader)
			if choice != "a" {
				return nil
			}
			// user wants to replace the api key - delete it to avoid looping
			_ = config.DeleteApiKey()
		}

		// Prompt for the key
		for attempts := 1; attempts <= maxLoginAttempts; attempts++ {
			fmt.Print("Paste your Paymo API key (press ENTER to cancel): ")
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != nil {
					// Unexpected I/O error
					return fmt.Errorf("read input: %w", err)
				}
			}
			apiKey := strings.TrimSpace(line)

			// Allow a quick way out
			if apiKey == "" {
				fmt.Println("Canceled")
				return nil

			}

			client := api.NewClient(apiKey)
			if _, err := client.Me(); err != nil {
				if errors.Is(err, api.ErrUnauthorized) {
					fmt.Printf("API key is invalid. Try again (%d/%d), or press ENTER to abort.\n", attempts, maxLoginAttempts)
					continue
				}
				return fmt.Errorf("Could not validate key: %v\n", err)
			}

			if err := config.SaveApiKey(apiKey); err != nil {
				return fmt.Errorf("Failed to save API key: %v\n", err)
			}
			fmt.Println("Saved in your Keychain")
			return nil
		}

		fmt.Println("Too many failed attempts. Aborting.")
		return nil
	},
}
