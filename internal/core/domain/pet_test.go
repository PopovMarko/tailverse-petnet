package core_domain

import (
	"testing"
	"time"
)

func TestPetAge(t *testing.T) {
	date := func(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }

	cases := []struct {
		birth, now time.Time
		want       int
	}{
		{date(2020, 5, 10), date(2026, 5, 9), 5},
		{date(2020, 5, 10), date(2026, 5, 10), 6},
		{date(2020, 2, 29), date(2026, 2, 28), 5},
		{date(2020, 2, 29), date(2026, 3, 1), 6},
		{date(2026, 1, 1), date(2026, 9, 28), 0},
	}
	for _, tc := range cases {
		got := Pet{BirthDate: tc.birth}.Age(tc.now)
		if got == nil || *got != tc.want {
			t.Errorf("Age(birth %s, now %s) = %v, want %d", tc.birth.Format(time.DateOnly), tc.now.Format(time.DateOnly), got, tc.want)
		}
	}

	if age := (Pet{}).Age(time.Now()); age != nil {
		t.Errorf("unknown birth date: age = %d, want nil", *age)
	}
}
