package core_http_types

import (
	"encoding/json"
	"fmt"
	"time"
)

const DateLayout = "2006-01-02"

// Date is a calendar date that is sent over JSON as "YYYY-MM-DD".
// An empty string decodes to the zero date, which means "no date": in a PATCH body
// `"birth_date": ""` clears the date, while `null` or an absent field leaves it unchanged.
type Date struct {
	time.Time
}

func NewDate(t time.Time) *Date {
	if t.IsZero() {
		return nil
	}
	return &Date{Time: t}
}

func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format(DateLayout))
}

func (d *Date) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("date must be a string in format YYYY-MM-DD: %w", err)
	}
	if raw == "" {
		d.Time = time.Time{}
		return nil
	}
	parsed, err := time.Parse(DateLayout, raw)
	if err != nil {
		return fmt.Errorf("date must be in format YYYY-MM-DD: %w", err)
	}
	d.Time = parsed
	return nil
}

// TimeOrZero returns the wrapped time, or the zero time for a nil date.
func (d *Date) TimeOrZero() time.Time {
	if d == nil {
		return time.Time{}
	}
	return d.Time
}
