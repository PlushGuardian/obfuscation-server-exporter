package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PlushGuardian/obfuscation-server-exporter/config"
	"github.com/PlushGuardian/obfuscation-server-exporter/threexui"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	file, err := os.OpenFile("/var/log/obfuscation-server-exporter.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	// ---------- 1. CLI flags (pflag) ----------
	pflag.String("config-file", "", "Path to YAML configuration file")
	pflag.String("metrics-ip", "", "IP to listen on")
	pflag.String("metrics-port", "", "Port to listen on")
	pflag.Int("update-interval", 0, "Scrape interval in seconds")
	pflag.Int("clients-bytes-rows", 0, "Top N rows for client bytes")
	pflag.String("panel-port", "", "3X‑UI panel port")
	pflag.String("panel-path", "", "3X‑UI panel path")
	pflag.String("panel-base-url", "", "3X‑UI base URL")
	pflag.String("panel-username", "", "3X‑UI username")
	pflag.String("panel-password", "", "3X‑UI password")
	pflag.Bool("insecure-skip-verify", false, "Skip TLS verification")
	pflag.Parse()

	// ---------- 2. Bind pflags to Viper ----------
	viper.BindPFlags(pflag.CommandLine)

	// ---------- 3. Environment variables ----------
	// VIper automatically binds env vars: e.g. METRICS_IP -> metrics-ip
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// ---------- 4. Config file (YAML) ----------
	if cfgFile := viper.GetString("config-file"); cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("failed to read config file: %v", err)
		}
	}

	// ---------- 5. Set defaults ----------
	viper.SetDefault("osmexporter.address", "localhost")
	viper.SetDefault("osmexporter.port", "9100")
	viper.SetDefault("osmexporter.scrape_timeout", 30)
	viper.SetDefault("threexui.timeout", 15)
	viper.SetDefault("threexui.clients_bytes_rows", 0)

	// ---------- 6. Unmarshal into typed config ----------
	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}

	// ---------- 7. Validation (optional) ----------
	if cfg.ThreeXUI.PanelPath == "" {
		log.Fatal("threexui.panel_path is required (set via YAML, --panel-path, or PANEL_PATH)")
	}

	// ---------- 8. Create loggers (tagged) ----------
	threeXUILogger := log.New(file, "[3x-ui] ", log.LstdFlags)
	obfsExporterLogger := log.New(file, "[obfs-exporter] ", log.LstdFlags)

	// ---------- 9. Build collectors from config ----------
	panelURL := "http://localhost:" + cfg.ThreeXUI.PanelPort + "/" + strings.Trim(cfg.ThreeXUI.PanelPath, "/")
	reg := prometheus.NewRegistry()
	reg.MustRegister(threexui.NewCollector(threexui.Config{
		BaseURL:            panelURL,
		Username:           cfg.ThreeXUI.Username,
		Password:           cfg.ThreeXUI.Password,
		InsecureSkipVerify: cfg.ThreeXUI.InsecureSkipVerify,
		ClientsBytesRows:   cfg.ThreeXUI.ClientsBytesRows,
		Timeout:            time.Duration(cfg.ThreeXUI.Timeout) * time.Second,
	}, threeXUILogger))

	// ---------- 10. HTTP handler ----------
	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		Timeout: time.Duration(cfg.OBFSExporter.ScrapeTimeout) * time.Second,
	})
	http.Handle("/metrics", handler)

	addr := cfg.OBFSExporter.Address + ":" + cfg.OBFSExporter.Port
	obfsExporterLogger.Printf("metrics server starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
