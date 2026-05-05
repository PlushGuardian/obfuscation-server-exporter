package main

import (
	"fmt"
	"log"
	"os"

	"github.com/PlushGuardian/obfuscation-server-exporter/config"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	// ---------- Create loggers ----------------------------
	file, err := os.OpenFile("./obfs-exporter.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) // TODO replace by var/log/obfs-exporter.log
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	obfsExporterLogger := log.New(file, "[obfs-exporter] ", log.LstdFlags)
	// systemLogger := log.New(file, "[system] ", log.LstdFlags)
	// threeXUILogger := log.New(file, "[3x-ui] ", log.LstdFlags)
	// mtproxyMaxLogger := log.New(file, "[mtproxymax] ", log.LstdFlags)

	// --- Create flags -------------------------------------

	v := viper.New()
	fs := pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	fs.SortFlags = false
	err = initializeFlags(v, fs)
	if err != nil {
		obfsExporterLogger.Fatal(err.Error())
	}

	fs.Parse(os.Args[1:])

}

func initializeFlags(v *viper.Viper, fs *pflag.FlagSet) error {
	fs.String("config-file", "", "Path to YAML configuration file")

	fs.Int("metrics-port", 9100, "Port for the obfuscation-server-exporter to listen on")
	fs.String("metrics-path", "/metrics", "Path the obfuscation-server-exporter listens on")
	fs.Int("scrape-timeout", 30, "Scrape timeout for the metrics port of the obfuscation-server-exporter")

	fs.Int("threexui-panel-port", 2053, "3X‑UI panel port")
	fs.String("threexui-panel-path", "", "3X‑UI panel path")
	fs.String("threexui-panel-username", "", "3X‑UI username")
	fs.String("threexui-panel-password", "", "3X‑UI password")
	fs.Bool("threexui-insecure-skip-verify", false, "Skip TLS verification")
	fs.Int("threexui-clients-bytes-rows", 0, "Top N rows for client bytes")
	fs.Int("threexui-timeout", 15, "Request timeout for the 3x-ui panel")

	fs.Int("mtproxymax-metrics-port", 9090, "3X‑UI panel port")
	fs.String("mtproxymax-metrics-path", "/metrics", "MTProxyMax metrics path")

	if err := v.BindPFlags(fs); err != nil {
		return fmt.Errorf("error when binding flags: %w", err)
	}

	rules := map[string]string{
		"^metrics-port$":   "obfs-exporter.metrics-port",
		"^metrics-path$":   "obfs-exporter.metrics-path",
		"^scrape-timeout$": "obfs-exporter.scrape-timeout",
		"^threexui-":       "threexui.",
		"^mtproxymax-":     "mtproxymax.",
	}

	if err := config.AliasFlags(v, fs, rules); err != nil {
		return fmt.Errorf("failed to setup flag aliases: %w", err)
	}
	return nil
}
