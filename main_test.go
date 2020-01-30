package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/xsteadfastx/smtpd-exporter/mocks"
)

func TestMetricCreate(t *testing.T) {
	tables := []struct {
		out string
		m   Metric
	}{
		{
			`bounce.envelope=0
            scheduler.delivery.ok=5318
            scheduler.delivery.permfail=972
            scheduler.delivery.tempfail=4
            uptime.human=11d19h42m11s`,
			Metric{
				Name:  "smtpd_delivery_ok",
				Help:  "Shows how often a delivery was ok",
				Regex: `scheduler\.delivery\.ok=(?P<number>\d+)`,
			},
		},
		{
			`bounce.envelope=0
            uptime.human=11d19h42m11s`,
			Metric{
				Name:  "smtpd_delivery_ok",
				Help:  "Shows how often a delivery was ok",
				Regex: `scheduler\.delivery\.ok=(?P<number>\d+)`,
			},
		},
	}

	for _, table := range tables {
		m := Metric{
			Name:  "smtpd_delivery_ok",
			Help:  "Shows how often a delivery was ok",
			Regex: `scheduler\.delivery\.ok=(?P<number>\d+)`,
		}
		if table.m.Name != m.Name || table.m.Help != m.Help || table.m.Regex != m.Regex {
			t.Errorf("not the same")
			t.Errorf("'%s' and '%s'", table.m.Name, m.Name)
			t.Errorf("'%s' and '%s'", table.m.Help, m.Help)
			t.Errorf("'%s' and '%s'", table.m.Regex, m.Regex)
		}
	}
}

func TestMetricValue(t *testing.T) {
	tables := []struct {
		out    string
		values []int
	}{
		{
			`bounce.envelope=0
            scheduler.delivery.ok=5318
            scheduler.delivery.permfail=972
            scheduler.delivery.tempfail=4
            uptime.human=11d19h42m11s`,
			[]int{5318, 972, 4},
		},
		{
			`bounce.envelope=0
            scheduler.delivery.ok=5318
            uptime.human=11d19h42m11s`,
			[]int{5318, -1, -1},
		},
	}

	for _, table := range tables {
		var values []int
		for _, m := range metrics {
			value, _ := m.value(table.out)
			values = append(values, value)
		}
		if !cmp.Equal(table.values, values) {
			t.Errorf("%+v is not %+v", table.values, values)
		}
	}
}

func TestCollectValues(t *testing.T) {
	mockStat := &mocks.Stat{}
	mockGauge := &mocks.Gauge{}
}
