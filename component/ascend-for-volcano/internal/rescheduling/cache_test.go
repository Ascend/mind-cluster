/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package rescheduling is using for HuaWei Ascend pin fault rescheduling.
*/
package rescheduling

import (
	"encoding/json"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

func fakeInvalidReSchedulerCMData() *DealReSchedulerConfigmap {
	return &DealReSchedulerConfigmap{
		CMName:      CmName,
		CMNameSpace: CmNameSpace,
		CMData:      nil,
	}
}

func dealMarshal(data interface{}) string {
	dataString, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(dataString)
}

func fakeNormalReSchedulerCMData() *DealReSchedulerConfigmap {
	faultNodes := []FaultNode{
		*fakeTestFaultNodeNodeHealthy("node0"),
		*fakeTestFaultNodeNodeHealthy("node1"),
	}
	faultJobs := []FaultJob{
		*fakeTestFaultJob([]string{"node0", "node1"}, []string{"0", "9"},
			nil, "job1", "test"),
	}

	fNodeBuffer := dealMarshal(faultNodes)
	fJobBuffer := dealMarshal(faultJobs)

	return &DealReSchedulerConfigmap{
		CMName:      CmName,
		CMNameSpace: CmNameSpace,
		CMData: map[string]string{
			CmFaultNodeKind:     string(fNodeBuffer),
			CmFaultJob910x8Kind: string(fJobBuffer),
		},
	}
}

type DealReSchedulerCacheSetFaultNodesFromCMTests struct {
	fields  *DealReSchedulerCache
	name    string
	wantErr bool
}

func buildDealReSchedulerCacheSetFaultNodesFromCMTests() []DealReSchedulerCacheSetFaultNodesFromCMTests {
	field1 := fakeReSchedulerCache()
	field1.DealReSchedulerConfigmap = fakeInvalidReSchedulerCMData()
	field2 := fakeReSchedulerCache()
	field2.DealReSchedulerConfigmap = fakeNormalReSchedulerCMData()

	test1 := DealReSchedulerCacheSetFaultNodesFromCMTests{
		name:    "01-DealReSchedulerCache_SetFaultNodesFromCM()  invalid cache structure",
		fields:  field1,
		wantErr: true,
	}
	test2 := DealReSchedulerCacheSetFaultNodesFromCMTests{
		name:    "01-DealReSchedulerCache_SetFaultNodesFromCM()  succeed",
		fields:  field2,
		wantErr: false,
	}
	testCases := []DealReSchedulerCacheSetFaultNodesFromCMTests{
		test1,
		test2,
	}
	return testCases
}

// TestDealReSchedulerCacheSetFaultNodesFromCM test for set FaultNodes struct from configmap
func TestDealReSchedulerCacheSetFaultNodesFromCM(t *testing.T) {
	tests := buildDealReSchedulerCacheSetFaultNodesFromCMTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reCache := &DealReSchedulerCache{
				FaultNodes:               tt.fields.FaultNodes,
				FaultJobs:                tt.fields.FaultJobs,
				DealReSchedulerConfigmap: tt.fields.DealReSchedulerConfigmap,
			}
			if err := reCache.SetFaultNodesFromCM(); (err != nil) != tt.wantErr {
				t.Errorf("SetFaultNodesFromCM() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type ReSchedulerCacheWriteReSchedulerCacheToEnvCacheFields struct {
	DealReSchedulerConfigmap   *DealReSchedulerConfigmap
	FaultNodes                 []FaultNode
	FaultJobs                  []FaultJob
	NodeHeartbeats             []NodeHeartbeat
	AllocNodeRankOccurrenceMap map[api.JobID][]*AllocNodeRankOccurrence
}

type ReSchedulerCacheWriteReSchedulerCacheToEnvCacheArgs struct {
	env     *plugin.ScheduleEnv
	jobType string
}

type ReSchedulerCacheWriteReSchedulerCacheToEnvCacheTests struct {
	name    string
	fields  ReSchedulerCacheWriteReSchedulerCacheToEnvCacheFields
	args    ReSchedulerCacheWriteReSchedulerCacheToEnvCacheArgs
	wantErr bool
}

func buildReSchedulerCacheWriteReSchedulerCacheToEnvCache() []ReSchedulerCacheWriteReSchedulerCacheToEnvCacheTests {
	test1 := ReSchedulerCacheWriteReSchedulerCacheToEnvCacheTests{
		name: "01-ReSchedulerCache_WriteReSchedulerCacheToEnvCache()-nothing to write",
		fields: ReSchedulerCacheWriteReSchedulerCacheToEnvCacheFields{
			DealReSchedulerConfigmap:   nil,
			FaultNodes:                 []FaultNode{},
			FaultJobs:                  []FaultJob{},
			NodeHeartbeats:             []NodeHeartbeat{},
			AllocNodeRankOccurrenceMap: map[api.JobID][]*AllocNodeRankOccurrence{},
		},
		args: ReSchedulerCacheWriteReSchedulerCacheToEnvCacheArgs{
			env: &plugin.ScheduleEnv{
				Cache: plugin.ScheduleCache{
					Names:           map[string]string{RePropertyName: CmName},
					Namespaces:      map[string]string{RePropertyName: CmNameSpace},
					FaultConfigMaps: map[api.JobID]*plugin.FaultRankIdData{},
					Data:            map[string]map[string]string{RePropertyName: make(map[string]string, util.MapInitNum)},
				},
			},
			jobType: CmFaultJob910x8Kind,
		},
		wantErr: false,
	}
	faultJob := fakeTestFaultJob([]string{"node0"}, []string{"0", "1"}, nil, "job0", "vcjob")
	test2 := ReSchedulerCacheWriteReSchedulerCacheToEnvCacheTests{
		name: "02-ReSchedulerCache_WriteReSchedulerCacheToEnvCache()-with faultJob",
		fields: ReSchedulerCacheWriteReSchedulerCacheToEnvCacheFields{
			DealReSchedulerConfigmap:   nil,
			FaultNodes:                 []FaultNode{},
			FaultJobs:                  []FaultJob{*faultJob},
			NodeHeartbeats:             []NodeHeartbeat{},
			AllocNodeRankOccurrenceMap: map[api.JobID][]*AllocNodeRankOccurrence{},
		},
		args: ReSchedulerCacheWriteReSchedulerCacheToEnvCacheArgs{
			env: &plugin.ScheduleEnv{
				Cache: plugin.ScheduleCache{
					Names:           map[string]string{RePropertyName: CmName},
					Namespaces:      map[string]string{RePropertyName: CmNameSpace},
					FaultConfigMaps: map[api.JobID]*plugin.FaultRankIdData{},
					Data: map[string]map[string]string{RePropertyName: make(map[string]string, util.MapInitNum),
						JobRecovery: make(map[string]string, util.MapInitNum)},
				},
			},
			jobType: CmFaultJob910x8Kind,
		},
		wantErr: false,
	}
	tests := []ReSchedulerCacheWriteReSchedulerCacheToEnvCacheTests{test1, test2}
	return tests
}

// TestDealReSchedulerCacheWriteReSchedulerCacheToEnvCache test for re-scheduler writing
func TestDealReSchedulerCacheWriteReSchedulerCacheToEnvCache(t *testing.T) {
	tests := buildReSchedulerCacheWriteReSchedulerCacheToEnvCache()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reCache := &DealReSchedulerCache{
				DealReSchedulerConfigmap:   tt.fields.DealReSchedulerConfigmap,
				FaultNodes:                 tt.fields.FaultNodes,
				FaultJobs:                  tt.fields.FaultJobs,
				NodeHeartbeats:             tt.fields.NodeHeartbeats,
				AllocNodeRankOccurrenceMap: tt.fields.AllocNodeRankOccurrenceMap,
			}
			if err := reCache.WriteReSchedulerCacheToEnvCache(
				tt.args.env, tt.args.jobType); (err != nil) != tt.wantErr {
				t.Errorf("WriteReSchedulerCacheToEnvCache() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
