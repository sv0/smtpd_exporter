package main

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// func TestMetricFill(t *testing.T) {
// 	tables := []struct {
// 		out string
// 		m   metric
// 	}{
// 		{
// 			`bounce.envelope=0
//             bounce.message=0
//             bounce.session=0
//             control.session=1
//             mta.connector=0
//             mta.domain=0
//             mta.envelope=0
//             mta.host=0
//             mta.relay=0
//             mta.route=0
//             mta.session=0
//             mta.source=0
//             mta.task=0
//             mta.task.running=0
//             queue.bounce=486
//             queue.evpcache.load.hit=12726
//             queue.evpcache.size=0
//             queue.evpcache.update.hit=4
//             scheduler.delivery.ok=5318
//             scheduler.delivery.permfail=972
//             scheduler.delivery.tempfail=4
//             scheduler.envelope=0
//             scheduler.envelope.incoming=0
//             scheduler.envelope.inflight=0
//             scheduler.ramqueue.envelope=0
//             scheduler.ramqueue.hold=0
//             scheduler.ramqueue.holdq=0
//             scheduler.ramqueue.message=0
//             scheduler.ramqueue.update=0
//             smtp.session=0
//             smtp.session.inet4=5707
//             smtp.session.local=507
//             smtp.tls=0
//             uptime=1021331
//             uptime.human=11d19h42m11s`,
// 			metric{
// 				SchedularDeliveryOk:       5318,
// 				SchedularDeliveryPermfail: 972,
// 				SchedularDeliveryTempfail: 4,
// 			},
// 		},
// 		{
// 			`bounce.envelope=0
//             scheduler.delivery.ok=5318
//             uptime.human=11d19h42m11s`,
// 			metric{
// 				SchedularDeliveryOk: 5318,
// 			},
// 		},
// 	}

// 	for _, table := range tables {
// 		m := metric{}
// 		m.fill(table.out)
// 		if !cmp.Equal(m, table.m) {
// 			t.Errorf("%+v and %+v are not equal", m, table.m)
// 		}
// 	}
// }

func TestMetricCreate(t *testing.T) {
	tables := []struct {
		out   string
		m     Metric
		avail bool
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
			true,
		},
		{
			`bounce.envelope=0
            uptime.human=11d19h42m11s`,
			Metric{
				Name:  "smtpd_delivery_ok",
				Help:  "Shows how often a delivery was ok",
				Regex: `scheduler\.delivery\.ok=(?P<number>\d+)`,
			},
			false,
		},
	}

	for _, table := range tables {
		m := Metric{
			Name:  "smtpd_delivery_ok",
			Help:  "Shows how often a delivery was ok",
			Regex: `scheduler\.delivery\.ok=(?P<number>\d+)`,
		}
		m.create(table.out)
		if table.m.Name != m.Name || table.m.Help != m.Help || table.m.Regex != m.Regex {
			t.Errorf("not the same")
			t.Errorf("'%s' and '%s'", table.m.Name, m.Name)
			t.Errorf("'%s' and '%s'", table.m.Help, m.Help)
			t.Errorf("'%s' and '%s'", table.m.Regex, m.Regex)
		}
		if table.avail {
			if m.Gauge == nil {
				t.Errorf("Metric.Gauge is nil")
			}
			if reflect.ValueOf(m.Gauge).Kind() != reflect.Ptr {
				t.Errorf("Metric.Gauge is no pointer")
			}
			if !m.Available {
				t.Errorf("Metric.Available is not true")
			}
		} else {
			if m.Gauge != nil {
				t.Errorf("Metric.Gauge is not nil")
			}
			if m.Available {
				t.Errorf("Metric.Available is true")
			}
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
			m.create(table.out)
			value, _ := m.value(table.out)
			values = append(values, value)
		}
		if !cmp.Equal(table.values, values) {
			t.Errorf("%+v is not %+v", table.values, values)
		}
	}
}
