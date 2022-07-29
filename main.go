package main

import (
	"bufio"
	"flag"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	gpsdAddr      = flag.String("d", "localhost:2947", "gpsd address")
	metricsListen = flag.String("l", ":9100", "metrics listen address")
	verbose       = flag.Bool("v", false, "verbose")
	trace         = flag.Bool("vv", false, "trace")
)

var (
	metricLastPoll = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gpsd_last_poll",
		Help: "Last time the GPSD daemon was polled",
	})
	metricVersion = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gpsd_version",
		Help: "GPSD version",
	}, []string{"version"})
)

var (
	dynMetricGauges    = map[string]prometheus.Gauge{}
	dynMetricGaugeVecs = map[string]*prometheus.GaugeVec{}
)

func main() {
	flag.Parse()
	if *verbose {
		log.SetLevel(log.DebugLevel)
		log.Debug("Running in verbose mode")
	}
	if *trace {
		log.SetLevel(log.TraceLevel)
		log.Debug("Running in trace mode")
	}

	var conn net.Conn
	var scanner *bufio.Scanner
	go func() {
		for {
			if conn == nil {
				log.Infof("Connecting to %s", *gpsdAddr)
				var err error
				conn, err = net.Dial("tcp", *gpsdAddr)
				if err != nil {
					log.Fatal(err)
				}
				if _, err := conn.Write([]byte("?WATCH={\"enable\": true}\n?POLL;\n")); err != nil {
					log.Warnf("Error sending POLL command: %v", err)
					_ = conn.Close()
					conn = nil
				}
				rdr := bufio.NewReader(conn)
				scanner = bufio.NewScanner(rdr)
				scanner.Split(bufio.ScanLines)
			}
			for scanner.Scan() {
				processLine(scanner.Text())
			}
		}
	}()

	// Poll for updates
	pollTicker := time.NewTicker(time.Second * 1)
	go func() {
		log.Debug("Starting poll ticker")
		for range pollTicker.C {
			if conn != nil {
				log.Debug("Sending POLL command")
				if _, err := conn.Write([]byte("?WATCH={\"enable\": true}\n?POLL;\n")); err != nil {
					log.Warnf("Error sending POLL command: %v", err)
					_ = conn.Close()
					conn = nil
				}
				metricLastPoll.SetToCurrentTime()
			} else {
				log.Debug("Not connected, not sending POLL command")
			}
		}
	}()

	// Metrics server
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	log.Infof("Starting metrics exporter on %s/metrics", *metricsListen)
	log.Fatal(http.ListenAndServe(*metricsListen, metricsMux))
}