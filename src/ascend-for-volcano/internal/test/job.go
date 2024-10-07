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
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// FakeSchedulerJobAttrByJob fake scheduler attr by job
func FakeSchedulerJobAttrByJob(job *api.JobInfo) util.SchedulerJobAttr {
	attr := util.SchedulerJobAttr{
		ComJob: util.ComJob{
			Name:      job.UID,
			NameSpace: job.Namespace,
			Selector:  nil,
			Label:     nil,
		},
	}
	name, num, err := plugin.GetVCJobReqNPUTypeFromJobInfo(job)
	if err != nil {
		return attr
	}
	NPUJob := &util.NPUJob{
		ReqNPUName: name,
		ReqNPUNum:  num,
		Tasks:      plugin.GetJobNPUTasks(job),
	}
	NPUJob.NPUTaskNum = NPUJob.GetNPUTaskNumInJob()
	attr.NPUJob = NPUJob
	return attr
}
