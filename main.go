package main

import (
	"flag"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	log "github.com/sirupsen/logrus"
)

//go:generate mockery -name Stat

var (
	debug    = flag.Bool("debug", false, "enable debug")
	interval = flag.Duration("interval", time.Duration(1), "seconds to wait before scraping")
	port     = flag.Int("port", 8080, "port to listen on")
)

var metrics = []*Metric{
	{
		Name:  "smtpd_delivery_ok",
		Help:  "Shows how often a delivery was ok.",
		Regex: `scheduler\.delivery\.ok=(?P<number>\d+)`,
	}, {
		Name:  "smtpd_delivery_permfail",
		Help:  "Shows how often a delivery permafailed.",
		Regex: `scheduler\.delivery\.permfail=(?P<number>\d+)`,
	}, {
		Name:  "smtd_delivery_tempfail",
		Help:  "Shows how often a delivery tempfailed.",
		Regex: `scheduler\.delivery\.tempfail=(?P<number>\d+)`,
	},
}

// TODO: value (int) to save current value to calculate the new value after daemon restart
type Metric struct {
	Name  string
	Help  string
	Regex string
	Gauge prometheus.Gauge
}

// value extracts the needed value out of the output of the smtpctl command
// TODO: dont send -1... send a zero for calculating
func (m *Metric) value(out string) (int, error) {
	re, err := regexp.Compile(m.Regex)
	if err != nil {
		return -1, fmt.Errorf("could not compile regex: %s", m.Regex)
	}

	match := re.FindStringSubmatch(out)
	// only go further if at least are two items in slice
	minMatch := 2
	if len(match) != minMatch {
		return -1, fmt.Errorf("could not match regex: %s", m.Regex)
	}

	// convert to int
	val, err := strconv.Atoi(match[1])
	if err != nil {
		return -1, fmt.Errorf("could not convert to int: %s", match[1])
	}

	return val, nil
}

type Stat interface {
	Now() (string, error)
}

type smtpctl struct{}

func (s smtpctl) Now() (string, error) {
	out, err := exec.Command("smtpctl", "show", "stats").Output()
	if err != nil {
		log.Error(err)
		return "", err
	}

	log.Debug(string(out))

	return string(out), nil
}

func collect(interval *time.Duration) {
	stats := smtpctl{}

	for {
		err := collectValues(stats)
		if err != nil {
			log.Error(err)
		}

		time.Sleep(*interval)
	}
}

func collectValues(stats Stat) error {
	out, err := stats.Now()
	if err != nil {
		return err
	}

	for _, m := range metrics {
		log.WithFields(log.Fields{"metric": fmt.Sprintf("%+v", m)}).Debug("using metric")

		value, err := m.value(out)

		if err != nil {
			log.WithFields(log.Fields{"metric": m}).Error(err)
		}

		m.Gauge.Set(float64(value))
	}

	return nil
}

func create() {
	for _, m := range metrics {
		m.Gauge = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: m.Name,
				Help: m.Help,
			},
		)
		prometheus.MustRegister(m.Gauge)
	}
}

func main() {
	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	create()

	go collect(interval)

	http.Handle("/metrics", promhttp.Handler())
	log.Info(fmt.Sprintf("Beginning to serve on port :%d", *port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
