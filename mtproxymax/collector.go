package mtproxymax

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PlushGuardian/obfuscation-server-exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

type MTPRoxyMax struct {
	descs []*prometheus.Desc
	mtx   sync.Mutex

	cfg    config.MTProxyMaxConfig
	logger *log.Logger
}

// NewCollector creates a collector that reads metrics from the given URL.
func NewCollector(cfg config.MTProxyMaxConfig, logger *log.Logger) *MTPRoxyMax {
	return &MTPRoxyMax{
		cfg:    cfg,
		logger: logger,
	}
}

// Describe sends the cached metric descriptions to the provided channel.
// On first call it fetches the remote endpoint to learn the available metrics.
func (c *MTPRoxyMax) Describe(ch chan<- *prometheus.Desc) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if c.descs == nil {
		families, err := c.fetchMetrics()
		if err != nil {
			c.logger.Printf("Describe: failed to fetch metrics from %s: %v", c.cfg.MetricsURL(), err)
			return
		}
		c.descs = generateDescs(families)
	}

	for _, d := range c.descs {
		ch <- d
	}
}

// Collect fetches the latest metrics from the remote endpoint and sends each
// Prometheus metric to the channel, converted from the scraped protocol buffer form.
func (c *MTPRoxyMax) Collect(ch chan<- prometheus.Metric) {
	families, err := c.fetchMetrics()
	if err != nil {
		c.logger.Printf("Collect: failed to fetch metrics from %s: %v", c.cfg.MetricsURL(), err)
		return
	}

	c.mtx.Lock()
	c.descs = generateDescs(families)
	c.mtx.Unlock()

	for name, mf := range families {
		if len(mf.Metric) == 0 {
			continue
		}

		labelKeys := make([]string, 0, len(mf.Metric[0].Label))
		for _, lp := range mf.Metric[0].Label {
			labelKeys = append(labelKeys, lp.GetName())
		}

		desc := prometheus.NewDesc(
			name,
			mf.GetHelp(),
			labelKeys,
			nil,
		)

		for _, m := range mf.Metric {
			labelVals := make([]string, len(labelKeys))
			for i, k := range labelKeys {
				labelVals[i] = getLabelValue(m, k)
			}

			var metric prometheus.Metric
			var err error

			switch mf.GetType() {
			case dto.MetricType_COUNTER:
				metric, err = prometheus.NewConstMetric(desc, prometheus.CounterValue, m.GetCounter().GetValue(), labelVals...)
			case dto.MetricType_GAUGE:
				metric, err = prometheus.NewConstMetric(desc, prometheus.GaugeValue, m.GetGauge().GetValue(), labelVals...)
			case dto.MetricType_UNTYPED:
				metric, err = prometheus.NewConstMetric(desc, prometheus.UntypedValue, m.GetUntyped().GetValue(), labelVals...)
			case dto.MetricType_HISTOGRAM:
				h := m.GetHistogram()
				buckets := make(map[float64]uint64)
				for _, b := range h.Bucket {
					buckets[b.GetUpperBound()] = b.GetCumulativeCount()
				}
				metric, err = prometheus.NewConstHistogram(desc, h.GetSampleCount(), h.GetSampleSum(), buckets, labelVals...)
			case dto.MetricType_SUMMARY:
				s := m.GetSummary()
				quantiles := make(map[float64]float64)
				for _, q := range s.Quantile {
					quantiles[q.GetQuantile()] = q.GetValue()
				}
				metric, err = prometheus.NewConstSummary(desc, s.GetSampleCount(), s.GetSampleSum(), quantiles, labelVals...)
			default:
				c.logger.Printf("Skipping unsupported metric type %v for metric %s", mf.GetType(), name)
				continue
			}

			if err != nil {
				c.logger.Printf("Error creating metric %s: %v", name, err)
				continue
			}

			ch <- metric
		}
	}
}

// fetchMetrics performs an HTTP GET against the configured URL and parses the
// response as Prometheus text format into metric families.
func (c *MTPRoxyMax) fetchMetrics() (map[string]*dto.MetricFamily, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(c.cfg.MetricsURL())
	if err != nil {
		return nil, fmt.Errorf("client http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	// Limit the amount we read to avoid huge responses.
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	parser := expfmt.TextParser{}
	families, err := parser.TextToMetricFamilies(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("parse metrics: %w", err)
	}
	return families, nil
}

// generateDescs creates a slice of *prometheus.Desc from the given metric families.
func generateDescs(families map[string]*dto.MetricFamily) []*prometheus.Desc {
	descs := make([]*prometheus.Desc, 0, len(families))
	for name, mf := range families {
		var labelKeys []string
		if len(mf.Metric) > 0 {
			for _, lp := range mf.Metric[0].Label {
				labelKeys = append(labelKeys, lp.GetName())
			}
		}
		descs = append(descs, prometheus.NewDesc(name, mf.GetHelp(), labelKeys, nil))
	}
	return descs
}

// getLabelValue returns the value for a given label key from a metric, or "" if missing.
func getLabelValue(m *dto.Metric, key string) string {
	for _, lp := range m.Label {
		if lp.GetName() == key {
			return lp.GetValue()
		}
	}
	return ""
}
