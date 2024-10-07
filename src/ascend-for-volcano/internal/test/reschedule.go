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
Package test is using for HuaWei Ascend testing.
*/
package test

import (
	"reflect"

	"github.com/agiledragon/gomonkey/v2"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/rescheduling"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
)

// PatchNew go monkey patch
func PatchNew() *gomonkey.Patches {
	return gomonkey.ApplyFunc(rescheduling.New, func(_ *plugin.ScheduleEnv,
		_ string) *rescheduling.ReScheduler {
		return &rescheduling.ReScheduler{GraceDeleteTime: rescheduling.DefaultGraceOverTime}
	})
}

// PatchNewComRes go monkey patch
func PatchNewComRes() *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(&rescheduling.ReScheduler{}), "NewCommonReScheduler",
		func(_ *rescheduling.ReScheduler, _ string) { return })
}

// PatchSynNode go monkey patch
func PatchSynNode() *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(&rescheduling.ReScheduler{}), "SynCacheFaultNodeWithSession",
		func(_ *rescheduling.ReScheduler, _ string) { return })
}

// PatchAddNode go monkey patch
func PatchAddNode() *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(&rescheduling.ReScheduler{}),
		"AddFaultNodeWithSession",
		func(_ *rescheduling.ReScheduler, _ string) { return })
}

// PatchSynJob go monkey patch
func PatchSynJob() *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(&rescheduling.ReScheduler{}),
		"SynCacheFaultJobWithSession",
		func(_ *rescheduling.ReScheduler, _ *framework.Session) { return })
}

// PatchForce go monkey patch
func PatchForce() *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(&rescheduling.ReScheduler{}),
		"SynCacheFaultJobWithSession",
		func(_ *rescheduling.ReScheduler, _ *framework.Session) { return })
}

// PatchGetRun go monkey patch
func PatchGetRun() *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(&rescheduling.ReScheduler{}),
		"GetRunningJobs",
		func(_ *rescheduling.ReScheduler, _ *framework.Session, _, _ string) (map[api.JobID]*api.JobInfo,
			error) {
			return map[api.JobID]*api.JobInfo{"job1": &api.JobInfo{}}, nil
		})
}

// PatchAddJob go monkey patch
func PatchAddJob() *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(&rescheduling.ReScheduler{}),
		"AddFaultJobWithSession",
		func(_ *rescheduling.ReScheduler, _ map[api.JobID]*api.JobInfo, _, _ string) error { return nil })
}

// PatchRestart go monkey patch
func PatchRestart() *gomonkey.Patches {
	return gomonkey.ApplyMethod(reflect.TypeOf(&rescheduling.ReScheduler{}),
		"RestartFaultJobs",
		func(_ *rescheduling.ReScheduler, _ *framework.Session) error { return nil })
}
