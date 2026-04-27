package config

import (
	"fmt"
	"strings"
)

type Config struct {
	OBFSExporter OBFSExporterConfig `mapstructure:"obfsexporter"`
	system       SystemConfig       `mapstructure:"system"`
	ThreeXUI     ThreeXUIConfig     `mapstructure:"threexui"`
}

type OBFSExporterConfig struct {
	Address       string `mapstructure:"address"`
	Port          string `mapstructure:"port"`
	ScrapeTimeout int    `mapstructure:"scrape-timeout"` // seconds
}

type SystemConfig struct {
}

type ThreeXUIConfig struct {
	PanelPort          string `mapstructure:"panel-port"`
	PanelPath          string `mapstructure:"panel-path"`
	Username           string `mapstructure:"username"`
	Password           string `mapstructure:"password"`
	InsecureSkipVerify bool   `mapstructure:"insecure_skip-verify"`
	ClientsBytesRows   int    `mapstructure:"clients-bytes-rows"`
	Timeout            int    `mapstructure:"timeout"` // seconds
}

func (c ThreeXUIConfig) PanelURL() string {
	return fmt.Sprintf("http://%s:%s/%s", "localhost", c.PanelPort, strings.Trim(c.PanelPath, "/"))
}
