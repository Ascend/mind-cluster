/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package prometheushttp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"ascend-common/common-utils/hwlog"
	"huawei.com/npu-exporter/v6/utils/logger"
)

// init initializes the package logger so that error paths logging warnings
// (for example exceeding the metric family limit) do not dereference a nil logger.
func init() {
	logger.HwLogConfig = &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := logger.InitLogger(logger.PrometheusPlatform); err != nil {
		panic(err)
	}
}

type fetchFunc func(context.Context) ([]byte, error)

func (f fetchFunc) FetchMetrics(ctx context.Context) ([]byte, error) {
	return f(ctx)
}

func TestCollectorPublishesOnlyValidUniqueGauges(t *testing.T) {
	body := `
# TYPE npu_load gauge
npu_load{chip="0",scope="node"} 0 1234
npu_load{scope="node",chip="1"} 4
npu_load{chip="1",scope="node"} NaN
npu_load{chip="3",extra="x"} 5
npu_load{chip="4",scope="node"} 101
# TYPE other_metric gauge
other_metric 8
`
	fetchErr := errors.New("offline")
	var errResponse error
	c, err := New(fetchFunc(func(context.Context) ([]byte, error) {
		return []byte(body), errResponse
	}), []GaugeSpec{{
		Name: "npu_load", Help: "load", LabelNames: []string{"scope", "chip"},
		Validate: func(value float64, _ []string) bool { return value >= 0 && value <= 100 },
	}}, 100)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.CollectToCache(t.Context()); err != nil {
		t.Fatal(err)
	}
	samples := c.GetSamples()
	if len(samples) != 1 || samples[0].Name != "npu_load" || samples[0].Labels["chip"] != "0" ||
		samples[0].Value != 0 || samples[0].Timestamp.IsZero() {
		t.Fatalf("unexpected snapshot: %+v", samples)
	}
	samples[0].Labels["chip"] = "changed"
	metrics := make(chan prometheus.Metric)
	go func() {
		c.UpdatePrometheus(metrics)
		close(metrics)
	}()
	var output []*dto.Metric
	for emitted := range metrics {
		metric := &dto.Metric{}
		if err := emitted.Write(metric); err != nil {
			t.Errorf("encode Prometheus metric: %v", err)
			continue
		}
		output = append(output, metric)
	}
	if len(output) != 1 {
		t.Fatalf("unexpected Prometheus output: %v", output)
	}
	metric := output[0]
	if metric.GetTimestampMs() != samples[0].Timestamp.UnixMilli() || metric.GetLabel()[0].GetValue() != "0" {
		t.Fatalf("unexpected sample: %v", metric)
	}
	errResponse = fetchErr
	if err := c.CollectToCache(t.Context()); !errors.Is(err, fetchErr) || len(c.GetSamples()) != 0 {
		t.Fatalf("fetch failure must clear the snapshot, got %v, %v", err, c.GetSamples())
	}

	errResponse = nil
	body = "# TYPE npu_load gauge\nnpu_load{scope=\"node\",chip=\"5\"} 12\n"
	if err := c.CollectToCache(t.Context()); err != nil {
		t.Fatal(err)
	}
	if samples := c.GetSamples(); len(samples) != 1 || samples[0].Value != 12 {
		t.Fatalf("expected refreshed snapshot, got %+v", samples)
	}
	body = "# TYPE npu_load gauge\nnpu_load{\n"
	if err := c.CollectToCache(t.Context()); err == nil || len(c.GetSamples()) != 0 {
		t.Fatalf("parse failure must clear the snapshot, got %v, %v", err, c.GetSamples())
	}
}

func TestCollectorReplacesSnapshotWhenAllocationsDisappear(t *testing.T) {
	const allocation = "# TYPE allocation_info gauge\nallocation_info 1\n"
	body := allocation
	c, err := New(fetchFunc(func(context.Context) ([]byte, error) {
		return []byte(body), nil
	}), []GaugeSpec{{Name: "allocation_info", Help: "Allocation presence"}}, 1)
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range []struct {
		body    string
		samples int
		wantErr bool
	}{
		{allocation, 1, false},
		{"", 0, false},
		{allocation, 1, false},
		{allocation + "# TYPE unknown gauge\nunknown 1\n", 0, true},
	} {
		body = step.body
		if err := c.CollectToCache(t.Context()); (err != nil) != step.wantErr {
			t.Fatalf("unexpected refresh result for %q: %v", body, err)
		}
		if got := len(c.GetSamples()); got != step.samples {
			t.Fatalf("expected %d samples for %q, got %d", step.samples, body, got)
		}
		if got := c.SampleCount(); got != step.samples {
			t.Fatalf("expected snapshot count %d for %q, got %d", step.samples, body, got)
		}
	}
}

func TestCollectorUsesOneTimestampAcrossFamilies(t *testing.T) {
	body := "# TYPE first gauge\nfirst 1 1234\n# TYPE second gauge\nsecond 2 5678\n"
	collector, err := New(fetchFunc(func(context.Context) ([]byte, error) { return []byte(body), nil }),
		[]GaugeSpec{{Name: "first", Help: "First gauge"}, {Name: "second", Help: "Second gauge"}}, 2)
	if err != nil {
		t.Fatal(err)
	}
	before := time.Now()
	if err := collector.CollectToCache(t.Context()); err != nil {
		t.Fatal(err)
	}
	after := time.Now()
	samples := collector.GetSamples()
	if len(samples) != 2 {
		t.Fatalf("expected two samples, got %v", samples)
	}
	for _, sample := range samples {
		if !sample.Timestamp.Equal(samples[0].Timestamp) || sample.Timestamp.Before(before) || sample.Timestamp.After(after) {
			t.Fatalf("unexpected collection timestamp: %v", sample.Timestamp)
		}
	}
}
