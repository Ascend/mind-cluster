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

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

type downgradeCompatibilityTest struct {
	name       string
	annotation map[string]string
	label      map[string]string
	wantValid  bool
}

func buildDowngradeCompatibilityTestCases() []downgradeCompatibilityTest {
	return []downgradeCompatibilityTest{
		{
			name:       "01 downgrade with pod rescheduling is rejected",
			annotation: map[string]string{util.SchedulerDowngradeAnnoKey: "true"},
			label:      map[string]string{util.SinglePodTag: util.EnableFunc},
			wantValid:  false,
		},
		{
			name:       "02 downgrade with process recover is rejected",
			annotation: map[string]string{util.SchedulerDowngradeAnnoKey: "true"},
			label:      map[string]string{util.ProcessRecoverEnable: util.EnableFunc},
			wantValid:  false,
		},
		{
			name:       "03 downgrade without incompatible configuration passes",
			annotation: map[string]string{util.SchedulerDowngradeAnnoKey: "true"},
			label:      map[string]string{},
			wantValid:  true,
		},
		{
			name:      "04 disabled downgrade with pod rescheduling passes",
			label:     map[string]string{util.SinglePodTag: util.EnableFunc},
			wantValid: true,
		},
		{
			name:       "05 pod rescheduling turned off passes with downgrade",
			annotation: map[string]string{util.SchedulerDowngradeAnnoKey: "true"},
			label:      map[string]string{util.SinglePodTag: "off"},
			wantValid:  true,
		},
	}
}

func TestValidDowngradeCompatibility(t *testing.T) {
	for _, tt := range buildDowngradeCompatibilityTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			sJob := &SchedulerJob{}
			sJob.SchedulerJobAttr = util.SchedulerJobAttr{
				ComJob: util.ComJob{Annotation: tt.annotation, Label: tt.label},
			}
			if result := sJob.validDowngradeCompatibility(); (result == nil) != tt.wantValid {
				t.Errorf("validDowngradeCompatibility() result = %#v, wantValid %v", result, tt.wantValid)
			}
		})
	}
}

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
	return []validJobDowngradeHookTest{
		{
			name:       "01 downgrade enabled job triggers the hook after validation",
			annotation: map[string]string{util.SchedulerDowngradeAnnoKey: "true"},
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
			annotation:  map[string]string{util.SchedulerDowngradeAnnoKey: "true"},
			reqNPUName:  "huawei.com/Ascend910",
			validResult: &api.ValidateResult{Pass: false},
		},
		{
			name:       "04 virtual device job returns before the hook",
			annotation: map[string]string{util.SchedulerDowngradeAnnoKey: "true"},
			reqNPUName: "huawei.com/Ascend910-16c",
			wantPass:   true,
		},
	}
}

// TestValidJobFnDegradeConstraintHook verifies the constraint downgrade is
// invoked right after the job validation passes and only for the jobs
// enabling the downgrade.
func TestValidJobFnDegradeConstraintHook(t *testing.T) {
	for _, tt := range buildValidJobDowngradeHookTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			sJob := &SchedulerJob{}
			sJob.SchedulerJobAttr = util.SchedulerJobAttr{
				ComJob: util.ComJob{
					Name:       api.JobID("downgrade-hook-job"),
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
