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

package silentfault

import (
	"encoding/json"
	"os"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/api"
	"ascend-common/api/annotation"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/domain/conf"
)

// TestMain initializes logging and then runs all unit tests in the package.
func TestMain(m *testing.M) {
	_ = hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, nil)
	os.Exit(m.Run())
}

// newNodeWithCards returns a fake K8s node with the given total card count (capacity) for the resource type.
func newNodeWithCards(resourceType string, total int) *corev1.Node {
	return &corev1.Node{Status: corev1.NodeStatus{
		Capacity: corev1.ResourceList{
			corev1.ResourceName(api.ResourceNamePrefix + resourceType): *resource.NewQuantity(int64(total), resource.DecimalSI),
		},
	}}
}

// newNodeWithBaseDevInfos returns a fake node annotated with huawei.com/npu.base-device-infos
// whose value maps the given "<type>-<id>" device names to empty objects.
func newNodeWithBaseDevInfos(devNames ...string) *corev1.Node {
	m := make(map[string]struct{}, len(devNames))
	for _, name := range devNames {
		m[name] = struct{}{}
	}
	b, _ := json.Marshal(m)
	return &corev1.Node{ObjectMeta: metav1.ObjectMeta{
		Annotations: map[string]string{annotation.NPUBaseDevInfosAnnotation: string(b)},
	}}
}

// enableSilentFault enables the silent fault switch and writes a set of valid parameters for reuse across test cases.
func enableSilentFault() {
	policy := conf.SilentFaultPolicy{Enabled: true}
	policy.Detect.ConsecutiveTimes = 3
	policy.Detect.HardwareFaultWindowSecond = 30
	policy.Detect.WindowSecond = 3600
	policy.Detect.MinTaskCards = 1
	policy.Release.FaultFreeSecond = 172800
	conf.SetSilentFaultPolicy(policy)
}
