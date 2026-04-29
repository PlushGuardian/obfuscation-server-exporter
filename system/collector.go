package system

import (
	"log"
	"sync"

	"github.com/PlushGuardian/obfuscation-server-exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/cpu"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
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

	// Uptime
	uptimeDesc *prometheus.Desc

	// Network metrics aggregated to all interfaces
	netAggRecvBytesDesc *prometheus.Desc
	netAggSentBytesDesc *prometheus.Desc
	netAggRecvPktsDesc  *prometheus.Desc
	netAggSentPktsDesc  *prometheus.Desc
	netAggRecvErrsDesc  *prometheus.Desc
	netAggSentErrsDesc  *prometheus.Desc
	netAggRecvDropDesc  *prometheus.Desc
	netAggSentDropDesc  *prometheus.Desc

	// Network – per interface (device label)
	netIfRecvBytesDesc *prometheus.Desc
	netIfSentBytesDesc *prometheus.Desc
	netIfRecvPktsDesc  *prometheus.Desc
	netIfSentPktsDesc  *prometheus.Desc
	netIfRecvErrsDesc  *prometheus.Desc
	netIfSentErrsDesc  *prometheus.Desc
	netIfRecvDropDesc  *prometheus.Desc
	netIfSentDropDesc  *prometheus.Desc

	// Protocol counters (TCP/UDP stats)
	netProtoDesc *prometheus.Desc

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

		uptimeDesc: prometheus.NewDesc(
			"system_uptime_seconds",
			"System uptime in seconds",
			nil, nil,
		),

		netAggRecvBytesDesc: prometheus.NewDesc(
			"system_network_receive_bytes_total",
			"Cumulative received bytes (all interfaces)",
			nil, nil,
		),
		netAggSentBytesDesc: prometheus.NewDesc(
			"system_network_transmit_bytes_total",
			"Cumulative transmitted bytes (all interfaces)",
			nil, nil,
		),
		netAggRecvPktsDesc: prometheus.NewDesc(
			"system_network_receive_packets_total",
			"Cumulative received packets (all interfaces)",
			nil, nil,
		),
		netAggSentPktsDesc: prometheus.NewDesc(
			"system_network_transmit_packets_total",
			"Cumulative transmitted packets (all interfaces)",
			nil, nil,
		),
		netAggRecvErrsDesc: prometheus.NewDesc(
			"system_network_receive_errors_total",
			"Cumulative receive errors (all interfaces)",
			nil, nil,
		),
		netAggSentErrsDesc: prometheus.NewDesc(
			"system_network_transmit_errors_total",
			"Cumulative transmit errors (all interfaces)",
			nil, nil,
		),
		netAggRecvDropDesc: prometheus.NewDesc(
			"system_network_receive_dropped_total",
			"Cumulative receive drops (all interfaces)",
			nil, nil,
		),
		netAggSentDropDesc: prometheus.NewDesc(
			"system_network_transmit_dropped_total",
			"Cumulative transmit drops (all interfaces)",
			nil, nil,
		),

		netIfRecvBytesDesc: prometheus.NewDesc(
			"system_network_interface_receive_bytes_total",
			"Cumulative received bytes per interface",
			[]string{"device"}, nil,
		),
		netIfSentBytesDesc: prometheus.NewDesc(
			"system_network_interface_transmit_bytes_total",
			"Cumulative transmitted bytes per interface",
			[]string{"device"}, nil,
		),
		netIfRecvPktsDesc: prometheus.NewDesc(
			"system_network_interface_receive_packets_total",
			"Cumulative received packets per interface",
			[]string{"device"}, nil,
		),
		netIfSentPktsDesc: prometheus.NewDesc(
			"system_network_interface_transmit_packets_total",
			"Cumulative transmitted packets per interface",
			[]string{"device"}, nil,
		),
		netIfRecvErrsDesc: prometheus.NewDesc(
			"system_network_interface_receive_errors_total",
			"Cumulative receive errors per interface",
			[]string{"device"}, nil,
		),
		netIfSentErrsDesc: prometheus.NewDesc(
			"system_network_interface_transmit_errors_total",
			"Cumulative transmit errors per interface",
			[]string{"device"}, nil,
		),
		netIfRecvDropDesc: prometheus.NewDesc(
			"system_network_interface_receive_dropped_total",
			"Cumulative receive drops per interface",
			[]string{"device"}, nil,
		),
		netIfSentDropDesc: prometheus.NewDesc(
			"system_network_interface_transmit_dropped_total",
			"Cumulative transmit drops per interface",
			[]string{"device"}, nil,
		),

		netProtoDesc: prometheus.NewDesc(
			"system_network_protocol_total",
			"Cumulative protocol statistic (e.g. TCP segments, UDP datagrams, errors)",
			[]string{"protocol", "stat"}, nil,
		),

		logger: logger,
	}
}

