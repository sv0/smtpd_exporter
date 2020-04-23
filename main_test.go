package main

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
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
	mockStat := new(MockStat)
	mockStat.On("Now").Return(out, nil)

	err := collectValues(metrics, mockStat)

	assert.Nil(err)

	for i, m := range metrics {
		assert.Equal(m.LastVal, setValues[i])
	}

	mockStat.AssertExpectations(t)
}

// TestCollectValueZero tries to reproduce a bug of
// not setting the counter.
func TestCollectValuesZero(t *testing.T) {
	assert := assert.New(t)
	mockStat := new(MockStat)
	m := &Metric{
		Name:  "testmetric",
		Help:  "just a test",
		Regex: `scheduler\.delivery\.ok=(?P<number>\d+)`,
	}
	i := initer{}
	i.Metric(m)
	metrics := []*Metric{m}
	out := `
        scheduler.delivery.ok=5318
        scheduler.delivery.permfail=972
        scheduler.delivery.tempfail=4
    `
	mockStat.On("Now").Return(out, nil)
	err := collectValues(metrics, mockStat)
	assert.Nil(err)
	assert.Equal(float64(5318), testutil.ToFloat64(metrics[0].Counter)) //nolint:gomnd
}

func TestCalcAddVal(t *testing.T) {
	assert := assert.New(t)
	tables := []struct {
		lastVal      int
		value        int
		expected     int
		unregistered bool
	}{
		{0, 100, 100, false},
		{100, 150, 50, false},
		{100, 50, 50, true},
	}

	for _, table := range tables {
		r := new(mocks.Registerer)
		c := new(mocks.Counter)
		i := new(MockInitializer)
		m := &Metric{
			LastVal:    table.lastVal,
			Counter:    c,
			Registerer: r,
		}

		if table.unregistered {
			r.On("Unregister", m.Counter).Return(true)
			i.On("Metric", m).Return(nil)
		}

		assert.Equal(m.calcAddVal(table.value, i), table.expected)
	}
}
