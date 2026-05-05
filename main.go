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

	// --- Create main function ------------------------------

	v := viper.New()
	fs := pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	fs.SortFlags = false
	err = initializeFlags(v, fs)
	if initializeFlags(v, fs) != nil {
		obfsExporterLogger.Fatal(err.Error())
	}

	fs.Parse(os.Args[1:])

}

func initializeFlags(v *viper.Viper, fs *pflag.FlagSet) error {
	fs.String("config-file", "", "Path to YAML configuration file")
	fs.Int("metrics-port", 9100, "Port for the metrics exporter")
	fs.Int("threexui-panel-port", 2053, "3X‑UI panel port")

	if err := v.BindPFlags(fs); err != nil {
		return fmt.Errorf("error when binding flags: %w", err)
	}

	rules := map[string]string{
		"^metrics-port$": "obfs-exporter.metrics-port",
		"^threexui-":     "threexui.",
	}

	if err := config.AliasFlags(v, fs, rules); err != nil {
		return fmt.Errorf("failed to setup flag aliases: %w", err)
	}
	return nil
}
