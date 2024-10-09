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
Package test is using for HuaWei Ascend pin fault rescheduling.
*/
package test

import (
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/tools/record"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/cache"
	"volcano.sh/volcano/pkg/scheduler/framework"
	"volcano.sh/volcano/pkg/scheduler/util"
)

// FakeSSNReSchedule fake normal test ssn 2jobs.
func FakeSSNReSchedule() *framework.Session {
	binder := &util.FakeBinder{
		Binds:   map[string]string{},
		Channel: make(chan string),
	}
	schedulerCache := &cache.SchedulerCache{
		Nodes:         make(map[string]*api.NodeInfo),
		Jobs:          make(map[api.JobID]*api.JobInfo),
		Queues:        make(map[api.QueueID]*api.QueueInfo),
		Binder:        binder,
		StatusUpdater: &util.FakeStatusUpdater{},
		VolumeBinder:  &util.FakeVolumeBinder{},

		Recorder: record.NewFakeRecorder(NPUIndex4),
	}
	nodes := FakeNormalTestNodes(NPUIndex4)
	for _, node := range nodes {
		schedulerCache.AddNode(node.Node)
	}
	jobInf0 := FakeNormalTestJob("job0", npuIndex2)
	jobInf1 := FakeNormalTestJob("job1", npuIndex2)
	for _, task := range jobInf0.Tasks {
		schedulerCache.AddPod(task.Pod)
	}
	for _, task := range jobInf1.Tasks {
		schedulerCache.AddPod(task.Pod)
	}
	AddTestJobPodGroup(jobInf0)
	AddTestJobPodGroup(jobInf1)

	snapshot := schedulerCache.Snapshot()
	ssn := &framework.Session{
		UID:            uuid.NewUUID(),
		Jobs:           snapshot.Jobs,
		Nodes:          snapshot.Nodes,
		RevocableNodes: snapshot.RevocableNodes,
		Queues:         snapshot.Queues,
		NamespaceInfo:  snapshot.NamespaceInfo,
	}
	modifyJobInf(jobInf0, jobInf1)
	AddJobIntoFakeSSN(ssn, jobInf0)
	AddJobIntoFakeSSN(ssn, jobInf1)
	return ssn
}

func modifyJobInf(jobInf0 *api.JobInfo, jobInf1 *api.JobInfo) {
	taskName0 := FakeTaskName0
	taskName1 := FakeTaskName1
	annoCards := "Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7"
	jobInf0.PodGroup.Labels = make(map[string]string, npuIndex3)
	jobInf1.PodGroup.Labels = make(map[string]string, npuIndex3)
	jobInf0.PodGroup.Labels["fault-scheduling"] = "grace"
	jobInf1.PodGroup.Labels["fault-scheduling"] = "grace"
	jobInf0.Tasks[api.TaskID(taskName0)].Pod.Annotations = map[string]string{podRankIndex: "0",
		NPU910CardName: annoCards, AscendNPUPodRealUse: annoCards}
	jobInf0.Tasks[api.TaskID(taskName1)].Pod.Annotations = map[string]string{podRankIndex: "1",
		NPU910CardName: annoCards, AscendNPUPodRealUse: annoCards}
	jobInf1.Tasks[api.TaskID(taskName0)].Pod.Annotations =
		map[string]string{podRankIndex: "2", NPU910CardName: annoCards, AscendNPUPodRealUse: annoCards}
	jobInf1.Tasks[api.TaskID(taskName1)].Pod.Annotations =
		map[string]string{podRankIndex: "3", NPU910CardName: annoCards, AscendNPUPodRealUse: annoCards}
	jobInf0.Tasks[api.TaskID(taskName0)].NodeName = "node0"
	jobInf0.Tasks[api.TaskID(taskName1)].NodeName = "node1"
	jobInf1.Tasks[api.TaskID(taskName0)].NodeName = "node2"
	jobInf1.Tasks[api.TaskID(taskName1)].NodeName = "node3"
}
