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

package plugins

import (
	"context"
	"testing"

	"huawei.com/npu-exporter/v6/collector/prometheushttp"
)

type textMetrics string

func (m textMetrics) FetchMetrics(context.Context) ([]byte, error) {
	return []byte(m), nil
}

func TestManagerMetricsContract(t *testing.T) {
	const metrics = `# TYPE enpu_allocation_info gauge
enpu_allocation_info{allocation_id="alloc-1",phy_id="0",vnpu_id="1",pod_uid="pod-1",container_name=""} 1 1793066400000
# TYPE enpu_vnpu_metric_valid gauge
enpu_vnpu_metric_valid{allocation_id="alloc-1",phy_id="0",vnpu_id="1",pod_uid="pod-1",container_name="",metric="hbm_used_bytes"} 0 1793066400000
enpu_vnpu_metric_valid{allocation_id="alloc-1",phy_id="0",vnpu_id="1",pod_uid="pod-1",container_name="",metric="unknown"} 1 1793066400000
# TYPE enpu_npu_utilization_percent gauge
enpu_npu_utilization_percent{phy_id="0"} 101 1793066400000
`
	specs := enpuGaugeSpecs()
	if len(specs) != enpuMetricFamilyCount {
		t.Fatalf("expected 16 documented metric families, got %d", len(specs))
	}
	collector, err := prometheushttp.New(textMetrics(metrics), specs, enpuFamilyLimit)
	if err != nil {
		t.Fatal(err)
	}
	if err := collector.CollectToCache(t.Context()); err != nil {
		t.Fatal(err)
	}
	samples := collector.GetSamples()
	if len(samples) != 2 {
		t.Fatalf("expected the allocation and its invalid-sample indicator, got %+v", samples)
	}
	for _, sample := range samples {
		if name, ok := sample.Labels["container_name"]; !ok || name != "" ||
			sample.Labels["pod_uid"] != "pod-1" || sample.Timestamp.IsZero() || !sample.Timestamp.Equal(samples[0].Timestamp) {
			t.Fatalf("unexpected sample metadata: %+v", sample)
		}
	}
}
