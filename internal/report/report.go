package report

import (
	"sort"

	"github.com/Ma-Kas/paymostats/internal/api"
)

type Row struct {
	Name    string
	Hours   float64
	Percent float64
}

func Build(entries []api.TimeEntry, projects map[int]string) (rows []Row, totalHours float64) {
	projectTotals := make(map[int]float64)
	var total float64
	for _, e := range entries {
		projectTotals[e.ProjectID] += e.Duration
		total += e.Duration
	}
	totalHours = total / 3600

	for pid, secs := range projectTotals {
		name := projects[pid]
		if name == "" {
			name = "Unassigned Project"
		}
		h := secs / 3600
		pct := (secs / total) * 100
		rows = append(rows, Row{Name: name, Hours: h, Percent: pct})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Percent > rows[j].Percent })
	return rows, totalHours
}
