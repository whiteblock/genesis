package services

import (
	"strconv"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/ssh/mocks"
	"github.com/whiteblock/genesis/state"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
)

func TestPrometheusService_Prepare(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mocks.NewMockClient(ctrl)
	details := db.DeploymentDetails{
		ID:           "123",
		Servers:      []int{4, 5, 6},
		Blockchain:   "eos",
		Nodes:        100,
		Images:       []string{},
		Params:       map[string]interface{}{"prometheusInstrumentationPort": "4000", "isThisATest?": "yes"},
		Resources:    []util.Resources{{Cpus: "", Memory: "", Volumes: []string{}, Ports: []string{}}},
		Environments: []map[string]string{},
		Files:        []map[string]string{},
		Logs:         []map[string]string{},
		Extras:       map[string]interface{}{},
	}
	tn := testnet.TestNet{
		TestNetID:          "10",
		Servers:            []db.Server{},
		Nodes:              []db.Node{},
		NewlyBuiltNodes:    []db.Node{},
		SideCars:           [][]db.SideCar{},
		NewlyBuiltSideCars: [][]db.SideCar{},
		Clients:            map[int]ssh.Client{0: client},
		BuildState:         state.NewBuildState([]int{}, "0"),
		Details:            []db.DeploymentDetails{details},
		CombinedDetails:    details,
		LDD:                &details,
	}
	promServ := PrometheusService{}

	client.EXPECT().Scp("test", "/tmp/prometheus.yml")

	err := promServ.Prepare(client, &tn)
	if err != nil {
		t.Error("return value of Prepare does not match expected value")
	}
}

func Test_port(t *testing.T) {
	var tests = []struct {
		params    map[string]interface{}
		nodeIndex int
		expected  string
	}{
		{params:    map[string]interface{}{"prometheusInstrumentationPort": []interface{}{"4000"}}, nodeIndex: 0, expected:  "4000",
		},
		{
			params:    map[string]interface{}{"prometheusInstrumentationPort": "3000"},
			nodeIndex: 0,
			expected:  "3000",
		},
		{
			params:    map[string]interface{}{"prometheusInstrumentationPort": []interface{}{"4000", "2000", "8888"}},
			nodeIndex: 1,
			expected:  "2000",
		},
		{
			params:    map[string]interface{}{},
			nodeIndex: 0,
			expected:  "8008",
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
