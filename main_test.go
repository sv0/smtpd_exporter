package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xsteadfastx/smtpd_exporter/mocks"
)

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
			[]int{5318, 0, 0},
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
		// sets expectations on Set method and returns nil
		g.On("Set", float64(setValues[i])).Return(nil)
		m.Gauge = g
	}
	// create stat mock
	mockStat := new(mocks.Stat)
	mockStat.On("Now").Return(out, nil)

	collectValues(mockStat)
}
