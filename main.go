package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type DockerStatsLine struct {
	Name     string `json:Name`
	BlockIO  string `json:"BlockIO"`
	CPUPerc  string `json:"CPUPerc"`
	MemPerc  string `json:"MemPerc"`
	MemUsage string `json:"MemUsage"`
	NetIO    string `json:NetIO`
}

type Metrics struct {
	registry        *prometheus.Registry
	blockIOInGauge  *prometheus.GaugeVec
	blockIOOutGauge *prometheus.GaugeVec
	cpuPercGauge    *prometheus.GaugeVec
	memPercGauge    *prometheus.GaugeVec
	memUsageGauge   *prometheus.GaugeVec
	memAllowedGauge *prometheus.GaugeVec
	netIOInGauge    *prometheus.GaugeVec
	netIOOutGauge   *prometheus.GaugeVec
}

func (m *Metrics) UpdateMetrics(stats DockerStatsLine) {
	blockIOIn, blockIOOut := strings.Split(stats.BlockIO, "/")[0], strings.Split(stats.BlockIO, "/")[1]
	blockIOInFloat := otherConvertToBytes(blockIOIn)
	m.blockIOInGauge.WithLabelValues(stats.Name).Set(float64(blockIOInFloat))
	m.blockIOOutGauge.WithLabelValues(stats.Name).Set(float64(otherConvertToBytes(blockIOOut)))

	cpuPercFloat, _ := strconv.ParseFloat(strings.TrimSuffix(stats.CPUPerc, "%"), 64)
	m.cpuPercGauge.WithLabelValues(stats.Name).Set(cpuPercFloat)

	memPercFloat, _ := strconv.ParseFloat(strings.TrimSuffix(stats.MemPerc, "%"), 64)
	m.memPercGauge.WithLabelValues(stats.Name).Set(memPercFloat)

	memUsageIn, memAllowed := strings.Split(stats.MemUsage, "/")[0], strings.Split(stats.MemUsage, "/")[1]
	memUsageInFloat := convertToBytes(memUsageIn)
	memAllowedFloat := convertToBytes(memAllowed)
	m.memUsageGauge.WithLabelValues(stats.Name).Set(float64(memUsageInFloat))
	m.memAllowedGauge.WithLabelValues(stats.Name).Set(float64(memAllowedFloat))

	netIOIn, netIOOut := strings.Split(stats.NetIO, "/")[0], strings.Split(stats.NetIO, "/")[1]
	netIOInFloat := otherConvertToBytes(netIOIn)
	netIOOutFloat := otherConvertToBytes(netIOOut)
	m.netIOInGauge.WithLabelValues(stats.Name).Set(float64(netIOInFloat))
	m.netIOOutGauge.WithLabelValues(stats.Name).Set(float64(netIOOutFloat))

}

func convertToBytes(input string) float64 {
	var bytes float64
	var err error
	input = strings.TrimSpace(input)
	if strings.HasSuffix(input, "GiB") {
		bytes, err = strconv.ParseFloat(strings.TrimSuffix(input, "GiB"), 64)
		bytes = math.Round(bytes * 1024 * 1024 * 1024)
	} else if strings.HasSuffix(input, "MiB") {
		bytes, err = strconv.ParseFloat(strings.TrimSuffix(input, "MiB"), 64)
		bytes = math.Round(bytes * 1024 * 1024)
	} else if strings.HasSuffix(input, "kiB") {
		bytes, err = strconv.ParseFloat(strings.TrimSuffix(input, "kiB"), 64)
		bytes = math.Round(bytes * 1024)
	} else if strings.HasSuffix(input, "B") {
		bytes, err = strconv.ParseFloat(strings.TrimSuffix(input, "B"), 64)
	}

	if err != nil {
		panic(err)
	}
	return bytes
}

func otherConvertToBytes(input string) float64 {
	input = strings.TrimSpace(input)
	var bytes float64
	var err error
	if strings.HasSuffix(input, "GB") {
		bytes, err = strconv.ParseFloat(strings.TrimSuffix(input, "GB"), 64)
		bytes = bytes * 1000 * 1000 * 1000

	} else if strings.HasSuffix(input, "MB") {
		bytes, err = strconv.ParseFloat(strings.TrimSuffix(input, "MB"), 64)
		bytes = bytes * 1000 * 1000
	} else if strings.HasSuffix(input, "kB") {
		bytes, err = strconv.ParseFloat(strings.TrimSuffix(input, "kB"), 64)
		bytes = bytes * 1000
	} else if strings.HasSuffix(input, "B") {
		bytes, err = strconv.ParseFloat(strings.TrimSuffix(input, "B"), 64)
	}

	if err != nil {
		fmt.Println("Error converting to bytes", input)
		panic(err)
	}
	return bytes
}

func ParseDockerStatsLine(line string) (DockerStatsLine, error) {
	var stats DockerStatsLine
	err := json.Unmarshal([]byte(line), &stats)
	if err != nil {
		return stats, err
	}
	return stats, nil
}

func GetDockerStats() []string {
	out, err := exec.Command("docker", "stats", "--no-stream", "--format", "{{json .}}").Output()
	if err != nil {
		panic(err)
	}
	return strings.Split(string(out), "\n")
}

func InitMetrics() Metrics {
	registry := prometheus.NewRegistry()
	blockIOInGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "docker_block_io_in",
		Help: "Docker stats",
	},
		[]string{"container_name"})
	registry.MustRegister(blockIOInGauge)

	blockIOOutGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "docker_block_io_out",
		Help: "Docker stats",
	},
		[]string{"container_name"})
	registry.MustRegister(blockIOOutGauge)

	cpuPercGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "docker_cpu_perc",
		Help: "Docker stats",
	},
		[]string{"container_name"})
	registry.MustRegister(cpuPercGauge)

	memPercGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "docker_mem_perc",
		Help: "Docker stats",
	},
		[]string{"container_name"})
	registry.MustRegister(memPercGauge)

	memUsageGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "docker_mem_usage",
		Help: "Docker stats",
	},
		[]string{"container_name"})
	registry.MustRegister(memUsageGauge)

	memAllowedGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "docker_mem_allowed",
		Help: "Docker stats",
	},
		[]string{"container_name"})
	registry.MustRegister(memAllowedGauge)

	netIOInGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "docker_net_io_in",
		Help: "Docker stats",
	},
		[]string{"container_name"})
	registry.MustRegister(netIOInGauge)

	netIOOutGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "docker_net_io_out",
		Help: "Docker stats",
	},
		[]string{"container_name"})
	registry.MustRegister(netIOOutGauge)

	return Metrics{
		registry:        registry,
		blockIOInGauge:  blockIOInGauge,
		blockIOOutGauge: blockIOOutGauge,
		cpuPercGauge:    cpuPercGauge,
		memPercGauge:    memPercGauge,
		memUsageGauge:   memUsageGauge,
		memAllowedGauge: memAllowedGauge,
		netIOInGauge:    netIOInGauge,
		netIOOutGauge:   netIOOutGauge,
	}
}

func main() {
	metrics := InitMetrics()
	go func() {
		for {
			stats := GetDockerStats()
			for _, line := range stats {
				if line == "" {
					continue
				}
				stats, err := ParseDockerStatsLine(line)
				if err != nil {
					panic(err)
				}
				metrics.UpdateMetrics(stats)
			}
			time.Sleep(2 * time.Second)
		}
	}()

	http.Handle("/metrics", promhttp.HandlerFor(metrics.registry, promhttp.HandlerOpts{Registry: metrics.registry}))
	log.Fatal(http.ListenAndServe(":9200", nil))
}
