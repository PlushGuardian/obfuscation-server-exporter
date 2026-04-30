package config

import (
	"fmt"
	"net"
	"net/url"
	"time"
)

type Config struct {
	OBFSExporter OBFSExporterConfig `mapstructure:"obfsexporter"`
	//nolint:unused
	System   SystemConfig   `mapstructure:"system"`
	ThreeXUI ThreeXUIConfig `mapstructure:"threexui"`
}

type OBFSExporterConfig struct {
	Address       string `mapstructure:"address"`
	Port          string `mapstructure:"port"`
	ScrapeTimeout int    `mapstructure:"scrape-timeout"` // seconds
}

type SystemConfig struct {
}

type ThreeXUIConfig struct {
	PanelPort          int           `mapstructure:"panel-port"`
	PanelPath          string        `mapstructure:"panel-path"`
	Username           string        `mapstructure:"username"`
	Password           string        `mapstructure:"password"`
	InsecureSkipVerify bool          `mapstructure:"insecure-skip-verify"`
	ClientsBytesRows   int           `mapstructure:"clients-bytes-rows"`
	Timeout            time.Duration `mapstructure:"timeout"`
}

func (c *ThreeXUIConfig) PanelURL() (string, error) {
	host := "localhost" // no other hosts used by design
	u := &url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, fmt.Sprint(c.PanelPort)),
	}

	return u.JoinPath(c.PanelPath).String(), nil
}
