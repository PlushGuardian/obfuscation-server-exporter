package system

import (
	"log"
	"sync"

	"github.com/PlushGuardian/obfuscation-server-exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/cpu"

	"github.com/shirou/gopsutil/v3/disk"
	// "github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
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

	// CPU & load average
	cpuUsageDesc *prometheus.Desc
	load1Desc    *prometheus.Desc
	load5Desc    *prometheus.Desc
	load15Desc   *prometheus.Desc

	// Disk
	diskTotalDesc *prometheus.Desc
	diskUsedDesc  *prometheus.Desc

	// CPU state for delta calculation
	lastCPUTimes *cpu.TimesStat
	mu           sync.Mutex

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

		cpuUsageDesc: prometheus.NewDesc(
			"system_cpu_usage_percent",
			"Current overall CPU usage as a percentage (0-100)",
			nil, nil,
		),
		load1Desc: prometheus.NewDesc(
			"system_load1",
			"1-minute load average",
			nil, nil,
		),
		load5Desc: prometheus.NewDesc(
			"system_load5",
			"5-minute load average",
			nil, nil,
		),
		load15Desc: prometheus.NewDesc(
			"system_load15",
			"15-minute load average",
			nil, nil,
		),

		diskTotalDesc: prometheus.NewDesc(
			"system_disk_total_bytes",
			"Total disk space in bytes per partition",
			[]string{"device", "mountpoint", "fstype"}, nil,
		),
		diskUsedDesc: prometheus.NewDesc(
			"system_disk_used_bytes",
			"Used disk space in bytes per partition",
			[]string{"device", "mountpoint", "fstype"}, nil,
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

	// CPU & load average
	ch <- c.cpuUsageDesc
	ch <- c.load1Desc
	ch <- c.load5Desc
	ch <- c.load15Desc

	// Disk
	ch <- c.diskTotalDesc
	ch <- c.diskUsedDesc

}

func (c *System) Collect(ch chan<- prometheus.Metric) {
	c.collectMemory(ch)
	c.collectSwap(ch)
	c.collectCPU(ch)
	c.collectLoad(ch)
	c.collectDisk(ch)
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

func (c *System) collectCPU(ch chan<- prometheus.Metric) {
	times, err := cpu.Times(false) // overall, no per-CPU breakdown
	if err != nil {
		c.logger.Printf("error collecting CPU times: %v", err)
		return
	}
	if len(times) == 0 {
		return
	}
	curr := times[0]

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lastCPUTimes != nil {
		totalDelta := (curr.User + curr.System + curr.Idle + curr.Nice + curr.Iowait +
			curr.Irq + curr.Softirq + curr.Steal + curr.Guest + curr.GuestNice) -
			(c.lastCPUTimes.User + c.lastCPUTimes.System + c.lastCPUTimes.Idle + c.lastCPUTimes.Nice +
				c.lastCPUTimes.Iowait + c.lastCPUTimes.Irq + c.lastCPUTimes.Softirq + c.lastCPUTimes.Steal +
				c.lastCPUTimes.Guest + c.lastCPUTimes.GuestNice)
		idleDelta := (curr.Idle + curr.Iowait) - (c.lastCPUTimes.Idle + c.lastCPUTimes.Iowait)
		if totalDelta > 0 {
			usage := (totalDelta - idleDelta) / totalDelta * 100.0
			ch <- prometheus.MustNewConstMetric(c.cpuUsageDesc, prometheus.GaugeValue, usage)
		}
	}
	c.lastCPUTimes = &curr
}

func (c *System) collectLoad(ch chan<- prometheus.Metric) {
	lavg, err := load.Avg()
	if err != nil {
		c.logger.Printf("error collecting load average: %v", err)
		return
	}
	ch <- prometheus.MustNewConstMetric(c.load1Desc, prometheus.GaugeValue, lavg.Load1)
	ch <- prometheus.MustNewConstMetric(c.load5Desc, prometheus.GaugeValue, lavg.Load5)
	ch <- prometheus.MustNewConstMetric(c.load15Desc, prometheus.GaugeValue, lavg.Load15)
}

func (c *System) collectDisk(ch chan<- prometheus.Metric) {
	parts, err := disk.Partitions(false)
	if err != nil {
		c.logger.Printf("error listing disk partitions: %v", err)
		return
	}
	for _, p := range parts {
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			c.logger.Printf("error getting usage for %s: %v", p.Mountpoint, err)
			continue
		}
		lbls := []string{p.Device, p.Mountpoint, p.Fstype}
		ch <- prometheus.MustNewConstMetric(c.diskTotalDesc, prometheus.GaugeValue, float64(usage.Total), lbls...)
		ch <- prometheus.MustNewConstMetric(c.diskUsedDesc, prometheus.GaugeValue, float64(usage.Used), lbls...)
	}
}
