package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
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

func (m *Metrics) InitMetrics() {
	registry := prometheus.NewRegistry()
	blockIOInGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "docker_block_io_in_bytes",
		Help: "Docker stats",
	},
		[]string{"container_name"})
	registry.MustRegister(blockIOInGauge)

	blockIOOutGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "docker_block_io_out_bytes",
		Help: "Docker stats",
	},
		[]string{"container_name"})
	registry.MustRegister(blockIOOutGauge)

	cpuPercGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "docker_cpu_percentage",
		Help: "Docker stats",
	},
		[]string{"container_name"})
	registry.MustRegister(cpuPercGauge)

	memPercGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "docker_memory_percentage",
		Help: "Docker stats",
	},
		[]string{"container_name"})
	registry.MustRegister(memPercGauge)

	memUsageGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "docker_memory_usage_bytes",
		Help: "Docker stats",
	},
		[]string{"container_name"})
	registry.MustRegister(memUsageGauge)

	memAllowedGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "docker_memory_allowed_bytes",
		Help: "Docker stats",
	},
		[]string{"container_name"})
	registry.MustRegister(memAllowedGauge)

	netIOInGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "docker_net_io_in_bytes",
		Help: "Docker stats",
	},
		[]string{"container_name"})
	registry.MustRegister(netIOInGauge)

	netIOOutGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "docker_net_io_out_bytes",
		Help: "Docker stats",
	},
		[]string{"container_name"})
	registry.MustRegister(netIOOutGauge)

	m.registry = registry
	m.blockIOInGauge = blockIOInGauge
	m.blockIOOutGauge = blockIOOutGauge
	m.cpuPercGauge = cpuPercGauge
	m.memPercGauge = memPercGauge
	m.memUsageGauge = memUsageGauge
	m.memAllowedGauge = memAllowedGauge
	m.netIOInGauge = netIOInGauge
	m.netIOOutGauge = netIOOutGauge

}

func (m *Metrics) UpdateMetrics(stats DockerStatsLine) {
	blockIOIn, blockIOOut := strings.Split(stats.BlockIO, "/")[0], strings.Split(stats.BlockIO, "/")[1]
	blockIOInFloat, err := convertBase10ToBytes(blockIOIn)
	if err != nil {
		log.Printf("Error converting blockIOIn to bytes: %s", err)
	}
	blockIOOutFloat, err := convertBase10ToBytes(blockIOOut)
	if err != nil {
		log.Printf("Error converting blockIOOut to bytes: %s", err)
	}
	m.blockIOInGauge.WithLabelValues(stats.Name).Set(float64(blockIOInFloat))
	m.blockIOOutGauge.WithLabelValues(stats.Name).Set(float64(blockIOOutFloat))

	cpuPercFloat, err := strconv.ParseFloat(strings.TrimSuffix(stats.CPUPerc, "%"), 64)
	if err != nil {
		log.Printf("Error converting CPUPerc to float: %s", err)
	}
	m.cpuPercGauge.WithLabelValues(stats.Name).Set(cpuPercFloat)

	memPercFloat, err := strconv.ParseFloat(strings.TrimSuffix(stats.MemPerc, "%"), 64)
	m.memPercGauge.WithLabelValues(stats.Name).Set(memPercFloat)

	memUsageIn, memAllowed := strings.Split(stats.MemUsage, "/")[0], strings.Split(stats.MemUsage, "/")[1]
	memUsageInFloat, err := convertBase2ToBytes(memUsageIn)
	if err != nil {
		log.Printf("Error converting memUsageIn to bytes: %s", err)
	}
	memAllowedFloat, err := convertBase2ToBytes(memAllowed)
	if err != nil {
		log.Printf("Error converting memAllowed to bytes: %s", err)
	}
	m.memUsageGauge.WithLabelValues(stats.Name).Set(float64(memUsageInFloat))
	m.memAllowedGauge.WithLabelValues(stats.Name).Set(float64(memAllowedFloat))

	netIOIn, netIOOut := strings.Split(stats.NetIO, "/")[0], strings.Split(stats.NetIO, "/")[1]
	netIOInFloat, err := convertBase10ToBytes(netIOIn)
	if err != nil {
		log.Printf("Error converting netIOIn to bytes: %s", err)
	}
	netIOOutFloat, err := convertBase10ToBytes(netIOOut)
	if err != nil {
		log.Printf("Error converting netIOOut to bytes: %s", err)
	}
	m.netIOInGauge.WithLabelValues(stats.Name).Set(float64(netIOInFloat))
	m.netIOOutGauge.WithLabelValues(stats.Name).Set(float64(netIOOutFloat))

}

func convertBase2ToBytes(input string) (float64, error) {
	var bytes float64
	var err error
	input = strings.TrimSpace(input)
	if strings.HasSuffix(input, "TiB") {
		bytes, err = strconv.ParseFloat(strings.TrimSuffix(input, "TiB"), 64)
		bytes = math.Round(bytes * math.Pow(1024, 4))
	} else if strings.HasSuffix(input, "GiB") {
		bytes, err = strconv.ParseFloat(strings.TrimSuffix(input, "GiB"), 64)
		bytes = math.Round(bytes * math.Pow(1024, 3))
	} else if strings.HasSuffix(input, "MiB") {
		bytes, err = strconv.ParseFloat(strings.TrimSuffix(input, "MiB"), 64)
		bytes = math.Round(bytes * math.Pow(1024, 2))
	} else if strings.HasSuffix(input, "kiB") {
		bytes, err = strconv.ParseFloat(strings.TrimSuffix(input, "kiB"), 64)
		bytes = math.Round(bytes * math.Pow(1024, 1))
	} else if strings.HasSuffix(input, "B") {
		bytes, err = strconv.ParseFloat(strings.TrimSuffix(input, "B"), 64)
	}

	return bytes, err
}

func convertBase10ToBytes(input string) (float64, error) {
	var bytes float64
	var err error
	input = strings.TrimSpace(input)
	if strings.HasSuffix(input, "TB") {
		bytes, err = strconv.ParseFloat(strings.TrimSuffix(input, "TB"), 64)
		bytes = math.Round(bytes * math.Pow(1000, 4))
	} else if strings.HasSuffix(input, "GB") {
		bytes, err = strconv.ParseFloat(strings.TrimSuffix(input, "GB"), 64)
		bytes = math.Round(bytes * math.Pow(1000, 3))
	} else if strings.HasSuffix(input, "MB") {
		bytes, err = strconv.ParseFloat(strings.TrimSuffix(input, "MB"), 64)
		bytes = math.Round(bytes * math.Pow(1000, 2))
	} else if strings.HasSuffix(input, "kB") {
		bytes, err = strconv.ParseFloat(strings.TrimSuffix(input, "kB"), 64)
		bytes = math.Round(bytes * math.Pow(1000, 1))
	} else if strings.HasSuffix(input, "B") {
		bytes, err = strconv.ParseFloat(strings.TrimSuffix(input, "B"), 64)
	}
	return bytes, err
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


func main() {
	metrics := Metrics{}	
	metrics.InitMetrics()
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
	port := os.Getenv("PORT")
	if port == "" {
		port = "9200"
	}
	fmt.Printf("Listening on port %s\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
