package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yvoilee/pod_tcpstate_exporter/collector"
	"github.com/yvoilee/pod_tcpstate_exporter/docker"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	dockerCli, err := docker.New()
	if err != nil {
		panic(err)
	}
	statsCollector := collector.New(&dockerCli, strings.Split(os.Getenv("NAMESPACES"), ","))
	prometheus.MustRegister(statsCollector)

	http.Handle("/metrics", promhttp.Handler())
	log.Println("Listening on port 8080...")
	_ = http.ListenAndServe(":8080", nil)
}
