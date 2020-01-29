package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMetricFill(t *testing.T) {
	tables := []struct {
		out string
		m   metric
	}{
		{
			`bounce.envelope=0
            bounce.message=0
            bounce.session=0
            control.session=1
            mta.connector=0
            mta.domain=0
            mta.envelope=0
            mta.host=0
            mta.relay=0
            mta.route=0
            mta.session=0
            mta.source=0
            mta.task=0
            mta.task.running=0
            queue.bounce=486
            queue.evpcache.load.hit=12726
            queue.evpcache.size=0
            queue.evpcache.update.hit=4
            scheduler.delivery.ok=5318
            scheduler.delivery.permfail=972
            scheduler.delivery.tempfail=4
            scheduler.envelope=0
            scheduler.envelope.incoming=0
            scheduler.envelope.inflight=0
            scheduler.ramqueue.envelope=0
            scheduler.ramqueue.hold=0
            scheduler.ramqueue.holdq=0
            scheduler.ramqueue.message=0
            scheduler.ramqueue.update=0
            smtp.session=0
            smtp.session.inet4=5707
            smtp.session.local=507
            smtp.tls=0
            uptime=1021331
            uptime.human=11d19h42m11s`,
			metric{
				SchedularDeliveryOk:       5318,
				SchedularDeliveryPermfail: 972,
				SchedularDeliveryTempfail: 4,
			},
		},
		{
			`bounce.envelope=0
            scheduler.delivery.ok=5318
            uptime.human=11d19h42m11s`,
			metric{
				SchedularDeliveryOk: 5318,
			},
		},
	}

	for _, table := range tables {
		m := metric{}
		m.fill(table.out)
		if !cmp.Equal(m, table.m) {
			t.Errorf("%+v and %+v are not equal", m, table.m)
		}
	}
}
