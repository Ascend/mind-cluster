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

package server

import (
	"context"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"Ascend-device-plugin/pkg/enpuconfig"
	"ascend-common/api"
	"ascend-common/enpumanager"
)

type allocationEnsurerFunc func(context.Context, enpumanager.AllocateRequest) (*enpuconfig.AllocationResult, error)

func (f allocationEnsurerFunc) EnsureAllocation(ctx context.Context,
	input enpumanager.AllocateRequest) (*enpuconfig.AllocationResult, error) {
	return f(ctx, input)
}

func TestEnpuConfigProviderReturnsSafeDirectory(t *testing.T) {
	config := enpuconfig.ConfigRoot + "/test/npu_info.config"
	const npuResource = v1.ResourceName(api.HuaweiAscend910)
	inference := v1.Container{
		Name: "infer",
		Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{npuResource: resource.MustParse("30")},
		},
	}
	input := SoftShareCreateInput{
		pod: &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{UID: "test-pod"},
			Spec: v1.PodSpec{Containers: []v1.Container{
				{Name: "sidecar"}, inference,
			}},
		},
		resourceName: npuResource,
		physicalID:   3, dieID: "die3",
		annotations: softShareDevAnnotations{aicoreQuota: "30", hbmQuota: "256", schedulingPolicy: "1"},
	}
	for _, tc := range []struct {
		name      string
		path      string
		shmID     string
		dieID     string
		wantError bool
	}{
		{name: "new allocation", path: config, shmID: "die3"},
		{name: "reused allocation", path: config, dieID: "die3"},
		{name: "wrong filename", path: enpuconfig.ConfigRoot + "/test/other.config", shmID: "die3", wantError: true},
		{name: "outside whitelist", path: "/tmp/npu_info.config", shmID: "die3", wantError: true},
		{name: "shared memory mismatch", path: config, shmID: "other", wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			provider := &EnpuConfigProvider{client: allocationEnsurerFunc(func(_ context.Context, req enpumanager.AllocateRequest) (
				*enpuconfig.AllocationResult, error) {
				if req.PodUID != "test-pod" || req.ContainerName != "infer" || req.PhysicalID != 3 || req.AICoreQuota != 30 ||
					req.HBMQuota != 256 || req.SchedPolicy != 1 {
					t.Fatalf("unexpected allocation request: %+v", req)
				}
				return &enpuconfig.AllocationResult{AllocateResponse: enpumanager.AllocateResponse{ConfigPath: tc.path, ShmID: tc.shmID, DieID: tc.dieID}}, nil
			})}
			directory, err := provider.GetConfigDir(t.Context(), input)
			if tc.wantError {
				if err == nil || directory != "" {
					t.Fatalf("expected invalid allocation response, got %v, %v", directory, err)
				}
				return
			}
			if err != nil || directory != enpuconfig.ConfigRoot+"/test" {
				t.Fatalf("unexpected configuration directory %q, %v", directory, err)
			}
		})
	}
}
