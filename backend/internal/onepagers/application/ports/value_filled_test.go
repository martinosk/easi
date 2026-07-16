package ports_test

import (
	"testing"
	"time"

	"easi/backend/internal/onepagers/application/ports"

	"github.com/stretchr/testify/assert"
)

func TestValueFilled(t *testing.T) {
	cases := []struct {
		name  string
		value ports.BuiltInFieldValue
		want  bool
	}{
		{"absent value is not filled", nil, false},
		{"text value is filled", ports.TextValue{Text: "Handles payments"}, true},
		{"date value is filled", ports.DateValue{Date: time.Now()}, true},
		{"maturity value is filled", ports.MaturityValue{Value: 0}, true},
		{"experts with entries is filled", ports.ExpertsValue{Experts: []ports.Expert{{Name: "Alice"}}}, true},
		{"experts with empty list is not filled", ports.ExpertsValue{Experts: nil}, false},
		{"references with entries is filled", ports.ReferenceListValue{References: []ports.Reference{{ID: "cap-1", Label: "Billing"}}}, true},
		{"references with empty list is not filled", ports.ReferenceListValue{References: nil}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, ports.ValueFilled(tc.value))
		})
	}
}
