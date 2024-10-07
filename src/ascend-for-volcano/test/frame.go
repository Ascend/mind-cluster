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
Package test is using for HuaWei Ascend pin scheduling test.
*/
package test

import (
	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/tools/record"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/cache"
	"volcano.sh/volcano/pkg/scheduler/conf"
	"volcano.sh/volcano/pkg/scheduler/framework"
	"volcano.sh/volcano/pkg/scheduler/util"
)

// AddResource add resource into resourceList
func AddResource(resourceList v1.ResourceList, name v1.ResourceName, need string) {
	resourceList[name] = resource.MustParse(need)
}

// AddJobIntoFakeSSN Add test job into fake SSN.
func AddJobIntoFakeSSN(ssn *framework.Session, info ...*api.JobInfo) {
	for _, testJob := range info {
		ssn.Jobs[testJob.UID] = testJob
	}
}

// AddConfigIntoFakeSSN Add test node into fake SSN.
func AddConfigIntoFakeSSN(ssn *framework.Session, configs []conf.Configuration) {
	ssn.Configurations = configs
}

// FakeNormalSSN fake normal test ssn.
func FakeNormalSSN() *framework.Session {
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

		Recorder: record.NewFakeRecorder(npuIndex3),
	}

	nodes := FakeNormalTestNodes(npuIndex3)
	for _, node := range nodes {
		schedulerCache.AddNode(node.Node)
	}
	jobInf := FakeNormalTestJob("pg1", npuIndex3)
	for _, task := range jobInf.Tasks {
		schedulerCache.AddPod(task.Pod)
	}
	AddTestJobPodGroup(jobInf)
	snapshot := schedulerCache.Snapshot()
	ssn := &framework.Session{
		UID:            uuid.NewUUID(),
		Jobs:           map[api.JobID]*api.JobInfo{jobInf.UID: jobInf},
		Nodes:          snapshot.Nodes,
		RevocableNodes: snapshot.RevocableNodes,
		Queues:         snapshot.Queues,
		NamespaceInfo:  snapshot.NamespaceInfo,
	}
	return ssn
}

// PatchReset go monkey patch reset
func PatchReset(patch *gomonkey.Patches) {
	if patch != nil {
		patch.Reset()
	}
}
