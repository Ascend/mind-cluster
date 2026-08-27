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

package metricscollector

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

// adapterDummy is a minimal collector used to exercise the adapter defaults.
type adapterDummy struct {
	MetricsCollectorAdapter
}

// TestMetricsCollectorAdapterDefaults verifies all default no-op implementations.
func TestMetricsCollectorAdapterDefaults(t *testing.T) {
	var c MetricsCollector = &adapterDummy{}

	descCh := make(chan *prometheus.Desc, 1)
	c.Describe(descCh) // no-op: must not block or panic
	close(descCh)

	metricCh := make(chan prometheus.Metric, 1)
	c.UpdatePrometheus(metricCh, nil) // no-op
	close(metricCh)
	for range descCh {
		t.Error("Describe should not send anything")
	}
	for range metricCh {
		t.Error("UpdatePrometheus should not send anything")
	}

	c.CollectToCache(nil) // no-op
	c.PreCollect(nil)     // no-op
	c.PostCollect(nil)    // no-op

	if !c.IsSupported(nil) {
		t.Error("IsSupported default = false, want true")
	}
	if got := c.GetInterval(); got != 0 {
		t.Errorf("GetInterval default = %v, want 0", got)
	}
}

// TestGetCacheKey covers cache key derivation for all input kinds.
func TestGetCacheKey(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{"ptr to struct", &adapterDummy{}, "adapterDummy"},
		{"struct value", adapterDummy{}, ""},
		{"ptr to non-struct", new(int), ""},
		{"nil", nil, ""},
	}
	for _, tt := range tests {
		if got := GetCacheKey(tt.input); got != tt.want {
			t.Errorf("GetCacheKey(%s) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

// TestBuildDescs covers the three Desc builders and their label sets.
func TestBuildDescs(t *testing.T) {
	tests := []struct {
		name   string
		desc   *prometheus.Desc
		fq     string
		labels []string
	}{
		{"BuildDesc", BuildDesc("dpu_metric", "help"), "dpu_metric", DpuCardLabels},
		{"BuildIfaceDesc", BuildIfaceDesc("iface_metric", "help"), "iface_metric", DpuIfaceLabels},
		{"BuildDescWithLabels", BuildDescWithLabels("custom_metric", "help", []string{"a", "b"}), "custom_metric", []string{"a", "b"}},
	}
	for _, tt := range tests {
		s := tt.desc.String()
		if !strings.Contains(s, tt.fq) {
			t.Errorf("%s: desc %q missing fqName %q", tt.name, s, tt.fq)
		}
		for _, l := range tt.labels {
			if !strings.Contains(s, l) {
				t.Errorf("%s: desc %q missing label %q", tt.name, s, l)
			}
		}
	}
}

// TestLabelConstants guards the exported label name contract.
func TestLabelConstants(t *testing.T) {
	if CardLabelName != "card" || CardTypeLabel != "card_type" || InterfaceLabel != "interface" {
		t.Errorf("label constants changed: %q %q %q", CardLabelName, CardTypeLabel, InterfaceLabel)
	}
	if len(DpuCardLabels) != 2 || len(DpuIfaceLabels) != 2 {
		t.Errorf("DpuCardLabels=%v DpuIfaceLabels=%v, want 2 labels each", DpuCardLabels, DpuIfaceLabels)
	}
}
