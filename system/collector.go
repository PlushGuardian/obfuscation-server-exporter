package system

import (
	"log"

	"github.com/PlushGuardian/obfuscation-server-exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/mem"
)

// System collects a single system metric.
type System struct {
	totalMemoryDesc *prometheus.Desc
	usedMemoryDesc  *prometheus.Desc
	logger          *log.Logger
}

// NewCollector creates a new system metrics collector.
func NewCollector(cfg config.SystemConfig, logger *log.Logger) *System {
	return &System{
		totalMemoryDesc: prometheus.NewDesc(
			"system_memory_total_bytes",
			"Total installed physical memory in bytes",
			nil, nil,
		),
		usedMemoryDesc: prometheus.NewDesc( // NEW: descriptor for used memory
			"system_memory_used_bytes",
			"Currently used physical memory in bytes",
			nil, nil,
		),
		logger: logger,
	}
}

func (c *System) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.totalMemoryDesc
	ch <- c.usedMemoryDesc
}

func (c *System) Collect(ch chan<- prometheus.Metric) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return
	}
	ch <- prometheus.MustNewConstMetric(
		c.totalMemoryDesc,
		prometheus.GaugeValue,
		float64(v.Total),
	)
	ch <- prometheus.MustNewConstMetric(
		c.usedMemoryDesc,
		prometheus.GaugeValue,
		float64(v.Used),
	)
}
