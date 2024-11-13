package server

import (
	"fmt"
	"time"

	"github.com/99designs/gqlgen/graphql"
)

func marshalTime(t time.Time) graphql.Marshaler {
	return graphql.MarshalTime(t)
}

func unmarshalTime(v interface{}) (time.Time, error) {
	if timeStr, ok := v.(string); ok {
		return time.Parse(time.RFC3339, timeStr)
	}
	return time.Time{}, fmt.Errorf("time should be a string")
} 