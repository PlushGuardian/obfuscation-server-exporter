package threexui

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/PlushGuardian/obfuscation-server-exporter/config"
	"github.com/prometheus/client_golang/prometheus"
)

type ThreeXUI struct {
	client *threeXUIClient
	logger *log.Logger

	onlineUsersDesc  *prometheus.Desc
	inboundUpDesc    *prometheus.Desc
	inboundDownDesc  *prometheus.Desc
	clientUpDesc     *prometheus.Desc
	clientDownDesc   *prometheus.Desc
	xrayVersionDesc  *prometheus.Desc
	panelThreadsDesc *prometheus.Desc
	panelMemoryDesc  *prometheus.Desc
	panelUptimeDesc  *prometheus.Desc
}

func NewCollector(cfg config.ThreeXUIConfig, logger *log.Logger) *ThreeXUI {
	cli := newClient(cfg, logger)
	return &ThreeXUI{
		client: cli,
		logger: logger,
		onlineUsersDesc: prometheus.NewDesc(
			"x_ui_total_online_users", "Total number of online users", nil, nil,
		),
		inboundUpDesc: prometheus.NewDesc(
			"x_ui_inbound_up_bytes", "Total uploaded bytes per inbound",
			[]string{"id", "remark"}, nil,
		),
		inboundDownDesc: prometheus.NewDesc(
			"x_ui_inbound_down_bytes", "Total downloaded bytes per inbound",
			[]string{"id", "remark"}, nil,
		),
		clientUpDesc: prometheus.NewDesc(
			"x_ui_client_up_bytes", "Total uploaded bytes per client",
			[]string{"id", "email"}, nil,
		),
		clientDownDesc: prometheus.NewDesc(
			"x_ui_client_down_bytes", "Total downloaded bytes per client",
			[]string{"id", "email"}, nil,
		),
		xrayVersionDesc: prometheus.NewDesc(
			"x_ui_xray_version", "XRay version used by 3X-UI",
			[]string{"version"}, nil,
		),
		panelThreadsDesc: prometheus.NewDesc(
			"x_ui_panel_threads", "3X-UI panel threads", nil, nil,
		),
		panelMemoryDesc: prometheus.NewDesc(
			"x_ui_panel_memory", "3X-UI panel memory usage", nil, nil,
		),
		panelUptimeDesc: prometheus.NewDesc(
			"x_ui_panel_uptime", "3X-UI panel uptime", nil, nil,
		),
	}
}

func (c *ThreeXUI) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.onlineUsersDesc
	ch <- c.inboundUpDesc
	ch <- c.inboundDownDesc
	ch <- c.clientUpDesc
	ch <- c.clientDownDesc
	ch <- c.xrayVersionDesc
	ch <- c.panelThreadsDesc
	ch <- c.panelMemoryDesc
	ch <- c.panelUptimeDesc
}

