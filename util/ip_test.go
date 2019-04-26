package util

import (
	"testing"
)

func TestGetInfoFromIP(t *testing.T) {
	conf.ServerBits = 8
	conf.NodeBits = 2
	conf.ClusterBits = 14
	tests := map[string][]int{
		"10.1.0.2":  {1, 0},
		"10.27.0.6": {27, 1},
		"10.0.0.10": {0, 2},
	}
	for test, expected := range tests {
		server, node := GetInfoFromIP(test)
		if server != expected[0] || node != expected[1] {
			t.Errorf("GetInfoFromIP(\"%s\") returned server=%d,node=%d. Expected server=%d,node=%d", test, server, node, expected[0], expected[1])
		}
	}
}
