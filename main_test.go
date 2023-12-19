package main

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestParseDockerStatsLine(t *testing.T) {
	input := `{"BlockIO":"13.5MB / 0B","CPUPerc":"0.00%","Container":"short_id","ID":"full_id","MemPerc":"0.02%","MemUsage":"9.809MiB / 62.3GiB","Name":"container_name","NetIO":"0B / 0B","PIDs":"5"}`
	expected := DockerStatsLine{
		Name: "container_name", BlockIO: "13.5MB / 0B",
		CPUPerc:  "0.00%",
		MemPerc:  "0.02%",
		MemUsage: "9.809MiB / 62.3GiB",
		NetIO:    "0B / 0B",
	}

	actual, err := ParseDockerStatsLine(input)
	if err != nil {
		t.Errorf("Error parsing docker stats line: %s", err)
	}
	if actual.Name != expected.Name {
		t.Errorf("Name: expected %s, got %s", expected.Name, actual.Name)
	}
	if actual.BlockIO != expected.BlockIO {
		t.Errorf("BlockIO: expected %s, got %s", expected.BlockIO, actual.BlockIO)
	}
	if actual.CPUPerc != expected.CPUPerc {
		t.Errorf("CPUPerc: expected %s, got %s", expected.CPUPerc, actual.CPUPerc)
	}
	if actual.MemPerc != expected.MemPerc {
		t.Errorf("MemPerc: expected %s, got %s", expected.MemPerc, actual.MemPerc)
	}
	if actual.MemUsage != expected.MemUsage {
		t.Errorf("MemUsage: expected %s, got %s", expected.MemUsage, actual.MemUsage)
	}
	if actual.NetIO != expected.NetIO {
		t.Errorf("NetIO: expected %s, got %s", expected.NetIO, actual.NetIO)
	}

}

func TestGetDockerStats(t *testing.T) {
	stats := GetDockerStats()
	if len(stats) == 0 {
		t.Errorf("GetDockerStats returned no stats")
	}

	firstLine := stats[0]
	_, err := ParseDockerStatsLine(firstLine)
	if err != nil {
		t.Errorf("Error parsing docker stats line: %s", err)
	}
}


func TestConvertToBytes(t *testing.T) {
	t.Run("convertBase10ToBytes", func(t *testing.T) {
		inputs := []string{"10B", "10kB", "10MB", "10GB", "10TB"}
		expected := []float64{10, 10000, 10000000, 10000000000, 10000000000000}
		for i, input := range inputs {
			actual, err := convertBase10ToBytes(input)
			if err != nil {
				t.Errorf("Error converting base10 to bytes: %s", err)
			}
			if actual != expected[i] {
				t.Errorf("ConvertBase10ToBytes: expected %f, got %f", expected[i], actual)
			}
		}
	})
	t.Run("convertBase2ToBytes", func(t *testing.T) {
		inputs := []string{"10B", "10KiB", "10MiB", "10GiB", "10TiB"}
		expected := []float64{10, 10240, 10485760, 10737418240, 10995116277760}
		for i, input := range inputs {
			actual, err := convertBase2ToBytes(input)
			if err != nil {
				t.Errorf("Error converting base2 to bytes: %s", err)
			}
			if actual != expected[i] {
				t.Errorf("ConvertBase2ToBytes: expected %f, got %f", expected[i], actual)
			}
		}
	})
}
	

func TestInitMetrics(t *testing.T) {
	metrics := Metrics{}
	metrics.InitMetrics()
	if metrics.registry == nil {
		t.Errorf("InitMetrics returned nil registry")
	}
	if metrics.blockIOInGauge == nil {
		t.Errorf("InitMetrics returned nil blockIOInGauge")
	}
	if metrics.blockIOOutGauge == nil {
		t.Errorf("InitMetrics returned nil blockIOOutGauge")
	}

	if metrics.cpuPercGauge == nil {
		t.Errorf("InitMetrics returned nil cpuPercGauge")
	}
	if metrics.memPercGauge == nil {
		t.Errorf("InitMetrics returned nil memPercGauge")
	}
	if metrics.memUsageGauge == nil {
		t.Errorf("InitMetrics returned nil memUsageGauge")
	}
	if metrics.memAllowedGauge == nil {
		t.Errorf("InitMetrics returned nil memAllowedGauge")
	}

	if metrics.netIOInGauge == nil {
		t.Errorf("InitMetrics returned nil netIOGauge")
	}
	if metrics.netIOOutGauge == nil {
		t.Errorf("InitMetrics returned nil netIOGauge")
	}
}

func TestUpdateMetrics(t *testing.T) {
	input := `{"BlockIO":"13.5MB / 0B","CPUPerc":"1.50%","Container":"short_id","ID":"full_id","MemPerc":"0.02%","MemUsage":"9.809MiB / 32GiB","Name":"container_name","NetIO":"0B / 0B","PIDs":"5"}`
	stats, err := ParseDockerStatsLine(input)
	if err != nil {
		t.Errorf("Error parsing docker stats line: %s", err)
	}
	metrics := Metrics{}
	metrics.InitMetrics()
	metrics.UpdateMetrics(stats)
	var actual float64

	actual = testutil.ToFloat64(metrics.blockIOInGauge.WithLabelValues(stats.Name))
	if actual != 13500000 {
		t.Errorf("blockIOInGauge: expected 13500000, got %f", actual)
	}

	actual = testutil.ToFloat64(metrics.blockIOOutGauge.WithLabelValues(stats.Name))
	if actual != 0 {
		t.Errorf("blockIOOutGauge: expected 0, got %f", actual)
	}

	actual = testutil.ToFloat64(metrics.cpuPercGauge.WithLabelValues(stats.Name))
	if actual != 1.5 {
		t.Errorf("cpuPercGauge: expected 1.5, got %f", actual)
	}

	actual = testutil.ToFloat64(metrics.memPercGauge.WithLabelValues(stats.Name))
	if actual != 0.02 {
		t.Errorf("memPercGauge: expected 0.02, got %f", actual)
	}

	actual = testutil.ToFloat64(metrics.memUsageGauge.WithLabelValues(stats.Name))
	if actual != 10285482 {
		t.Errorf("memUsageGauge: expected 10285482, got %f", actual)
	}

	actual = testutil.ToFloat64(metrics.memAllowedGauge.WithLabelValues(stats.Name))
	if actual != 34359738368 {
		t.Errorf("memAllowedGauge: expected 34359738368, got %f", actual)
	}

	actual = testutil.ToFloat64(metrics.netIOInGauge.WithLabelValues(stats.Name))
	if actual != 0 {
		t.Errorf("netIOGauge: expected 0, got %f", actual)
	}

	actual = testutil.ToFloat64(metrics.netIOOutGauge.WithLabelValues(stats.Name))
	if actual != 0 {
		t.Errorf("netIOGauge: expected 0, got %f", actual)
	}

}
