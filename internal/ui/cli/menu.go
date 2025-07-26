package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Ma-Kas/paymostats/internal/api"
	"github.com/Ma-Kas/paymostats/internal/config"
	"github.com/Ma-Kas/paymostats/internal/report"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func runMenu() error {
	apiKey, err := config.ResolveApiKey()
	if err == config.ErrNoApiKey {
		fmt.Println("No API key found, please run `paymostats login` first.")
		return nil
	}
	if err != nil {
		return err
	}
	client := api.NewClient(apiKey)
	// client.EnableDebug()

	userID, err := client.Me()
	if err != nil {
		fmt.Println("Failed to get user:", err)
		return nil
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println()
		fmt.Println(strings.ToUpper("Check your Paymo stats"))
		fmt.Println(strings.Repeat("=", 40))
		fmt.Println("a) Last two weeks")
		fmt.Println("b) Last month")
		fmt.Println("c) Last 3 months")
		fmt.Println("d) Last 6 months")
		fmt.Println("e) Forever")
		fmt.Println("q) Quit")
		fmt.Println(strings.Repeat("=", 40))
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
		fmt.Printf("No entries found for %s (%s to %s)\n",
			label, start.Format("2006-01-02"), end.Format("2006-01-02"))
		return nil
	}

	projects, err := c.Projects()
	if err != nil {
		return fmt.Errorf("fetch projects: %w", err)
	}

	rows, totalHours := report.Build(entries, projects)

	// Build the table
	tw := table.NewWriter()
	tw.SetOutputMirror(os.Stdout)
	tw.SetStyle(table.StyleLight)
	tw.Style().Format.Header = text.FormatTitle
	tw.SetTitle(fmt.Sprintf("%s\n%s to %s",
		strings.ToUpper(label),
		start.Format("2006-01-02"),
		end.Format("2006-01-02"),
	))

	tw.AppendHeader(table.Row{strings.ToUpper("Project"), strings.ToUpper("Hours"), strings.ToUpper("Percent")})
	for _, r := range rows {
		tw.AppendRow(table.Row{r.Name, fmt.Sprintf("%.1f", r.Hours), fmt.Sprintf("%.1f%%", r.Percent)})
	}

	tw.AppendSeparator()

	pctSum := 0.0
	for _, r := range rows {
		pctSum += r.Percent
	}
	tw.AppendFooter(table.Row{"", fmt.Sprintf("%.1f hrs", totalHours), fmt.Sprintf("%.1f%%", pctSum)})

	tw.Render()
	return nil
}
