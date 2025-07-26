package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// UnixTS is a UNIX timestamp that can be unmarshaled from either a JSON number
// or a quoted JSON string (e.g. 1712345678 or "1712345678")
type UnixTS int64

// Paymo API is inconsistent in the date format, so need to handle with extended unmarshal
func (u *UnixTS) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}

	// Decode into interface{} first so we can branch by actual JSON type.
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch vv := v.(type) {
	case float64: // raw JSON number
		*u = UnixTS(int64(vv))
		return nil

	case string: // could be numeric or an ISO/RFC3339 date/time
		if vv == "" {
			return nil
		}

		// Try numeric first
		if i, err := strconv.ParseInt(vv, 10, 64); err == nil {
			*u = UnixTS(i)
			return nil
		}

		// Try RFC3339
		if t, err := time.Parse(time.RFC3339, vv); err == nil {
			*u = UnixTS(t.Unix())
			return nil
		}

		// Try date-only (YYYY-MM-DD)
		if t, err := time.Parse("2006-01-02", vv); err == nil {
			*u = UnixTS(t.Unix())
			return nil
		}

		return fmt.Errorf("unsupported timestamp format: %q", vv)

	default:
		return fmt.Errorf("unsupported JSON type for UnixTS: %T", vv)
	}
}

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

	// Paymo may send these as numbers or as strings – handle both with UnixTS
	StartTime *UnixTS `json:"start_time,omitempty"`
	Date      *UnixTS `json:"date,omitempty"`
}

type Project struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Return the current user id
func (c *Client) Me() (int, error) {
	req, _ := http.NewRequest("GET", "https://app.paymoapp.com/api/me", nil)

	var out struct {
		Users []User `json:"users"`
	}
	if err := c.do(req, &out); err != nil {
		return 0, err
	}
	if len(out.Users) == 0 {
		return 0, fmt.Errorf("no users in /me response")
	}
	return out.Users[0].ID, nil
}

// Fetch time entries for a user within [start, end] using time_interval
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

	var out struct {
		Entries []TimeEntry `json:"entries"`
	}
	if err := c.do(req, &out); err != nil {
		return nil, err
	}
	return out.Entries, nil
}

// Return a map of projectID to projectName
func (c *Client) Projects() (map[int]string, error) {
	req, _ := http.NewRequest("GET", "https://app.paymoapp.com/api/projects", nil)

	var out struct {
		Projects []Project `json:"projects"`
	}
	if err := c.do(req, &out); err != nil {
		return nil, err
	}

	m := make(map[int]string, len(out.Projects))
	for _, p := range out.Projects {
		m[p.ID] = p.Name
	}
	return m, nil
}

// Centralize HTTP call, parse JSON, and maps 401/403 to ErrUnauthorized (wrapped).
func (c *Client) do(req *http.Request, out any) error {
	req.SetBasicAuth(c.apiKey, "X")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		if out == nil {
			return nil
		}
		return json.Unmarshal(body, out)

	case http.StatusUnauthorized, http.StatusForbidden: // 401/403
		return fmt.Errorf("%w: %s", ErrUnauthorized, resp.Status)

	default:
		return fmt.Errorf("api %s: %s", resp.Status, string(body))
	}
}
