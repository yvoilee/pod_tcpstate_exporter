// Copyright 2021 yvoilee.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/yvoilee/pod_tcpstate_exporter/docker"
	"log"
	"path"
	"strconv"
	"time"
)

func New(docker *docker.ClientWithCache, namespaces []string) *Collector {
	nss := make(map[string]struct{}, len(namespaces))
	for _, ns := range namespaces {
		nss[ns] = struct{}{}
	}
	return &Collector{
		docker:    docker,
		namespace: nss,
	}
}

type Collector struct {
	docker    *docker.ClientWithCache
	namespace map[string]struct{}
}

var connectionStatesDesc = prometheus.NewDesc(
	prometheus.BuildFQName("pod_exporter", "tcp", "connection_states"),
	"Number of connections by state, pod and namespace",
	[]string{"pod", "namespace", "state"},
	nil,
)

func (c *Collector) Describe(d chan<- *prometheus.Desc) {
	d <- connectionStatesDesc
}

func (c *Collector) Collect(m chan<- prometheus.Metric) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	sandboxes, err := c.docker.ListPodSandboxes(ctx, c.namespace)
	if err != nil {
		log.Printf("Error listing pod sandboxes: %+v", err)
		return
	}

	for _, s := range sandboxes {
		err = c.update(s, m)
		if err != nil {
			log.Printf("Error updating metrics values for pod %s: %+v", s.PodName, err)
			return
		}
	}
}

func (c *Collector) update(sandbox docker.PodSandbox, m chan<- prometheus.Metric) error {
	statsFile := path.Join("/proc", strconv.Itoa(sandbox.Pid), "net", "tcp")
	tcpStats, err := getTCPStats(statsFile)
	if err != nil {
		return fmt.Errorf("couldn't get tcpstats: %w", err)
	}

	for st, value := range tcpStats {
		m <- prometheus.MustNewConstMetric(connectionStatesDesc, prometheus.GaugeValue, value, sandbox.PodName, sandbox.Namespace, st.String())
	}

	return nil
}
