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

// flag for login
var loginAPIKey string

var loginCmd = &cobra.Command{
	Use:   "login [flags]",
	Short: "Store your Paymo API key in the macOS Keychain",
	Long: `Validate and store your Paymo API key securely in the macOS Keychain.

If --api-key is provided, it will be validated and stored immediately (overwriting any existing key). 
If not provided, you'll be prompted interactively.`,
	Example: `  paymostats login --api-key 1234567890abcdef
  paymostats login`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Fast path: flag provided -> validate -> overwrite without prompting
		if strings.TrimSpace(loginAPIKey) != "" {
			key := strings.TrimSpace(loginAPIKey)
			client := api.NewClient(key)
			if _, err := client.Me(); err != nil {
				if errors.Is(err, api.ErrUnauthorized) {
					return fmt.Errorf("the provided API key is invalid (401)")
				}
				return fmt.Errorf("could not validate API key: %w", err)
			}
			if err := config.SaveApiKey(key); err != nil {
				return fmt.Errorf("failed to save API key: %w", err)
			}
			fmt.Println("Saved in your Keychain")
			return nil
		}

		// Interactive path (no --api-key flag)
		reader := bufio.NewReader(os.Stdin)

		// If an api key already exists, offer to replace it
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

func init() {
	// Bind subcommand flags here (keeps them co-located to the command)
	loginCmd.Flags().StringVarP(&loginAPIKey, "api-key", "k", "", "Paymo API key (validated and stored)")
}
