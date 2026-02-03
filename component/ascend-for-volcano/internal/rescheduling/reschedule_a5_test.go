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
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/apis/pkg/apis/scheduling"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

type SinglePodReschedulingUpgradeFor910A5TestCase struct {
	name        string
	jobInfo     *api.JobInfo
	fJob        *FaultJob
	reScheduler *ReScheduler
	wantPending int
	wantBackup  int
	wantDelete  bool
}

func buildSinglePodReschedulingUpgradeFor910A5Test1() SinglePodReschedulingUpgradeFor910A5TestCase {
	return SinglePodReschedulingUpgradeFor910A5TestCase{
		name: "Label not enable, should return early",
		jobInfo: &api.JobInfo{
			PodGroup: &api.PodGroup{
				PodGroup: scheduling.PodGroup{
					ObjectMeta: v1.ObjectMeta{
						Labels: map[string]string{util.SinglePodTag: "off"},
					},
				},
			},
		},
		fJob:        &FaultJob{PendingSessionNum: 1},
		reScheduler: &ReScheduler{},
		wantPending: 1,
	}
}

func buildSinglePodReschedulingUpgradeFor910A5Test2() SinglePodReschedulingUpgradeFor910A5TestCase {
	return SinglePodReschedulingUpgradeFor910A5TestCase{
		name: "SuperPodAnnoKey present, PendingSessionNum == spPendingTimes",
		jobInfo: &api.JobInfo{
			PodGroup: &api.PodGroup{
				PodGroup: scheduling.PodGroup{
					ObjectMeta: v1.ObjectMeta{
						Labels:      map[string]string{util.SinglePodTag: util.EnableFunc},
						Annotations: map[string]string{util.SuperPodAnnoKey: "8"},
					},
				},
			},
		},
		fJob: &FaultJob{
			PendingSessionNum: spPendingTimes,
			JobUID:            "job2",
		},
		reScheduler: &ReScheduler{
			Jobs: map[api.JobID]plugin.SchedulerJob{
				"job2": {},
			},
		},
		wantPending: spPendingTimes + 1,
		wantDelete:  false,
	}
}
func buildSinglePodReschedulingUpgradeFor910A5Test3() SinglePodReschedulingUpgradeFor910A5TestCase {
	return SinglePodReschedulingUpgradeFor910A5TestCase{
		name: "PendingSessionNum == pendingTimes, DeleteExecutedFlag set to false",
		jobInfo: &api.JobInfo{
			PodGroup: &api.PodGroup{
				PodGroup: scheduling.PodGroup{
					ObjectMeta: v1.ObjectMeta{
						Labels: map[string]string{util.SinglePodTag: util.EnableFunc},
					},
				},
			},
		},
		fJob: &FaultJob{
			PendingSessionNum: pendingTimes,
			JobUID:            "job3",
		},
		reScheduler: &ReScheduler{
			Jobs: map[api.JobID]plugin.SchedulerJob{
				"job3": {},
			},
		},
		wantPending: pendingTimes + 1,
		wantDelete:  false,
	}
}
func buildSinglePodReschedulingUpgradeFor910A5Test4() SinglePodReschedulingUpgradeFor910A5TestCase {
	return SinglePodReschedulingUpgradeFor910A5TestCase{
		name: "PendingSessionNum == tpPendingTimes, DeleteExecutedFlag set to false",
		jobInfo: &api.JobInfo{
			PodGroup: &api.PodGroup{
				PodGroup: scheduling.PodGroup{
					ObjectMeta: v1.ObjectMeta{
						Labels: map[string]string{util.SinglePodTag: util.EnableFunc},
					},
				},
			},
		},
		fJob: &FaultJob{
			PendingSessionNum: tpPendingTimes,
			JobUID:            "job4",
		},
		reScheduler: &ReScheduler{
			Jobs: map[api.JobID]plugin.SchedulerJob{
				"job4": {A5Fields: plugin.A5Fields{}},
			},
		},
		wantPending: tpPendingTimes + 1,
		wantDelete:  false,
	}
}

func buildSinglePodReschedulingUpgradeFor910A5Test5() SinglePodReschedulingUpgradeFor910A5TestCase {
	return SinglePodReschedulingUpgradeFor910A5TestCase{
		name: "Acjob upgrades to job rescheduling when set recover_policy_path != pod",
		jobInfo: &api.JobInfo{
			PodGroup: &api.PodGroup{
				PodGroup: scheduling.PodGroup{
					ObjectMeta: v1.ObjectMeta{
						Labels: map[string]string{util.SinglePodTag: "on"},
					},
				},
			},
		},
		fJob: &FaultJob{
			PendingSessionNum: 0,
			ReScheduleLimit:   "",
		},
		reScheduler: &ReScheduler{},
		wantPending: 1,
	}
}

func buildSinglePodReschedulingUpgradeFor910A5Test6() SinglePodReschedulingUpgradeFor910A5TestCase {
	return SinglePodReschedulingUpgradeFor910A5TestCase{
		name: "Deployment and statefulset upgrades to job rescheduling when set recover_policy_path != pod",
		jobInfo: &api.JobInfo{
			PodGroup: &api.PodGroup{
				PodGroup: scheduling.PodGroup{
					ObjectMeta: v1.ObjectMeta{
						Labels: map[string]string{util.SinglePodTag: ""},
					},
				},
			},
		},
		fJob: &FaultJob{
			PendingSessionNum: 0,
			ReScheduleLimit:   "",
			Labels:            map[string]string{util.SinglePodTag: "on"},
		},
		reScheduler: &ReScheduler{},
		wantPending: 1,
	}
}

func TestSinglePodReschedulingUpgradeFor910A5(t *testing.T) {
	tests := []SinglePodReschedulingUpgradeFor910A5TestCase{
		buildSinglePodReschedulingUpgradeFor910A5Test1(), buildSinglePodReschedulingUpgradeFor910A5Test2(),
		buildSinglePodReschedulingUpgradeFor910A5Test3(), buildSinglePodReschedulingUpgradeFor910A5Test4(),
		buildSinglePodReschedulingUpgradeFor910A5Test5(), buildSinglePodReschedulingUpgradeFor910A5Test6(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.reScheduler.singlePodReschedulingUpgradeFor910A5(tt.jobInfo, tt.fJob)
			if tt.wantPending != 0 && tt.fJob.PendingSessionNum != tt.wantPending {
				t.Errorf("PendingSessionNum = %d, want %d", tt.fJob.PendingSessionNum, tt.wantPending)
			}
			if tt.wantDelete != tt.fJob.DeleteExecutedFlag && tt.wantDelete != false {
				t.Errorf("DeleteExecutedFlag = %v, want %v", tt.fJob.DeleteExecutedFlag, tt.wantDelete)
			}
		})
	}
}
