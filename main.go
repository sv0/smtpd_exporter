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

var res = map[string]string{
	"SchedularDeliveryOk":       `scheduler\.delivery\.ok=(?P<number>\d+)`,
	"SchedularDeliveryPermfail": `scheduler\.delivery\.permfail=(?P<number>\d+)`,
	"SchedularDeliveryTempfail": `scheduler\.delivery\.tempfail=(?P<number>\d+)`,
}

var deliveryOk = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "smtpd_delivery_ok",
		Help: "Shows how often a delivery was ok",
	},
)

type metric struct {
	SchedularDeliveryOk       int
	SchedularDeliveryPermfail int
	SchedularDeliveryTempfail int
}

func (m *metric) fill(out string) {
	matches := make(map[string]int)

	for k, v := range res {
		re, err := regexp.Compile(v)
		if err != nil {
			log.Panic(err)
		}

		match := re.FindStringSubmatch(out)
		// only go further if at least are two items in slice
		minMatch := 2
		if len(match) != minMatch {
			log.WithFields(log.Fields{"re": v}).Panic("could not match")
		}

		// convert to int
		ival, err := strconv.Atoi(match[1])

		if err != nil {
			log.Panic(err)
		}

		matches[k] = ival
	}

	m.SchedularDeliveryOk = matches["SchedularDeliveryOk"]
	m.SchedularDeliveryPermfail = matches["SchedularDeliveryPermfail"]
	m.SchedularDeliveryTempfail = matches["SchedularDeliveryTempfail"]
}

func (m *metric) exec() string {
	out, err := exec.Command("smtpctl", "show", "stats").Output()
	if err != nil {
		log.Fatal(err)
	}

	log.Debug(string(out))

	return string(out)
}

func (m *metric) set() {
	deliveryOk.Set(float64(m.SchedularDeliveryOk))
}

func collect(sleepTime *int) {
	dur := time.Duration(*sleepTime)

	for {
		m := metric{}
		out := m.exec()
		m.fill(out)
		m.set()
		time.Sleep(dur * time.Second)
	}
}

func main() {
	debug := flag.Bool("debug", false, "enable debug")
	execTime := flag.Int("execTime", 1, "seconds to wait before scraping")
	port := flag.Int("port", 8080, "port to listen on")

	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	prometheus.MustRegister(deliveryOk)
	deliveryOk.Set(0)

	go collect(execTime)

	http.Handle("/metrics", promhttp.Handler())
	log.Info(fmt.Sprintf("Beginning to serve on port :%d", *port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
