package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	apiKey string
	http   *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		http:   http.DefaultClient,
	}
}

type User struct {
	ID int `json:"id"`
}

type TimeEntry struct {
	ProjectID int     `json:"project_id"`
	Duration  float64 `json:"duration"` // seconds
}

type Project struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (c *Client) Me() (int, error) {
	req, _ := http.NewRequest("GET", "https://app.paymoapp.com/api/me", nil)
	req.SetBasicAuth(c.apiKey, "X")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("GET /me: %s - %s", resp.Status, string(b))
	}

	var out struct {
		Users []User `json:"users"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return 0, err
	}
	if len(out.Users) == 0 {
		return 0, fmt.Errorf("no users in /me response")
	}
	return out.Users[0].ID, nil
}

func (c *Client) Entries(userID int, start, end time.Time) ([]TimeEntry, error) {
	u, _ := url.Parse("https://app.paymoapp.com/api/entries")

	startISO := start.UTC().Format("2006-01-02T15:04:05Z")
	endISO := end.UTC().Format("2006-01-02T15:04:05Z")

	where := fmt.Sprintf(`user_id=%d and time_interval in ("%s","%s")`,
		userID, startISO, endISO)

	q := u.Query()
	q.Set("where", where)
	u.RawQuery = q.Encode()

	req, _ := http.NewRequest("GET", u.String(), nil)
	req.SetBasicAuth(c.apiKey, "X")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GET /entries: %s - %s", resp.Status, string(b))
	}

	var out struct {
		Entries []TimeEntry `json:"entries"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Entries, nil
}

func (c *Client) Projects() (map[int]string, error) {
	req, _ := http.NewRequest("GET", "https://app.paymoapp.com/api/projects", nil)
	req.SetBasicAuth(c.apiKey, "X")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GET /projects: %s - %s", resp.Status, string(b))
	}

	var out struct {
		Projects []Project `json:"projects"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	m := make(map[int]string, len(out.Projects))
	for _, p := range out.Projects {
		m[p.ID] = p.Name
	}
	return m, nil
}
