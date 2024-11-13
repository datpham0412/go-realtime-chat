package server

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/99designs/gqlgen/graphql"
)

// Time is a custom scalar for handling time.Time
type Time struct {
	time.Time
}

// MarshalGQL implements the graphql.Marshaler interface
func (t Time) MarshalGQL(w io.Writer) {
	graphql.MarshalString(t.Time.Format(time.RFC3339)).MarshalGQL(w)
}

// UnmarshalGQL implements the graphql.Unmarshaler interface
func (t *Time) UnmarshalGQL(v interface{}) error {
	switch v := v.(type) {
	case string:
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return err
		}
		t.Time = parsed
		return nil
	case time.Time:
		t.Time = v
		return nil
	default:
		return fmt.Errorf("wrong type for Time: %T", v)
	}
}

// MarshalJSON implements the json.Marshaler interface
func (t Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Time.Format(time.RFC3339))
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (t *Time) UnmarshalJSON(b []byte) error {
	var timeStr string
	if err := json.Unmarshal(b, &timeStr); err != nil {
		return err
	}

	parsed, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return err
	}

	t.Time = parsed
	return nil
}
