// Copyright ProbeLab
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package enacollector exports the allowance counters of the AWS Elastic
// Network Adapter (ENA) driver. The ECS task metadata API does not serve these
// values, so the collector reads them with ethtool from the network interfaces
// that the process can see. Ref:
// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/monitoring-network-performance-ena.html
package enacollector

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

// enaDriver is the driver name that ethtool reports for ENA interfaces.
const enaDriver = "ena"

var interfaceLabels = []string{"interface"}

// allowanceMetrics maps ethtool statistic names to the metrics they are
// exported as. An interface only reports the statistics that its driver
// version and instance type support, so each one is optional.
var allowanceMetrics = []struct {
	stat      string
	valueType prometheus.ValueType
	desc      *prometheus.Desc
}{
	{
		stat:      "bw_in_allowance_exceeded",
		valueType: prometheus.CounterValue,
		desc: prometheus.NewDesc(
			"ecs_network_bw_in_allowance_exceeded_total",
			"Number of packets queued or dropped because the inbound aggregate bandwidth exceeded the maximum for the instance.",
			interfaceLabels, nil),
	},
	{
		stat:      "bw_out_allowance_exceeded",
		valueType: prometheus.CounterValue,
		desc: prometheus.NewDesc(
			"ecs_network_bw_out_allowance_exceeded_total",
			"Number of packets queued or dropped because the outbound aggregate bandwidth exceeded the maximum for the instance.",
			interfaceLabels, nil),
	},
	{
		stat:      "pps_allowance_exceeded",
		valueType: prometheus.CounterValue,
		desc: prometheus.NewDesc(
			"ecs_network_pps_allowance_exceeded_total",
			"Number of packets queued or dropped because the bidirectional packets per second exceeded the maximum for the instance.",
			interfaceLabels, nil),
	},
	{
		stat:      "conntrack_allowance_exceeded",
		valueType: prometheus.CounterValue,
		desc: prometheus.NewDesc(
			"ecs_network_conntrack_allowance_exceeded_total",
			"Number of packets dropped because connection tracking exceeded the maximum for the instance and new connections could not be established.",
			interfaceLabels, nil),
	},
	{
		stat:      "conntrack_allowance_available",
		valueType: prometheus.GaugeValue,
		desc: prometheus.NewDesc(
			"ecs_network_conntrack_allowance_available",
			"Number of tracked connections that the instance can establish before it reaches the connections tracked allowance of its instance type. Only available on Nitro-based instances.",
			interfaceLabels, nil),
	},
	{
		stat:      "linklocal_allowance_exceeded",
		valueType: prometheus.CounterValue,
		desc: prometheus.NewDesc(
			"ecs_network_linklocal_allowance_exceeded_total",
			"Number of packets dropped because the packets per second to local proxy services (Amazon DNS, instance metadata, Amazon Time Sync) exceeded the maximum for the network interface.",
			interfaceLabels, nil),
	},
}

// statsReader reads driver statistics of network interfaces.
type statsReader interface {
	// Interfaces returns the names of the network interfaces to check.
	Interfaces() ([]string, error)
	// DriverName returns the name of the driver of the interface.
	DriverName(iface string) (string, error)
	// Stats returns the ethtool statistics of the interface.
	Stats(iface string) (map[string]uint64, error)
	// Close releases the resources of the reader.
	Close()
}

// NewCollector returns a new Collector that exports the ENA allowance
// counters of all ENA interfaces. It exports no metrics on hosts without ENA
// interfaces, for example outside of AWS or on operating systems other than
// Linux.
func NewCollector(logger *slog.Logger) prometheus.Collector {
	return &collector{newReader: newEthtoolReader, logger: logger}
}

type collector struct {
	newReader func() (statsReader, error)
	logger    *slog.Logger
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range allowanceMetrics {
		ch <- metric.desc
	}
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	reader, err := c.newReader()
	if err != nil {
		c.logger.Debug("Failed to open ethtool", "error", err)
		return
	}
	defer reader.Close()

	ifaces, err := reader.Interfaces()
	if err != nil {
		c.logger.Debug("Failed to list network interfaces", "error", err)
		return
	}

	for _, iface := range ifaces {
		driver, err := reader.DriverName(iface)
		if err != nil {
			// Interfaces without a driver, for example loopback, end up here.
			c.logger.Debug("Failed to get interface driver", "interface", iface, "error", err)
			continue
		}
		if driver != enaDriver {
			continue
		}

		stats, err := reader.Stats(iface)
		if err != nil {
			c.logger.Debug("Failed to get interface stats", "interface", iface, "error", err)
			continue
		}

		for _, metric := range allowanceMetrics {
			value, ok := stats[metric.stat]
			if !ok {
				continue
			}
			ch <- prometheus.MustNewConstMetric(metric.desc, metric.valueType, float64(value), iface)
		}
	}
}
