package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PlushGuardian/obfuscation-server-exporter/config"
	"github.com/PlushGuardian/obfuscation-server-exporter/mtproxymax"
	"github.com/PlushGuardian/obfuscation-server-exporter/system"
	"github.com/PlushGuardian/obfuscation-server-exporter/threexui"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	logFile := "/var/log/obfs-exporter.log"
	cfgFile := "/etc/obfs-exporter/config.yaml"

	// ---------- Create loggers ----------------------------
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: cannot open log file %q: %v. Logging to stderr instead.\n", logFile, err)
		file = os.Stderr
	}
	if file != nil {
		defer func() { _ = file.Close() }()
	}
	obfsExporterLogger := log.New(file, "[obfs-exporter] ", log.LstdFlags)
	systemLogger := log.New(file, "[system] ", log.LstdFlags)
	threeXUILogger := log.New(file, "[3x-ui] ", log.LstdFlags)
	mtproxyMaxLogger := log.New(file, "[mtproxymax] ", log.LstdFlags)

	// --- Create flags -------------------------------------
	v := viper.New()
	fs := pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	fs.SortFlags = false

	err = initializeFlags(v, fs)
	if err != nil {
		obfsExporterLogger.Fatal(err.Error())
	}

	// ---------- Config file (YAML) ----------
	if val := v.GetString("config-file"); val != "" {
		cfgFile = val
	}
	v.SetConfigFile(cfgFile)
	if err := v.ReadInConfig(); err != nil {
		obfsExporterLogger.Printf("WARNING: failed to read config file %s: %v. Using default configuration.", cfgFile, err)
	}

	// ---------- Environment variables ----------
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	// ---------- Unmarshal into typed config ----------
	var cfg config.Config
	if err := v.Unmarshal(&cfg); err != nil {
		obfsExporterLogger.Fatalf("failed to unmarshal config: %v", err)
	}
	config.SetupConfigWatch(v, &cfg, obfsExporterLogger)

	// ---------- Build collectors from config ----------
	reg := prometheus.NewRegistry()
	reg.MustRegister(system.NewCollector(config.SystemConfig{}, systemLogger))
	reg.MustRegister(threexui.NewCollector(config.ThreeXUIConfig{
		PanelPort:          cfg.ThreeXUI.PanelPort,
		PanelPath:          cfg.ThreeXUI.PanelPath,
		Username:           cfg.ThreeXUI.Username,
		Password:           cfg.ThreeXUI.Password,
		InsecureSkipVerify: cfg.ThreeXUI.InsecureSkipVerify,
		ClientsBytesRows:   cfg.ThreeXUI.ClientsBytesRows,
		Timeout:            cfg.ThreeXUI.Timeout,
	}, threeXUILogger))
	reg.MustRegister(mtproxymax.NewCollector(config.MTProxyMaxConfig{
		MetricsPort: cfg.MTProxyMax.MetricsPort,
		MetricsPath: cfg.MTProxyMax.MetricsPath,
	}, mtproxyMaxLogger))

	// ---------- HTTP handler ----------
	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		Timeout: cfg.OBFSExporter.ScrapeTimeout,
	})
	http.Handle(cfg.OBFSExporter.MetricsPath, handler)

	addr, _ := cfg.OBFSExporter.Addr()

	obfsExporterLogger.Printf("metrics server starting on %s", addr)
	obfsExporterLogger.Fatal(http.ListenAndServe(addr, nil))
}

func initializeFlags(v *viper.Viper, fs *pflag.FlagSet) error {
	fs.String("config-file", "", "Path to YAML configuration file")

	fs.Int("metrics-port", 9100, "Port for the obfuscation-server-exporter to listen on")
	fs.String("metrics-path", "/metrics", "Path the obfuscation-server-exporter listens on")
	fs.Int("scrape-timeout", 30, "Scrape timeout for the metrics port of the obfuscation-server-exporter")

	fs.Int("threexui-panel-port", 2053, "3X‑UI panel port")
	fs.String("threexui-panel-path", "", "3X‑UI panel path")
	fs.String("threexui-username", "", "3X‑UI username")
	fs.String("threexui-password", "", "3X‑UI password")
	fs.Bool("threexui-insecure-skip-verify", false, "Skip TLS verification")
	fs.Int("threexui-clients-bytes-rows", 0, "Top N rows for client bytes")
	fs.Int("threexui-timeout", 15, "Request timeout for the 3x-ui panel")

	fs.Int("mtproxymax-metrics-port", 9090, "3X‑UI panel port")
	fs.String("mtproxymax-metrics-path", "/metrics", "MTProxyMax metrics path")

	var bindErr error

	fs.VisitAll(func(f *pflag.Flag) {
		configKey := f.Name

		switch {
		case f.Name == "metrics-port" || f.Name == "metrics-path" || f.Name == "scrape-timeout":
			configKey = "obfs-exporter." + f.Name
		case strings.HasPrefix(f.Name, "threexui-"):
			configKey = strings.Replace(f.Name, "threexui-", "threexui.", 1)
		case strings.HasPrefix(f.Name, "mtproxymax-"):
			configKey = strings.Replace(f.Name, "mtproxymax-", "mtproxymax.", 1)
		}

		if err := v.BindPFlag(configKey, f); err != nil {
			bindErr = fmt.Errorf("failed to bind flag %s to key %s: %w", f.Name, configKey, err)
		}
	})

	if bindErr != nil {
		return bindErr
	}

	return nil
}