// Describe sends all metric descriptors to the channel.
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

	// Uptime
	ch <- c.uptimeDesc

	// Network aggregated
	ch <- c.netAggRecvBytesDesc
	ch <- c.netAggSentBytesDesc
	ch <- c.netAggRecvPktsDesc
	ch <- c.netAggSentPktsDesc
	ch <- c.netAggRecvErrsDesc
	ch <- c.netAggSentErrsDesc
	ch <- c.netAggRecvDropDesc
	ch <- c.netAggSentDropDesc

	// Network per‑interface
	ch <- c.netIfRecvBytesDesc
	ch <- c.netIfSentBytesDesc
	ch <- c.netIfRecvPktsDesc
	ch <- c.netIfSentPktsDesc
	ch <- c.netIfRecvErrsDesc
	ch <- c.netIfSentErrsDesc
	ch <- c.netIfRecvDropDesc
	ch <- c.netIfSentDropDesc

	// Network protocol counters (TCP/UDP stats)
	ch <- c.netProtoDesc
}

// Collect gathers all metrics and sends them to the channel.
func (c *System) Collect(ch chan<- prometheus.Metric) {
	c.collectMemory(ch)
	c.collectSwap(ch)
	c.collectCPU(ch)
	c.collectLoad(ch)
	c.collectDisk(ch)
	c.collectUptime(ch)
	c.collectNetworkAggregated(ch)
	c.collectNetworkInterfaces(ch)
	c.collectProtocolStats(ch)
}

// === Helpers ===========================================

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

func (c *System) collectUptime(ch chan<- prometheus.Metric) {
	upt, err := host.Uptime()
	if err != nil {
		c.logger.Printf("error collecting uptime: %v", err)
		return
	}
	ch <- prometheus.MustNewConstMetric(c.uptimeDesc, prometheus.GaugeValue, float64(upt))
}

// collectNetworkAggregated exposes overall interface counters (sum of all NICs).
func (c *System) collectNetworkAggregated(ch chan<- prometheus.Metric) {
	agg, err := net.IOCounters(false) // false = aggregate of all interfaces
	if err != nil {
		c.logger.Printf("error collecting aggregated network stats: %v", err)
		return
	}
	if len(agg) != 1 {
		return
	}
	a := agg[0]
	ch <- prometheus.MustNewConstMetric(c.netAggRecvBytesDesc, prometheus.CounterValue, float64(a.BytesRecv))
	ch <- prometheus.MustNewConstMetric(c.netAggSentBytesDesc, prometheus.CounterValue, float64(a.BytesSent))
	ch <- prometheus.MustNewConstMetric(c.netAggRecvPktsDesc, prometheus.CounterValue, float64(a.PacketsRecv))
	ch <- prometheus.MustNewConstMetric(c.netAggSentPktsDesc, prometheus.CounterValue, float64(a.PacketsSent))
	ch <- prometheus.MustNewConstMetric(c.netAggRecvErrsDesc, prometheus.CounterValue, float64(a.Errin))
	ch <- prometheus.MustNewConstMetric(c.netAggSentErrsDesc, prometheus.CounterValue, float64(a.Errout))
	ch <- prometheus.MustNewConstMetric(c.netAggRecvDropDesc, prometheus.CounterValue, float64(a.Dropin))
	ch <- prometheus.MustNewConstMetric(c.netAggSentDropDesc, prometheus.CounterValue, float64(a.Dropout))
}

// collectNetworkInterfaces exposes per‑interface counters (e.g., eth0, tun0, wg0).
func (c *System) collectNetworkInterfaces(ch chan<- prometheus.Metric) {
	ifaces, err := net.IOCounters(true) // per NIC
	if err != nil {
		c.logger.Printf("error collecting per-interface network stats: %v", err)
		return
	}
	for _, iface := range ifaces {
		dev := []string{iface.Name}
		ch <- prometheus.MustNewConstMetric(c.netIfRecvBytesDesc, prometheus.CounterValue, float64(iface.BytesRecv), dev...)
		ch <- prometheus.MustNewConstMetric(c.netIfSentBytesDesc, prometheus.CounterValue, float64(iface.BytesSent), dev...)
		ch <- prometheus.MustNewConstMetric(c.netIfRecvPktsDesc, prometheus.CounterValue, float64(iface.PacketsRecv), dev...)
		ch <- prometheus.MustNewConstMetric(c.netIfSentPktsDesc, prometheus.CounterValue, float64(iface.PacketsSent), dev...)
		ch <- prometheus.MustNewConstMetric(c.netIfRecvErrsDesc, prometheus.CounterValue, float64(iface.Errin), dev...)
		ch <- prometheus.MustNewConstMetric(c.netIfSentErrsDesc, prometheus.CounterValue, float64(iface.Errout), dev...)
		ch <- prometheus.MustNewConstMetric(c.netIfRecvDropDesc, prometheus.CounterValue, float64(iface.Dropin), dev...)
		ch <- prometheus.MustNewConstMetric(c.netIfSentDropDesc, prometheus.CounterValue, float64(iface.Dropout), dev...)
	}
}

// collectProtocolStats emits TCP/UDP protocol counters (segments, opens, errors, etc.).
func (c *System) collectProtocolStats(ch chan<- prometheus.Metric) {
	protos, err := net.ProtoCounters([]string{"tcp", "udp"})
	if err != nil {
		c.logger.Printf("error collecting protocol stats: %v", err)
		return
	}
	for _, proto := range protos {
		for statName, val := range proto.Stats {
			ch <- prometheus.MustNewConstMetric(
				c.netProtoDesc,
				prometheus.CounterValue,
				float64(val),
				proto.Protocol,
				statName,
			)
		}
	}
}
