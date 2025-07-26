package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/Ma-Kas/paymostats/internal/config"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Store your Paymo API key securely in the macOS Keychain",
	RunE: func(cmd *cobra.Command, args []string) error {
		// already logged in?
		if _, err := config.ResolveToken(); err == nil {
			fmt.Println("You're already logged in.")
			fmt.Print("a) Login with different key\nq) Quit\n> ")
			reader := bufio.NewReader(os.Stdin)
			ans, _ := reader.ReadString('\n')
			ans = strings.TrimSpace(strings.ToLower(ans))
			if ans != "a" {
				return nil
			}
		}

		fmt.Print("Paste your Paymo API key: ")
		b, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			return err
		}
		token := strings.TrimSpace(string(b))
		if token == "" {
			return fmt.Errorf("empty API key")
		}
		if err := config.SaveToken(token); err != nil {
			return err
		}
		fmt.Println("Saved in your Keychain")
		return nil
	},
}
