package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xsteadfastx/smtpd-exporter/mocks"
)

func TestMetricCreate(t *testing.T) {
	assert := assert.New(t)
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
		assert.Equal(table.m.Name, m.Name)
		assert.Equal(table.m.Help, m.Help)
		assert.Equal(table.m.Regex, m.Regex)
	}
}

func TestMetricValue(t *testing.T) {
	assert := assert.New(t)
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
		assert.Equal(table.values, values)
	}
}

func TestCollectValues(t *testing.T) {
	setValues := []int{5318, 972, 4}
	out := `
        scheduler.delivery.ok=5318
        scheduler.delivery.permfail=972
        scheduler.delivery.tempfail=4
    `
	// init gauge mocks
	for i, m := range metrics {
		g := new(mocks.Gauge)
		g.On("Set", float64(setValues[i])).Return(nil)
		m.Gauge = g
	}
	// create stat mock
	mockStat := new(mocks.Stat)
	mockStat.On("Now").Return(out, nil)

	collectValues(mockStat)
}
