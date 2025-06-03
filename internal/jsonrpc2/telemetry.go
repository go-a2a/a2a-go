// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jsonrpc2

import (
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

var (
	rPCID              metric.Int64Counter
	rPCDirection       any // string
	startedCounter     metric.Int64Counter
	sentBytesGauge     metric.Int64Gauge
	receivedBytesGauge metric.Int64Gauge
	statusCode         any // string
	latency            metric.Float64Histogram
)

var metricOnce sync.Once

func newMetrics(m metric.Meter) {
	metricOnce.Do(func() {
		var err error

		startedCounter, err = m.Int64Counter("started",
			metric.WithDescription("Count of started RPCs"),
		)
		if err != nil {
			otel.Handle(err)
			startedCounter = noop.Int64Counter{}
		}

		sentBytesGauge, err = m.Int64Gauge("sent_bytes",
			metric.WithDescription("Bytes sent"),
		)
		if err != nil {
			otel.Handle(err)
			sentBytesGauge = noop.Int64Gauge{}
		}

		receivedBytesGauge, err = m.Int64Gauge("received_bytes",
			metric.WithDescription("Bytes received"),
		)
		if err != nil {
			otel.Handle(err)
			receivedBytesGauge = noop.Int64Gauge{}
		}
	})
}
