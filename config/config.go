package config

import (
	"fmt"
	"net"
	"net/url"
	"time"
)

type Config struct {
	OBFSExporter OBFSExporterConfig `mapstructure:"obfs-exporter"`
	//nolint:unused
	System     SystemConfig     `mapstructure:"system"`
	ThreeXUI   ThreeXUIConfig   `mapstructure:"threexui"`
	MTProxyMax MTProxyMaxConfig `mapstructure:"mtproxymax"`
}

type OBFSExporterConfig struct {
	MetricsPort   int           `mapstructure:"metrics-port"`
	MetricsPath   string        `mapstructure:"metrics-path"`
	ScrapeTimeout time.Duration `mapstructure:"scrape-timeout"`
}

func (c *OBFSExporterConfig) Addr() (string, error) {
	host := "localhost" // no other hosts used by design

	return fmt.Sprint(net.JoinHostPort(host, fmt.Sprint(c.MetricsPort))), nil
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

type MTProxyMaxConfig struct {
	MetricsPort int    `mapstructure:"metrics-port"`
	MetricsPath string `mapstructure:"metrics-path"`
}

func (c *MTProxyMaxConfig) MetricsURL() string {
	host := "localhost" // no other hosts used by design
	u := &url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, fmt.Sprint(c.MetricsPort)),
	}
	return u.JoinPath(c.MetricsPath).String()
}
