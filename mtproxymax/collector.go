package mtproxymax

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PlushGuardian/obfuscation-server-exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

type MTPRoxyMax struct {
	url string
	descs []*prometheus.Desc
	mtx   sync.Mutex

	cfg    *config.MTProxyMaxConfig
	logger *log.Logger
}

// newCollector creates a collector that reads metrics from the given URL.
func newCollector(cfg *config.MTProxyMaxConfig, logger *log.Logger) *MTPRoxyMax {
	return &MTPRoxyMax{
		url: "http://localhost<port>/path"
		cfg: cfg,
		logger: logger,
	}
}

// Describe sends the cached metric descriptions to the provided channel.
// On first call it fetches the remote endpoint to learn the available metrics.
func (rc *MTPRoxyMax) Describe(ch chan<- *prometheus.Desc) {
	rc.mtx.Lock()
	defer rc.mtx.Unlock()

	if rc.descs == nil {
		families, err := rc.fetchMetrics()
		if err != nil {
			log.Printf("Describe: failed to fetch metrics from %s: %v", rc.url, err)
			return
		}
		rc.descs = generateDescs(families)
	}

	for _, d := range rc.descs {
		ch <- d
	}
}

// Collect fetches the latest metrics from the remote endpoint and sends each
// Prometheus metric to the channel, converted from the scraped protocol buffer form.
func (rc *MTPRoxyMax) Collect(ch chan<- prometheus.Metric) {
	families, err := rc.fetchMetrics()
	if err != nil {
		log.Printf("Collect: failed to fetch metrics from %s: %v", rc.url, err)
		return
	}

	// Update cached descriptions for consistency (they may change over time).
	rc.mtx.Lock()
	rc.descs = generateDescs(families)
	rc.mtx.Unlock()

	for name, mf := range families {
		if len(mf.Metric) == 0 {
			continue
		}

		// Label keys are taken from the first metric (must be the same for all in the family).
		labelKeys := make([]string, 0, len(mf.Metric[0].Label))
		for _, lp := range mf.Metric[0].Label {
			labelKeys = append(labelKeys, lp.GetName())
		}

		// Build a description that matches the remote metric.
		desc := prometheus.NewDesc(
			name,
			mf.GetHelp(),
			labelKeys,
			nil, // no const labels
		)

		// Convert each metric in the family.
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
				log.Printf("Skipping unsupported metric type %v for metric %s", mf.GetType(), name)
				continue
			}

			if err != nil {
				log.Printf("Error creating metric %s: %v", name, err)
				continue
			}

			ch <- metric
		}
	}
}

// fetchMetrics performs an HTTP GET against the configured URL and parses the
// response as Prometheus text format into metric families.
func (rc *MTPRoxyMax) fetchMetrics() (map[string]*dto.MetricFamily, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(rc.url)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
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

func main() {
	listenAddr := flag.String("web.listen-address", ":9090", "Address to listen on for HTTP requests.")
	scrapeURL := flag.String("scrape.url", "http://localhost:9100/metrics", "URL of the Prometheus metrics endpoint to scrape.")
	flag.Parse()

	// Create a fresh Prometheus registry and add the built-in collectors.
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
		prometheus.NewGoCollector(),
	)

	// Register our remote scraping collector.
	reg.MustRegister(newRemoteCollector(*scrapeURL))

	// Expose the merged metrics.
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	log.Printf("Starting server on %s, scraping %s", *listenAddr, *scrapeURL)
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
