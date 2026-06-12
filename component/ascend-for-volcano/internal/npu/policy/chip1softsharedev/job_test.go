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

// Package chip1softsharedev is using for HuaWei chip1softsharedev schedule.
package chip1softsharedev

import (
	"reflect"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/ascend910/ascend910b"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/base"
)

const (
	testJobName         = "test-job"
	validAICoreQuota    = 50
	invalidAICoreQuota  = 101
	mismatchAICoreQuota = 40
	validHbmQuota       = 1024
	zeroHbmQuota        = 0
	validHbmQuota2      = 16384
	testReqNPUNum       = 100
	testNPUTaskNum      = 2
	invalidPolicy       = "invalid"
)

func TestGetSoftShareDevResource(t *testing.T) {
	tp := &chip1softsharedev{Base910b: ascend910b.Base910b{
		NPUHandler: base.NPUHandler{SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{}}}}}
	tp.Name = testJobName
	tests := []struct {
		name    string
		labels  map[string]string
		want    softShareDevResource
		wantErr bool
	}{
		{
			name:    "missing labels",
			labels:  map[string]string{},
			want:    softShareDevResource{},
			wantErr: true,
		},
		{
			name: "valid labels",
			labels: map[string]string{
				util.SchedulerSoftShareDevAicoreQuotaKey: "50",
				util.SchedulerSoftShareDevHbmQuotaKey:    "16384",
				util.SchedulerSoftShareDevPolicyKey:      util.SoftShareDevPolicyFixedShare,
			},
			want: softShareDevResource{
				aicoreQuota:      validAICoreQuota,
				hbmQuota:         validHbmQuota2,
				schedulingPolicy: util.SoftShareDevPolicyFixedShare,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp.Label = tt.labels
			got, err := tp.getSoftShareDevResource()
			if (err != nil) != tt.wantErr {
				t.Errorf("getSoftShareDevResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSoftShareDevResource() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

type checkoftShareDevResourceTestCase struct {
	name        string
	reqResource softShareDevResource
	wantErr     bool
}

func buildCheckSoftShareDevResourceTestCases() []checkoftShareDevResourceTestCase {
	return []checkoftShareDevResourceTestCase{
		{
			name: "valid resource",
			reqResource: softShareDevResource{
				aicoreQuota:      validAICoreQuota,
				hbmQuota:         validHbmQuota,
				schedulingPolicy: util.SoftShareDevPolicyFixedShare,
			},
			wantErr: false,
		},
		{
			name: "aicore quota out of range",
			reqResource: softShareDevResource{
				aicoreQuota:      invalidAICoreQuota,
				hbmQuota:         validHbmQuota,
				schedulingPolicy: util.SoftShareDevPolicyFixedShare,
			},
			wantErr: true,
		},
		{
			name: "hbm quota too small",
			reqResource: softShareDevResource{
				aicoreQuota:      validAICoreQuota,
				hbmQuota:         zeroHbmQuota,
				schedulingPolicy: util.SoftShareDevPolicyFixedShare,
			},
			wantErr: true,
		},
		{
			name: "invalid policy",
			reqResource: softShareDevResource{
				aicoreQuota:      validAICoreQuota,
				hbmQuota:         validHbmQuota,
				schedulingPolicy: invalidPolicy,
			},
			wantErr: true,
		},
	}
}

func TestCheckSoftShareDevResource(t *testing.T) {
	tp := &chip1softsharedev{Base910b: ascend910b.Base910b{
		NPUHandler: base.NPUHandler{SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{}}}}}
	tp.ReqNPUNum = testReqNPUNum
	tp.NPUTaskNum = testNPUTaskNum
	tests := buildCheckSoftShareDevResourceTestCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tp.checkSoftShareDevResource(tt.reqResource)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkSoftShareDevResource() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidNPUJob(t *testing.T) {
	tp := &chip1softsharedev{Base910b: ascend910b.Base910b{
		NPUHandler: base.NPUHandler{SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{}}}}}
	tp.ReqNPUNum = testReqNPUNum
	tp.NPUTaskNum = testNPUTaskNum
	tp.Label = map[string]string{
		util.SchedulerSoftShareDevAicoreQuotaKey: "50",
		util.SchedulerSoftShareDevHbmQuotaKey:    "1024",
		util.SchedulerSoftShareDevPolicyKey:      util.SoftShareDevPolicyFixedShare,
	}
	result := tp.validNPUJob()
	if result != nil {
		t.Errorf("validNPUJob() = %v, want nil", result)
	}
	tp.Label[util.SchedulerSoftShareDevAicoreQuotaKey] = "101"
	result = tp.validNPUJob()
	if result == nil || result.Pass {
		t.Error("validNPUJob() should return invalid result for aicore out of range")
	}
	var nilTp *chip1softsharedev
	result = nilTp.validNPUJob()
	if result == nil || result.Pass {
		t.Error("validNPUJob() should return invalid result for nil plugin")
	}
}
