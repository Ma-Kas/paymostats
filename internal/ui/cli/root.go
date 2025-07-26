package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Ma-Kas/paymostats/internal/api"
	"github.com/Ma-Kas/paymostats/internal/config"
)

var (
	// root flags
	flagRange string // week|2w|month|3m|6m|ytd|all
	flagStart string // YYYY-MM-DD
	flagEnd   string // YYYY-MM-DD
)

// computeRangeFromFlags returns (label, start, end) based on flags.
// Date flags override --range if provided
func computeRangeFromFlags(rng, startStr, endStr string) (string, time.Time, time.Time, error) {
	now := time.Now().UTC()

	// Date flags override range
	if startStr != "" || endStr != "" {
		if startStr == "" {
			return "", time.Time{}, time.Time{}, fmt.Errorf("--start is required when using --start/--end")
		}
		start, err := time.Parse("2006-01-02", startStr)
		if err != nil {
			return "", time.Time{}, time.Time{}, fmt.Errorf("invalid --start date, use YYYY-MM-DD")
		}
		var end time.Time
		if endStr == "" {
			end = now
		} else {
			end, err = time.Parse("2006-01-02", endStr)
			if err != nil {
				return "", time.Time{}, time.Time{}, fmt.Errorf("invalid --end date, use YYYY-MM-DD")
			}
		}
		if end.Before(start) {
			return "", time.Time{}, time.Time{}, fmt.Errorf("--end must be >= --start")
		}
		return "Custom", start, end, nil
	}

	// Predefined ranges
	switch strings.ToLower(strings.TrimSpace(rng)) {
	case "week", "1w", "last-week":
		return "Last week", now.AddDate(0, 0, -7), now, nil
	case "2w", "two-weeks", "last-2-weeks":
		return "Last two weeks", now.AddDate(0, 0, -14), now, nil
	case "month", "1m", "last-month":
		return "Last month", now.AddDate(0, -1, 0), now, nil
	case "3m", "quarter", "last-3-months":
		return "Last 3 months", now.AddDate(0, -3, 0), now, nil
	case "6m", "last-6-months":
		return "Last 6 months", now.AddDate(0, -6, 0), now, nil
	case "ytd", "year-to-date":
		y := now.Year()
		return "Year to date", time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC), now, nil
	case "all", "forever":
		return "All time", time.Unix(0, 0), now, nil
	case "":
		return "", time.Time{}, time.Time{}, fmt.Errorf("no flags passed; run with --range or --start/--end, or use interactive mode")
	default:
		return "", time.Time{}, time.Time{}, fmt.Errorf("unknown --range %q (use: week|2w|month|3m|6m|ytd|all)", rng)
	}
}

var rootCmd = &cobra.Command{
	Use:   "paymostats [flags]",
	Short: "Paymo time tracker stats",
	Long: `paymostats shows the percentage of time you spent per Paymo project over a period.

Use it interactively (no flags) or non-interactively with flags.

- Predefined ranges: --range week|2w|month|3m|6m|ytd|all
- Explicit dates:    --start YYYY-MM-DD [--end YYYY-MM-DD]`,
	Example: `  paymostats --range 2w
  paymostats --start 2025-07-01 --end 2025-07-25
  paymostats                 # interactive menu`,

	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)

		// Resolve apiKey (env/keychain handled in config.ResolveApiKey())
		apiKey, err := config.ResolveApiKey()
		switch {
		case err == config.ErrNoApiKey:
			// If user passed flags but has no API key, don't go interactive
			if flagRange != "" || flagStart != "" || flagEnd != "" {
				fmt.Println("No API key found. Run `paymostats login --api-key <YOUR_KEY>` first")
				return nil
			}
			// Interactive path: offer to login or quit quietly
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

		// Validate API key; if invalid and flags were provided, suggest login & exit
		client := api.NewClient(apiKey)
		if _, err := client.Me(); err != nil {
			if errors.Is(err, api.ErrUnauthorized) {
				if flagRange != "" || flagStart != "" || flagEnd != "" {
					fmt.Println("Stored API key is invalid or expired. Run `paymostats login --api-key <NEW_KEY>` and try again")
					return nil
				}
				// Interactive
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

		// Valid API key paths:

		// Non-interactive mode if any flags were set
		if flagRange != "" || flagStart != "" || flagEnd != "" {
			userID, err := client.Me()
			if err != nil {
				return fmt.Errorf("failed to get user: %w", err)
			}
			label, start, end, err := computeRangeFromFlags(flagRange, flagStart, flagEnd)
			if err != nil {
				return err
			}
			return runRange(client, userID, label, start, end)
		}

		// Interactive menu
		return runMenu()
	},
}

func Execute() {
	// Subcommands
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)

	// Root flags (central, before Execute)
	rootCmd.Flags().StringVarP(&flagRange, "range", "r", "", "predefined range: week|2w|month|3m|6m|ytd|all")
	rootCmd.Flags().StringVarP(&flagStart, "start", "s", "", "start date (YYYY-MM-DD)")
	rootCmd.Flags().StringVarP(&flagEnd, "end", "e", "", "end date (YYYY-MM-DD)")

	if err := rootCmd.Execute(); err != nil {
		// Only unexpected errors reach here
		fmt.Fprintln(os.Stderr, "Error:", err)
	}
}
