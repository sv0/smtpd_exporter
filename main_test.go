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
	assert := assert.New(t)
	setValues := []int{5318, 972, 4}
	out := `
        scheduler.delivery.ok=5318
        scheduler.delivery.permfail=972
        scheduler.delivery.tempfail=4
    `
	// init gauge mocks
	for i, m := range metrics {
		c := new(mocks.Counter)
		// sets expectations on Set method and returns nil
		c.On("Add", float64(setValues[i])).Return(nil)
		m.Counter = c

		assert.Equal(m.LastVal, 0)
	}
	// create stat mock
	mockStat := new(mocks.Stat)
	mockStat.On("Now").Return(out, nil)

	collectValues(mockStat)

	for i, m := range metrics {
		assert.Equal(m.LastVal, setValues[i])
	}

	mockStat.AssertExpectations(t)
}

func TestCalcAddVal(t *testing.T) {
	assert := assert.New(t)
	tables := []struct {
		lastVal      int
		value        int
		expected     int
		unregistered bool
	}{
		// {0, 100, 100, false},
		// {100, 150, 50, false},
		{100, 50, 50, true},
	}

	for _, table := range tables {
		r := new(mocks.Registerer)
		c := new(mocks.Counter)
		m := Metric{
			LastVal:    table.lastVal,
			Counter:    c,
			Registerer: r,
		}
		r.On("Unregister", m.Counter).Return(true)
		r.On("MustRegister", m.Counter).Return(nil)
		assert.Equal(m.calcAddVal(table.value), table.expected)
		if table.unregistered {
			r.AssertCalled(t, "Unregister")
		}
	}
}
