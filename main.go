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

var (
	debug    = flag.Bool("debug", false, "enable debug")
	execTime = flag.Int("execTime", 1, "seconds to wait before scraping")
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

type Metric struct {
	Name      string
	Help      string
	Regex     string
	Available bool
	Gauge     prometheus.Gauge
}

func (m *Metric) create(out string) {
	if matched, err := regexp.MatchString(m.Regex, out); (err != nil) || !matched {
		if err != nil {
			log.WithFields(log.Fields{"metric": m.Name}).Error(err)
		} else {
			log.WithFields(
				log.Fields{"metric": m.Name, "regex": m.Regex},
			).Info("could not find metric")
			log.WithFields(
				log.Fields{"metric": m.Name, "regex": m.Regex, "out": out},
			).Debug("could not find metric in output")
		}

		m.Available = false

		return
	}

	m.Available = true
	m.Gauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: m.Name,
			Help: m.Help,
		},
	)
}

// value extracts the needed value out of the output of the smtpctl command
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

type stat interface {
	now() (string, error)
}

type smtpctl struct{}

func (s smtpctl) now() (string, error) {
	out, err := exec.Command("smtpctl", "show", "stats").Output()
	if err != nil {
		log.Error(err)
		return "", err
	}

	log.Debug(string(out))

	return string(out), nil
}

func collect(sleepTime *int, stats stat) {
	dur := time.Duration(*sleepTime)

	for {
		collectValues(stats)

		time.Sleep(dur * time.Second)
	}
}

func collectValues(stats stat) {
	out, err := stats.now()
	if err != nil {
		log.Error(err)
		return
	}

	for _, m := range metrics {
		if m.Available {
			log.WithFields(log.Fields{"metric": fmt.Sprintf("%+v", m)}).Debug("using metric")

			value, err := m.value(out)

			if err != nil {
				log.WithFields(log.Fields{"metric": m}).Error(err)
			}

			m.Gauge.Set(float64(value))
		}
	}
}

func main() {
	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	stats := smtpctl{}
	out, err := stats.now()

	if err != nil {
		log.Fatal("could not get stats")
	}

	for _, m := range metrics {
		m.create(out)
		log.WithFields(log.Fields{"metric": fmt.Sprintf("%+v", m)}).Debug("created metric")

		if m.Available {
			prometheus.MustRegister(m.Gauge)
		}
	}

	go collect(execTime, stats)

	http.Handle("/metrics", promhttp.Handler())
	log.Info(fmt.Sprintf("Beginning to serve on port :%d", *port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
