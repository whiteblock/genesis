package services

import (
	"bytes"
	"fmt"
	"html/template"
	"reflect"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/genesis/db"
	"github.com/whiteblock/genesis/protocols/helpers"
	"github.com/whiteblock/genesis/ssh"
	"github.com/whiteblock/genesis/testnet"
	"github.com/whiteblock/genesis/util"
)

// PrometheusService represents the Prometheus service
type PrometheusService struct {
	SimpleService
}

// Prepare prepares the prometheus service
func (p PrometheusService) Prepare(client ssh.Client, tn *testnet.TestNet) error {

	configTxt := "scrape_configs:\n"
	for nodeIndex, node := range tn.Nodes {
		tmpl, err := template.New("prometheus-source").Parse(`
- job_name:       '{{.Tn.LDD.Blockchain}}-{{.Node.ID}}-{{.Node.IP}}'
  scrape_interval: 5s
  metrics_path: /metrics
  static_configs:
    - targets: ['{{.Node.IP}}:{{.InstrumentationPort}}']
      labels:
        blockchain: '{{.Tn.LDD.Blockchain}}'
        testnet: '{{.Tn.TestNetID}}'
        ip: '{{.Node.IP}}'

`)

		if err != nil {
			return util.LogError(err)
		}


		prometheusInstrumentationPort := port(tn.CombinedDetails.Params, nodeIndex)

		var tpl bytes.Buffer
		if err = tmpl.Execute(&tpl, struct {
			Tn                  *testnet.TestNet
			Node                db.Node
			Conf                *util.Config
			InstrumentationPort string
		}{tn, node, conf, prometheusInstrumentationPort}); err != nil {
			log.Error(err)
		} else {
			configTxt += tpl.String()
		}

	}
	log.Debug(configTxt)
	log.Debug(conf.PrometheusConfig)

	tmpFilename, err := util.GetUUIDString()
	if err != nil {
		return util.LogError(err)
	}

	if _, exists := tn.CombinedDetails.Params["isThisATest?"]; exists {
		tmpFilename = "test"
	}

	err = tn.BuildState.Write(tmpFilename, configTxt)
	if err != nil {
		return util.LogError(err)
	}

	return helpers.CopyAllToServers(tn, tmpFilename, conf.PrometheusConfig)
}

func port(params map[string]interface{}, nodeIndex int) string {
	prometheusInstrumentationPort := "8008"

	obj := params["prometheusInstrumentationPort"]
	if obj != nil {
		promPorts, ok := obj.([]interface{})
		if ok {
			prometheusInstrumentationPort = fmt.Sprintf("%v", promPorts[nodeIndex])
		} else if reflect.TypeOf(obj).Kind() == reflect.String {
			prometheusInstrumentationPort = obj.(string)
		}
	}

	return prometheusInstrumentationPort
}

// RegisterPrometheus exposes a Prometheus service on the testnet.
func RegisterPrometheus() Service {
	return PrometheusService{
		SimpleService{
			Name:    "prometheus",
			Image:   "prom/prometheus",
			Env:     map[string]string{},
			Ports:   []string{strconv.Itoa(conf.PrometheusPort) + ":9090"},
			Volumes: []string{conf.PrometheusConfig + ":/etc/prometheus/prometheus.yml"},
		},
	}
}
