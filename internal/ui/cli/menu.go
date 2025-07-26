package cli

import (
	"bufio"
	"fmt"
	"math"
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
	// client.EnableDebug() // Uncomment for verbose HTTP dumps

	userID, err := client.Me()
	if err != nil {
		fmt.Println("Failed to get user:", err)
		return nil
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println()
		fmt.Println(strings.ToUpper("Display your Paymo stats"))
		fmt.Println(strings.Repeat("=", 40))
		fmt.Println("a) Last week")
		fmt.Println("b) Last two weeks")
		fmt.Println("c) Last month")
		fmt.Println("d) Last 3 months")
		fmt.Println("e) Last 6 months")
		fmt.Println("f) Year to date")
		fmt.Println("g) All time")
		fmt.Println(strings.Repeat("-", 40))
		fmt.Println("q) Quit")
		fmt.Println(strings.Repeat("=", 40))
		fmt.Print(">> ")

		choice, _ := readChoice(reader)
		if choice == "q" {
			return nil
		}
		spec, ok := choices[choice]
		if !ok {
			fmt.Println("Unknown option, try again.")
			continue
		}

		start, end := bounds(spec)
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

	// For "All Time" case, replace caption start date with earliest actual entry time
	displayStart := start
	if start.Unix() == 0 {
		minTS := int64(math.MaxInt64)
		for _, e := range entries {
			if e.StartTime != nil {
				if ts := int64(*e.StartTime); ts < minTS {
					minTS = ts
				}
			}
			if e.Date != nil {
				if ts := int64(*e.Date); ts < minTS {
					minTS = ts
				}
			}
		}
		if minTS != int64(math.MaxInt64) {
			displayStart = time.Unix(minTS, 0).UTC()
		}
	}

	rows, totalHours := report.Build(entries, projects)

	// Build the table
	tw := table.NewWriter()
	tw.SetOutputMirror(os.Stdout)
	tw.SetStyle(table.StyleLight)
	tw.Style().Format.Header = text.FormatTitle
	tw.SetTitle(fmt.Sprintf("%s\n%s to %s",
		strings.ToUpper(label),
		displayStart.Format("2006-01-02"),
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
