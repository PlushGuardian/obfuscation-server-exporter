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
	Address       string        `mapstructure:"address"`
	Port          int32         `mapstructure:"port"`
	ScrapeTimeout time.Duration `mapstructure:"scrape-timeout"`
}

type SystemConfig struct {
}

type ThreeXUIConfig struct {
	PanelPort          int32         `mapstructure:"xui-panel-port"`
	PanelPath          string        `mapstructure:"xui-panel-path"`
	Username           string        `mapstructure:"xui-username"`
	Password           string        `mapstructure:"xui-password"`
	InsecureSkipVerify bool          `mapstructure:"xui-insecure-skip-verify"`
	ClientsBytesRows   int32         `mapstructure:"xui-clients-bytes-rows"`
	Timeout            time.Duration `mapstructure:"xui-timeout"`
}

func (c *ThreeXUIConfig) PanelURL() (string, error) {
	host := "localhost" // no other hosts used by design
	u := &url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, fmt.Sprint(c.PanelPort)),
	}
	return u.JoinPath(c.PanelPath).String(), nil
}
