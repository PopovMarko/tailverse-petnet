package core_http_types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDateJSON(t *testing.T) {
	var body struct {
		BirthDate *Date `json:"birth_date"`
	}

	if err := json.Unmarshal([]byte(`{"birth_date":"2021-04-10"}`), &body); err != nil {
		t.Fatal(err)
	}
	if body.BirthDate == nil || !body.BirthDate.Equal(time.Date(2021, 4, 10, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("date = %v, want 2021-04-10", body.BirthDate)
	}
	if out, _ := json.Marshal(body); string(out) != `{"birth_date":"2021-04-10"}` {
		t.Errorf("marshal = %s", out)
	}

	// "" is "no date" (PATCH clears the date), null leaves the pointer nil (PATCH keeps the date).
	body.BirthDate = nil
	if err := json.Unmarshal([]byte(`{"birth_date":""}`), &body); err != nil {
		t.Fatalf(`"" must decode: %v`, err)
	}
	if body.BirthDate == nil || !body.BirthDate.IsZero() || body.BirthDate.TimeOrZero() != (time.Time{}) {
		t.Errorf(`"" = %v, want a non-nil zero date`, body.BirthDate)
	}
	body.BirthDate = nil
	if err := json.Unmarshal([]byte(`{"birth_date":null}`), &body); err != nil || body.BirthDate != nil {
		t.Errorf("null = %v (err %v), want nil", body.BirthDate, err)
	}

	for _, bad := range []string{`"10.04.2021"`, `"2021-13-01"`, `20210410`} {
		if err := json.Unmarshal([]byte(`{"birth_date":`+bad+`}`), &body); err == nil {
			t.Errorf("%s must be rejected", bad)
		}
	}
}
