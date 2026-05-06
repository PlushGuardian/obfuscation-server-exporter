package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"

	// "github.com/PlushGuardian/obfuscation-server-exporter/config"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	logFile := "./obfs-exporter.log" // TODO replace by var/log/obfs-exporter.log
	cfgFile := "./config.yaml"       // TODO find better naming
	// ---------- Create loggers ----------------------------
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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

	// ---------- Config file (YAML) ----------

	if val := v.GetString("config-file"); val != "" {
		cfgFile = val
	}
	v.SetConfigFile(cfgFile)
	if err := v.ReadInConfig(); err != nil {
		obfsExporterLogger.Fatalf("failed to read config file: %v", err)
	} else {
		setupConfigWatch(v, obfsExporterLogger)
	}
	printConfig(v)
	// ---------- Environment variables ----------
	viper.AutomaticEnv()

}

func printConfig(v *viper.Viper) {
	settings := v.AllSettings()
	out, err := yaml.Marshal(settings)
	if err != nil {
		log.Fatalf("failed to marshal config: %v", err)
	}

	fmt.Println("--- Current Configuration ---")
	fmt.Println(string(out))
	fmt.Println("------------------------------")
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

	// Parse flags only if they haven't been parsed yet to avoid panics
	if !fs.Parsed() {
		if err := fs.Parse(os.Args[1:]); err != nil {
			return fmt.Errorf("failed to parse flags: %w", err)
		}
	}

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

func setupConfigWatch(v *viper.Viper, logger *log.Logger) {
	logger.Printf("Watching config file %s\n", v.ConfigFileUsed())
	v.OnConfigChange(func(e fsnotify.Event) {
		// This fires whenever the file is saved
		logger.Printf("Config file changed: %s\n", e.Name)

		// If you are unmarshaling into a struct, you should re-unmarshal here:
		// v.Unmarshal(&myConfigStruct)
	})

	v.WatchConfig()
}
