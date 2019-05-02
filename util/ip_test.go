package util

import (
	"testing"
)

type strXErr struct {
	str string
	err bool
}

func TestGetInfoFromIP(t *testing.T) {
	conf.ServerBits = 8
	conf.NodeBits = 2
	conf.ClusterBits = 14
	tests := map[string][]int{
		"10.1.0.2":  {1, 0, 0},
		"10.27.0.6": {27, 1, 0},
		"10.0.0.10": {0, 2, 0},
	}
	for test, expected := range tests {
		server, node, index := GetInfoFromIP(test)
		if server != expected[0] || node != expected[1] || index != expected[2] {
			t.Errorf("GetInfoFromIP(\"%s\") returned server=%d,node=%d,index=%d."+
				" Expected server=%d,node=%d,index%d", test, server, node, index, expected[0], expected[1], expected[2])
		}
	}
}

type getNodeIPTest struct {
	params   []int
	expected strXErr
}

func TestGetNodeIP(t *testing.T) {
	conf.ServerBits = 8
	conf.NodeBits = 4
	conf.ClusterBits = 12
	conf.IPPrefix = 10
	tests := []getNodeIPTest{
		//Normal Nodes
		getNodeIPTest{params: []int{1, 0, 0}, expected: strXErr{str: "10.1.0.2", err: false}},
		getNodeIPTest{params: []int{27, 1, 0}, expected: strXErr{str: "10.27.0.18", err: false}},
		getNodeIPTest{params: []int{0, 2, 0}, expected: strXErr{str: "10.0.0.34", err: false}},
		//Side Cars (Valid)
		getNodeIPTest{params: []int{1, 0, 2}, expected: strXErr{str: "10.1.0.4", err: false}},
		getNodeIPTest{params: []int{27, 1, 1}, expected: strXErr{str: "10.27.0.19", err: false}},
		getNodeIPTest{params: []int{0, 2, 3}, expected: strXErr{str: "10.0.0.37", err: false}},

		//Side Cars (Invalid)
		getNodeIPTest{params: []int{1, 0, 16}, expected: strXErr{str: "", err: true}},
		getNodeIPTest{params: []int{27, 1, 25}, expected: strXErr{str: "", err: true}},
		getNodeIPTest{params: []int{0, 2, 22}, expected: strXErr{str: "", err: true}},
	}

	for _, test := range tests {
		ip, err := GetNodeIP(test.params[0], test.params[1], test.params[2])
		success := true
		if err != nil && !test.expected.err {
			success = false
		} else if err == nil && test.expected.err {
			success = false
		} else if ip != test.expected.str {
			success = false
		}

		if !success {
			t.Errorf("GetNodeIP(%d,%d,%d) returned {%s,%v}. Expected {%s,%v}\n", test.params[0],
				test.params[1], test.params[2], ip, err, test.expected.str, test.expected.err)
		}
	}
}

func TestGetGateway(t *testing.T) {
	conf.ServerBits = 8
	conf.NodeBits = 4
	conf.ClusterBits = 12
	conf.IPPrefix = 10
	tests := []getNodeIPTest{
		//Normal Nodes
		getNodeIPTest{params: []int{1, 0}, expected: strXErr{str: "10.1.0.1", err: false}},
		getNodeIPTest{params: []int{27, 1}, expected: strXErr{str: "10.27.0.17", err: false}},
		getNodeIPTest{params: []int{0, 2}, expected: strXErr{str: "10.0.0.33", err: false}},
	}

	for _, test := range tests {
		ip := GetGateway(test.params[0], test.params[1])

		if ip != test.expected.str {
			t.Errorf("GetGateway(%d,%d) returned %s. Expected %s\n", test.params[0],
				test.params[1], ip, test.expected.str)
		}
	}
}

func TestGetNetworkAddress(t *testing.T) {
	conf.ServerBits = 8
	conf.NodeBits = 4
	conf.ClusterBits = 12
	conf.IPPrefix = 10
	tests := []getNodeIPTest{
		//Normal Nodes
		getNodeIPTest{params: []int{1, 0}, expected: strXErr{str: "10.1.0.0/28", err: false}},
		getNodeIPTest{params: []int{1, 1}, expected: strXErr{str: "10.1.0.16/28", err: false}},
		getNodeIPTest{params: []int{27, 1}, expected: strXErr{str: "10.27.0.16/28", err: false}},
		getNodeIPTest{params: []int{0, 2}, expected: strXErr{str: "10.0.0.32/28", err: false}},
	}

	for _, test := range tests {
		ip := GetNetworkAddress(test.params[0], test.params[1])

		if ip == test.expected.str {
			t.Errorf("GetNetworkAddress(%d,%d) returned %s. Expected %s\n", test.params[0],
				test.params[1], ip, test.expected.str)
		}
	}
}
