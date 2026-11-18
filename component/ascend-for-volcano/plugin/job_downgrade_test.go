/*
Copyright(C)2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package plugin is using for HuaWei Ascend pin affinity schedule.
*/
package plugin

import (
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/downgrade"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

// mockDowngradePolicyHandler extends mockPolicyHandler with the optional
// downgrade hook and a controllable validation result, tracking whether the
// glue invoked the hook after the validation passed.
type mockDowngradePolicyHandler struct {
	mockPolicyHandler
	validResult  *api.ValidateResult
	degradeJobID api.JobID
	degradeHits  int
}

func (m *mockDowngradePolicyHandler) ValidNPUJob() *api.ValidateResult { return m.validResult }

func (m *mockDowngradePolicyHandler) DowngradeConstraint(jobID api.JobID) {
	m.degradeHits++
	m.degradeJobID = jobID
}

type validJobDowngradeHookTest struct {
	name        string
	annotation  map[string]string
	reqNPUName  string
	validResult *api.ValidateResult // nil makes ValidNPUJob pass
	wantPass    bool
	wantHits    int
}

func buildValidJobDowngradeHookTestCases() []validJobDowngradeHookTest {
	enabledAnno := map[string]string{util.SchedulerDowngradeTimeoutAnnoKey: "60"}
	return []validJobDowngradeHookTest{
		{
			name:       "01 downgrade enabled job triggers the hook after validation",
			annotation: enabledAnno,
			reqNPUName: "huawei.com/Ascend910",
			wantPass:   true,
			wantHits:   1,
		},
		{
			name:       "02 downgrade disabled job skips the hook",
			reqNPUName: "huawei.com/Ascend910",
			wantPass:   true,
		},
		{
			name:        "03 failed validation skips the hook",
			annotation:  enabledAnno,
			reqNPUName:  "huawei.com/Ascend910",
			validResult: &api.ValidateResult{Pass: false},
		},
		{
			name:       "04 virtual device job returns before the hook",
			annotation: enabledAnno,
			reqNPUName: "huawei.com/Ascend910-16c",
			wantPass:   true,
		},
	}
}

// TestValidJobFnDegradeConstraintHook verifies the constraint downgrade is
// invoked right after the job validation passes and only for the jobs
// enabling the downgrade.
func TestValidJobFnDegradeConstraintHook(t *testing.T) {
	const hookJobID = api.JobID("downgrade-hook-job")
	const hookWindow = 60
	for _, tt := range buildValidJobDowngradeHookTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			downgrade.ResetDowngradeCacheForTest()
			if tt.wantHits > 0 {
				downgrade.StampSession(map[api.JobID]bool{hookJobID: true},
					map[api.JobID]int{hookJobID: hookWindow})
			}
			sJob := &SchedulerJob{}
			sJob.SchedulerJobAttr = util.SchedulerJobAttr{
				ComJob: util.ComJob{
					Name:       hookJobID,
					Annotation: tt.annotation,
				},
				NPUJob: &util.NPUJob{ReqNPUName: tt.reqNPUName},
			}
			handler := &mockDowngradePolicyHandler{validResult: tt.validResult}
			sJob.policyHandler = handler

			result := sJob.validJobFn()

			if (result == nil) != tt.wantPass {
				t.Errorf("validJobFn() result = %#v, wantPass %v", result, tt.wantPass)
			}
			if handler.degradeHits != tt.wantHits {
				t.Errorf("DegradeConstraint hits = %d, want %d", handler.degradeHits, tt.wantHits)
			}
			if handler.degradeHits > 0 && handler.degradeJobID != sJob.Name {
				t.Errorf("DegradeConstraint jobID = %q, want %q", handler.degradeJobID, sJob.Name)
			}
		})
	}
}
