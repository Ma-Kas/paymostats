package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Ma-Kas/paymostats/internal/api"
	"github.com/Ma-Kas/paymostats/internal/report"
)

func runMenu() error {
	token := os.Getenv("PAYMOSTATS_TOKEN")
	if token == "" {
		fmt.Println("Set PAYMOSTATS_TOKEN first (export PAYMOSTATS_TOKEN=...)")
		return nil
	}
	client := api.NewClient(token)
	// client.EnableDebug()

	userID, err := client.Me()
	if err != nil {
		fmt.Println("Failed to get user:", err)
		return nil
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("Check your Paymo stats")
		fmt.Println("--------------------------")
		fmt.Println("a) Last two weeks")
		fmt.Println("b) Last month")
		fmt.Println("c) Last 3 months")
		fmt.Println("d) Last 6 months")
		fmt.Println("e) Forever")
		fmt.Println("q) Quit")
		fmt.Println("--------------------------")
		fmt.Print("> ")

		input, _ := reader.ReadString('\n')
		choice := strings.TrimSpace(strings.ToLower(input))

		if choice == "q" {
			fmt.Println("Bye!")
			return nil
		}

		spec, ok := choices[choice]
		if !ok {
			fmt.Println("Unknown option, try again.")
			continue
		}

		start, end := bounds(spec.days)
		if err := runRange(client, userID, spec.label, start, end); err != nil {
			fmt.Println("Error:", err)
		}
		fmt.Println()
	}
}

func runRange(c *api.Client, userID int, label string, start, end time.Time) error {
	entries, err := c.Entries(userID, start, end)
	if err != nil {
		return fmt.Errorf("fetch entries: %w", err)
	}
	if len(entries) == 0 {
		fmt.Printf("No entries found for %s (%s → %s)\n",
			label, start.Format("2006-01-02"), end.Format("2006-01-02"))
		return nil
	}

	projects, err := c.Projects()
	if err != nil {
		return fmt.Errorf("fetch projects: %w", err)
	}

	rows, totalHours := report.Build(entries, projects)

	fmt.Printf("\n%s — %s → %s  (Total: %.2f hrs)\n",
		label, start.Format("2006-01-02"), end.Format("2006-01-02"), totalHours)
	fmt.Println("--------------------------")
	for _, r := range rows {
		fmt.Printf("%-30s %6.2f%%  (%5.1f hrs)\n", r.Name, r.Percent, r.Hours)
	}
	return nil
}
