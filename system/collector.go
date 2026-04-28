package system

import (
	"log"

	"github.com/PlushGuardian/obfuscation-server-exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	// "github.com/shirou/gopsutil/v3/cpu"
	// "github.com/shirou/gopsutil/v3/disk"
	// "github.com/shirou/gopsutil/v3/host"
	// "github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	// "github.com/shirou/gopsutil/v3/net"
)

// System collects a single system metric.
type System struct {
	// Memory & swap
	totalMemoryDesc *prometheus.Desc
	usedMemoryDesc  *prometheus.Desc
	swapTotalDesc   *prometheus.Desc
	swapUsedDesc    *prometheus.Desc

	logger *log.Logger
}

// NewCollector creates a new system metrics collector.
func NewCollector(cfg config.SystemConfig, logger *log.Logger) *System {
	return &System{
		totalMemoryDesc: prometheus.NewDesc(
			"system_memory_total_bytes",
			"Total installed physical memory in bytes",
			nil, nil,
		),
		usedMemoryDesc: prometheus.NewDesc(
			"system_memory_used_bytes",
			"Currently used physical memory in bytes",
			nil, nil,
		),
		swapTotalDesc: prometheus.NewDesc(
			"system_swap_total_bytes",
			"Total swap space in bytes",
			nil, nil,
		),
		swapUsedDesc: prometheus.NewDesc(
			"system_swap_used_bytes",
			"Used swap space in bytes",
			nil, nil,
		),
		logger: logger,
	}
}

func (c *System) Describe(ch chan<- *prometheus.Desc) {
	// Memory & swap
	ch <- c.totalMemoryDesc
	ch <- c.usedMemoryDesc
	ch <- c.swapTotalDesc
	ch <- c.swapUsedDesc
}

func (c *System) Collect(ch chan<- prometheus.Metric) {
	c.collectMemory(ch)
	c.collectSwap(ch)
}

func (c *System) collectMemory(ch chan<- prometheus.Metric) {
	v, err := mem.VirtualMemory()
	if err != nil {
		c.logger.Printf("error collecting memory: %v", err)
		return
	}
	ch <- prometheus.MustNewConstMetric(c.totalMemoryDesc, prometheus.GaugeValue, float64(v.Total))
	ch <- prometheus.MustNewConstMetric(c.usedMemoryDesc, prometheus.GaugeValue, float64(v.Used))
}

func (c *System) collectSwap(ch chan<- prometheus.Metric) {
	s, err := mem.SwapMemory()
	if err != nil {
		c.logger.Printf("error collecting swap: %v", err)
		return
	}
	ch <- prometheus.MustNewConstMetric(c.swapTotalDesc, prometheus.GaugeValue, float64(s.Total))
	ch <- prometheus.MustNewConstMetric(c.swapUsedDesc, prometheus.GaugeValue, float64(s.Used))
}
