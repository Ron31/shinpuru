package metrics

import (
	"net/http"
	"time"

	"github.com/go-ping/ping"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zekroTJA/shinpuru/internal/util"
)

var (
	DiscordEventTriggers = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "discord_eventtriggers_total",
		Help: "Total number of discord events triggered.",
	}, []string{"event"})

	DiscordGatewayPing = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "discord_gatewayping",
		Help: "The ping time in milliseconds to the discord API gateay.",
	})
)

// MetricsServer wraps a simple HTTP server serving
// a prometheus metrics endpoint.
type MetricsServer struct {
	server *http.Server
}

// NewMetricsServer initializes a new MectricsServer
// instance with the given addr and registers all
// instruments.
func NewMetricsServer(addr string) (ms *MetricsServer, err error) {
	prometheus.MustRegister(
		DiscordEventTriggers,
		DiscordGatewayPing)

	pw, err := NewPingWatcher(30 * time.Second)
	if err != nil {
		return
	}
	pw.OnElapsed = func(p *ping.Statistics, err error) {
		var v float64
		if err == nil && p != nil {
			v = float64(p.AvgRtt.Milliseconds())
		} else if err != nil {
			util.Log.Warningf("failed getting rtt to discord API: %s", err.Error())
		}
		DiscordGatewayPing.Set(v)
	}

	ms = new(MetricsServer)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	ms.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return
}

// ListenAndServeBlocking starts the listening loop of
// the web server which blocks the current goroutine.
func (ms *MetricsServer) ListenAndServeBlocking() error {
	return ms.server.ListenAndServe()
}