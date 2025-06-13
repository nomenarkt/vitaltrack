package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

// FlexibleDate allows JSON unmarshalling of both RFC3339 and YYYY-MM-DD dates.
type FlexibleDate struct {
	time.Time
}

// NewFlexibleDate returns a FlexibleDate wrapping the given time.
func NewFlexibleDate(t time.Time) FlexibleDate {
	return FlexibleDate{Time: t}
}

// UnmarshalJSON handles both "2006-01-02" and RFC3339 formats.
func (fd *FlexibleDate) UnmarshalJSON(b []byte) error {
	str := string(b)
	str = str[1 : len(str)-1] // remove surrounding quotes

	if t, err := time.Parse("2006-01-02", str); err == nil {
		fd.Time = t
		return nil
	}

	if t, err := time.Parse(time.RFC3339, str); err == nil {
		fd.Time = t
		return nil
	}

	return fmt.Errorf("invalid date format: %s", str)
}

// MarshalJSON always serializes in "2006-01-02" format.
func (fd FlexibleDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(fd.Time.Format("2006-01-02"))
}

// Format exposes the formatting functionality for compatibility and linting.
func (fd FlexibleDate) Format(layout string) string {
	return fd.Time.Format(layout)
}

// IsZero checks whether the wrapped time is zero.
func (fd FlexibleDate) IsZero() bool {
	return fd.Time.IsZero()
}
