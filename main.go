package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PlushGuardian/obfuscation-server-exporter/config"
	"github.com/PlushGuardian/obfuscation-server-exporter/system"
	"github.com/PlushGuardian/obfuscation-server-exporter/threexui"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	// ---------- Create loggers -------------
	file, err := os.OpenFile("/var/log/obfuscation-server-exporter.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	obfsExporterLogger := log.New(file, "[obfs-exporter] ", log.LstdFlags)
	systemLogger := log.New(file, "[system] ", log.LstdFlags)
	threeXUILogger := log.New(file, "[3x-ui] ", log.LstdFlags)

	// ---------- CLI flags (pflag) ----------
	pflag.String("config-file", "", "Path to YAML configuration file")
	pflag.String("metrics-ip", "", "IP to listen on")
	pflag.String("metrics-port", "", "Port to listen on")
	pflag.Int("update-interval", 0, "Scrape interval in seconds")

	pflag.Int("xui-panel-port", 2053, "3X‑UI panel port")
	pflag.String("xui-panel-path", "", "3X‑UI panel path")
	pflag.String("xui-panel-username", "", "3X‑UI username")
	pflag.String("xui-panel-password", "", "3X‑UI password")
	pflag.Bool("xui-insecure-skip-verify", false, "Skip TLS verification")
	pflag.Int("xui-clients-bytes-rows", 0, "Top N rows for client bytes")
	pflag.Int("xui-timeout", 15, "Request timeout for the 3x-ui panel")
	pflag.Parse()

	// ---------- Bind pflags to Viper ----------
	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		obfsExporterLogger.Fatalf("failed to bind pflags: %v", err)
	}

	// ---------- Environment variables ----------
	// VIper automatically binds env vars: e.g. METRICS_IP -> metrics-ip
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// ---------- Config file (YAML) ----------
	if cfgFile := viper.GetString("config-file"); cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			obfsExporterLogger.Fatalf("failed to read config file: %v", err)
		}
	}

	// ---------- Unmarshal into typed config ----------
	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		obfsExporterLogger.Fatalf("failed to unmarshal config: %v", err)
	}

	// ---------- Validation (optional) ----------
	if cfg.ThreeXUI.PanelPath == "" {
		obfsExporterLogger.Fatal("threexui.panel_path is required (set via YAML, --panel-path, or PANEL_PATH)")
	}

	// ---------- Build collectors from config ----------
	reg := prometheus.NewRegistry()
	reg.MustRegister(threexui.NewCollector(config.ThreeXUIConfig{
		PanelPort:          cfg.ThreeXUI.PanelPort,
		PanelPath:          cfg.ThreeXUI.PanelPath,
		Username:           cfg.ThreeXUI.Username,
		Password:           cfg.ThreeXUI.Password,
		InsecureSkipVerify: cfg.ThreeXUI.InsecureSkipVerify,
		ClientsBytesRows:   cfg.ThreeXUI.ClientsBytesRows,
		Timeout:            cfg.ThreeXUI.Timeout,
	}, threeXUILogger))
	reg.MustRegister(system.NewCollector(config.SystemConfig{}, systemLogger))

	// ---------- HTTP handler ----------
	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		Timeout: cfg.OBFSExporter.ScrapeTimeout,
	})
	http.Handle("/metrics", handler)

	addr := cfg.OBFSExporter.Address + ":" + cfg.OBFSExporter.Port
	obfsExporterLogger.Printf("metrics server starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
