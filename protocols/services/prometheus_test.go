package services

import (
	"strconv"
	"testing"
)

func Test_port(t *testing.T) {
	var tests = []struct {
		params map[string]interface{}
		nodeIndex int
		expected string
	}{
		{
			params: map[string]interface{}{"prometheusInstrumentationPort": []interface{}{"4000"}},
			nodeIndex: 0,
			expected: "4000",
		},
		{
			params: map[string]interface{}{"prometheusInstrumentationPort": "3000"},
			nodeIndex: 0,
			expected: "3000",
		},
		{
			params: map[string]interface{}{"prometheusInstrumentationPort": []interface{}{"4000", "2000", "8888"}},
			nodeIndex: 1,
			expected: "2000",
		},
		{
			params: map[string]interface{}{},
			nodeIndex: 0,
			expected: "8008",
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if port(tt.params, tt.nodeIndex) != tt.expected {
				t.Error("return value of port() did not match expected value")
			}
		})
	}
}