func (c *ThreeXUI) Collect(ch chan<- prometheus.Metric) {
	cookie, err := c.client.getAuthToken()
	if err != nil {
		c.logger.Printf("3x-ui: authentication failed, skipping all metrics: %v", err)
		return
	}

	// Online users
	if body, err := c.client.do(http.MethodPost, "/panel/api/inbounds/onlines", cookie); err == nil && len(body) > 0 {
		var resp apiResponse
		if json.Unmarshal(body, &resp) == nil {
			var arr []json.RawMessage
			if json.Unmarshal(resp.Obj, &arr) == nil {
				ch <- prometheus.MustNewConstMetric(c.onlineUsersDesc, prometheus.GaugeValue, float64(len(arr)))
			} else {
				c.logger.Println("3x-ui: failed to parse online users list")
			}
		} else {
			c.logger.Println("3x-ui: failed to unmarshal online users response")
		}
	} else if err != nil {
		c.logger.Printf("3x-ui: online users request error: %v", err)
	}

	// Server status
	if body, err := c.client.do(http.MethodGet, "/panel/api/server/status", cookie); err == nil && len(body) > 0 {
		var resp serverStatusResponse
		if json.Unmarshal(body, &resp) == nil {
			ver := resp.Obj.Xray.Version
			num, _ := strconv.ParseFloat(strings.ReplaceAll(ver, ".", ""), 64)
			ch <- prometheus.MustNewConstMetric(c.xrayVersionDesc, prometheus.GaugeValue, num, ver)
			ch <- prometheus.MustNewConstMetric(c.panelThreadsDesc, prometheus.GaugeValue, float64(resp.Obj.AppStats.Threads))
			ch <- prometheus.MustNewConstMetric(c.panelMemoryDesc, prometheus.GaugeValue, float64(resp.Obj.AppStats.Mem))
			ch <- prometheus.MustNewConstMetric(c.panelUptimeDesc, prometheus.GaugeValue, float64(resp.Obj.AppStats.Uptime))
		} else {
			c.logger.Println("3x-ui: failed to unmarshal server status")
		}
	} else if err != nil {
		c.logger.Printf("3x-ui: server status request error: %v", err)
	}

	// Inbounds list
	if body, err := c.client.do(http.MethodGet, "/panel/api/inbounds/list", cookie); err == nil {
		var resp getInboundsResponse
		if json.Unmarshal(body, &resp) == nil {
			for _, inbound := range resp.Obj {
				iid := strconv.Itoa(inbound.ID)
				ch <- prometheus.MustNewConstMetric(c.inboundUpDesc, prometheus.GaugeValue, float64(inbound.Up), iid, inbound.Remark)
				ch <- prometheus.MustNewConstMetric(c.inboundDownDesc, prometheus.GaugeValue, float64(inbound.Down), iid, inbound.Remark)

				clients := inbound.ClientStats
				n := c.client.config.ClientsBytesRows
				if n == 0 {
					for _, cl := range clients {
						ch <- prometheus.MustNewConstMetric(c.clientUpDesc, prometheus.GaugeValue, float64(cl.Up),
							strconv.Itoa(cl.ID), cl.Email)
						ch <- prometheus.MustNewConstMetric(c.clientDownDesc, prometheus.GaugeValue, float64(cl.Down),
							strconv.Itoa(cl.ID), cl.Email)
					}
				} else {
					// Top N by upload
					sortedUp := make([]struct {
						ID    int
						Email string
						Up    int64
						Down  int64
					}, len(clients))
					for i, cl := range clients {
						sortedUp[i] = struct {
							ID    int
							Email string
							Up    int64
							Down  int64
						}{cl.ID, cl.Email, cl.Up, cl.Down}
					}
					sort.Slice(sortedUp, func(i, j int) bool { return sortedUp[i].Up > sortedUp[j].Up })
					for i := 0; i < n && i < len(sortedUp); i++ {
						ch <- prometheus.MustNewConstMetric(c.clientUpDesc, prometheus.GaugeValue, float64(sortedUp[i].Up),
							strconv.Itoa(sortedUp[i].ID), sortedUp[i].Email)
					}
					// Top N by download
					sortedDown := make([]struct {
						ID    int
						Email string
						Up    int64
						Down  int64
					}, len(clients))
					for i, cl := range clients {
						sortedDown[i] = struct {
							ID    int
							Email string
							Up    int64
							Down  int64
						}{cl.ID, cl.Email, cl.Up, cl.Down}
					}
					sort.Slice(sortedDown, func(i, j int) bool { return sortedDown[i].Down > sortedDown[j].Down })
					for i := 0; i < n && i < len(sortedDown); i++ {
						ch <- prometheus.MustNewConstMetric(c.clientDownDesc, prometheus.GaugeValue, float64(sortedDown[i].Down),
							strconv.Itoa(sortedDown[i].ID), sortedDown[i].Email)
					}
				}
			}
		} else {
			c.logger.Println("3x-ui: failed to unmarshal inbounds list")
		}
	} else {
		c.logger.Printf("3x-ui: inbounds list request error: %v", err)
	}
}
