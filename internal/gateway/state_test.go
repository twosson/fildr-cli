package gateway

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetMetric(t *testing.T) {
	tests := []MetricData{
		{
			instance: "hostname",
			job:      "node",
		},
		{
			instance: "hostname1",
			job:      "node1",
		},
	}

	for i := range tests {
		push(&tests[i])
	}
	assert.Equal(t, "hostname", pop().instance)
	assert.Equal(t, "hostname1", pop().instance)
}
