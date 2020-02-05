package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	log "github.com/sirupsen/logrus"
)

//go:generate mockery -name Stat

var (
	// Version string
	Version = "development"

	version  = flag.Bool("version", false, "version.")
	debug    = flag.Bool("debug", false, "enable debug.")
	interval = flag.Duration("interval", 1*time.Second, "seconds to wait before scraping.")
	port     = flag.Int("port", 9967, "port to listen on.")
	host     = flag.String("host", "localhost", "host to listen on.")
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

// Metric stores a metric to export and all it needed data.
type Metric struct {
	Name    string
	Help    string
	Regex   string
	Counter prometheus.Counter

	mux     sync.Mutex
	LastVal int
}

// value extracts the needed value out of the output of the smtpctl command
func (m *Metric) value(out string) (int, error) {
	re, err := regexp.Compile(m.Regex)
	if err != nil {
		return 0, fmt.Errorf("could not compile regex: %s", m.Regex)
	}

	match := re.FindStringSubmatch(out)
	// only go further if at least are two items in slice
	minMatch := 2
	if len(match) != minMatch {
		return 0, fmt.Errorf("could not match regex: %s", m.Regex)
	}

	// convert to int
	val, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, fmt.Errorf("could not convert to int: %s", match[1])
	}

	return val, nil
}

func (m *Metric) calcAddVal(value int) int {
	// this scenario should kick in if smtpd gets restarted and the extracted
	// value is smaller than the last stored value.
	// we unregister the counter and create a new one.
	if value < m.LastVal {
		prometheus.Unregister(m.Counter)
		initMetric(m)

		return value
	}

	if m.LastVal == 0 || value == 0 {
		return value
	}

	return (value - m.LastVal)
}

// Stat is an interface for getting some stats from a command.
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

		// just log error but still set counter to 0
		if err != nil {
			log.WithFields(log.Fields{"metric": m, "error": err}).Debug("could not get value")
		}

		// store last value in struct
		m.mux.Lock()
		addVal := m.calcAddVal(value)

		log.WithFields(log.Fields{"metric": m, "add": addVal, "value": value}).Debug("adds value")

		m.Counter.Add(float64(addVal))
		m.LastVal = value
		m.mux.Unlock()
	}

	return nil
}

// initMetric creates a counter for the metric struct and register it to prometheus.
func initMetric(m *Metric) {
	// init counter
	m.Counter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: m.Name,
			Help: m.Help,
		},
	)
	prometheus.MustRegister(m.Counter)
}

func createMetrics() {
	for _, m := range metrics {
		log.Debugf("%+v", m)
		initMetric(m)
		log.Debugf("%+v", m)
	}
}

// type Prom interface {
// 	MustRegister()
// 	NewCouter()
// 	CounterOpts()
// }

// type prom struct {
// 	p Prom
// }

func main() {
	flag.Parse()

	if *version {
		fmt.Printf("%s", Version)
		os.Exit(0)
	}

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	createMetrics()

	go collect(interval)

	http.Handle("/metrics", promhttp.Handler())
	log.Info(fmt.Sprintf("Beginning to serve on port :%d", *port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", *host, *port), nil))
}
