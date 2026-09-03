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
Package plugin is using for HuaWei Ascend pin affinity schedule frame.
*/
package plugin

import (
	"reflect"
	"testing"
	"time"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/affinity/chip/topo"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/k8s"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

type nodeFields struct {
	Name       string
	Capability map[v1.ResourceName]float64
	Allocate   map[v1.ResourceName]float64
	Idle       map[v1.ResourceName]float64
	Annotation map[string]string
	Label      map[string]string
}

type checkNPUResourceStableArgs struct {
	vcJob SchedulerJob
}

type checkNPUResourceStableTest struct {
	name    string
	fields  nodeFields
	args    checkNPUResourceStableArgs
	wantErr bool
}

func buildVCheckNPUResourceStableTest() []checkNPUResourceStableTest {
	tJob := SchedulerJob{policyHandler: New(testPluginName),
		SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{ReqNPUName: util.NPU310PCardName}}}
	vJob := SchedulerJob{policyHandler: New(testPluginName),
		SchedulerJobAttr: util.SchedulerJobAttr{NPUJob: &util.NPUJob{ReqNPUName: util.AscendNPUCore}}}
	tests := []checkNPUResourceStableTest{
		{
			name:    "01-checkNPUResourceStable no annotation test",
			fields:  nodeFields{Name: "haha", Idle: map[v1.ResourceName]float64{testCardName: 1}, Annotation: nil},
			args:    checkNPUResourceStableArgs{vcJob: tJob},
			wantErr: true,
		},
		{
			name: "02-checkNPUResourceStable ok test.",
			fields: nodeFields{Name: "haha", Idle: map[v1.ResourceName]float64{util.NPU310PCardName: util.NPUHexKilo},
				Annotation: map[string]string{util.NPU310PCardName: "Ascend310P-0"},
				Capability: map[v1.ResourceName]float64{util.NPU310PCardName: util.NPUHexKilo},
			},
			args:    checkNPUResourceStableArgs{vcJob: tJob},
			wantErr: false,
		},
		{
			name: "03-checkNPUResourceStable vNPU ok test.",
			fields: nodeFields{Name: "haha", Idle: map[v1.ResourceName]float64{testCardName: 1},
				Annotation: map[string]string{testCardName: "haha"}},
			args:    checkNPUResourceStableArgs{vcJob: vJob},
			wantErr: false,
		},
	}
	return tests
}

func TestCheckNPUResourceStable(t *testing.T) {
	tests := buildVCheckNPUResourceStableTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NPUNode{
				CommonNode: CommonNode{
					Name:       tt.fields.Name,
					Capability: tt.fields.Capability,
					Allocate:   tt.fields.Allocate,
					Idle:       tt.fields.Idle,
					Annotation: tt.fields.Annotation,
					Label:      tt.fields.Label,
				},
			}
			if err := n.checkNPUResourceStable(tt.args.vcJob); (err != nil) != tt.wantErr {
				t.Errorf("checkNPUResourceStable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type nodePredicateArgs struct {
	taskInfo *api.TaskInfo
	nodeInfo *api.NodeInfo
}

type nodePredicateTest struct {
	name    string
	fields  fields
	args    nodePredicateArgs
	wantErr bool
}

func buildNodePredicateTest() []nodePredicateTest {
	tTasks := test.FakeNormalTestTasks(1)
	tNode := test.FakeNormalTestNode("haha")
	tests := []nodePredicateTest{
		{
			name:    "01-NodePredicate nil test.",
			fields:  fields{},
			args:    nodePredicateArgs{taskInfo: &api.TaskInfo{}, nodeInfo: nil},
			wantErr: true,
		},
		{
			name: "02-NodePredicate job not in test.",
			fields: fields{ScheduleEnv: ScheduleEnv{
				ClusterCache: ClusterCache{
					Jobs: map[api.JobID]SchedulerJob{"haha": {}},
				}}},
			args:    nodePredicateArgs{taskInfo: tTasks[0], nodeInfo: tNode},
			wantErr: false,
		},
		{
			name: "03-NodePredicate node not in test.",
			fields: fields{ScheduleEnv: ScheduleEnv{
				ClusterCache: ClusterCache{
					Jobs:  map[api.JobID]SchedulerJob{tTasks[0].Job: {policyHandler: New(PluginName)}},
					Nodes: map[string]NPUNode{"lala": {}}}}},
			args:    nodePredicateArgs{taskInfo: tTasks[0], nodeInfo: tNode},
			wantErr: false,
		},
		{
			name: "04-NodePredicate node not in test.",
			fields: fields{ScheduleEnv: ScheduleEnv{
				ClusterCache: ClusterCache{
					Jobs:  map[api.JobID]SchedulerJob{tTasks[0].Job: {policyHandler: New(PluginName)}},
					Nodes: map[string]NPUNode{"haha": {}}}}},
			args:    nodePredicateArgs{taskInfo: tTasks[0], nodeInfo: tNode},
			wantErr: true,
		},
		{
			name: "05-NodePredicate ok test.",
			fields: fields{ScheduleEnv: ScheduleEnv{
				ClusterCache: ClusterCache{
					Jobs:  map[api.JobID]SchedulerJob{tTasks[0].Job: {policyHandler: New(PluginName)}},
					Nodes: map[string]NPUNode{"haha": {}}}}},
			args:    nodePredicateArgs{taskInfo: tTasks[0], nodeInfo: tNode},
			wantErr: true,
		},
		{
			name: "06-NodePredicate UnHealthy Node test.",
			fields: fields{ScheduleEnv: ScheduleEnv{
				ClusterCache: ClusterCache{
					Jobs: map[api.JobID]SchedulerJob{tTasks[0].Job: {policyHandler: New(PluginName)}},
					Nodes: map[string]NPUNode{"haha": {
						CommonNode: CommonNode{
							Label:      map[string]string{util.NodeHealthyStatusKey: util.PreSeparateFaultCode},
							Annotation: map[string]string{util.NodeHealthyStatusKey: util.PreSeparateFaultCode},
						}}}}}},
			args:    nodePredicateArgs{taskInfo: tTasks[0], nodeInfo: tNode},
			wantErr: true,
		},
	}
	return tests
}

func TestSNodePredicate(t *testing.T) {
	tests := buildNodePredicateTest()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sHandle := &ScheduleHandler{
				NPUPlugins:  tt.fields.NPUPlugins,
				ScheduleEnv: tt.fields.ScheduleEnv,
			}
			tmpJob := sHandle.ScheduleEnv.Jobs["vcjob/pg0"]
			tmpJob.NPUJob = &util.NPUJob{}
			tmpJob.ReqNPUName = util.NPU910CardName
			if len(sHandle.ScheduleEnv.Jobs) != 0 {
				sHandle.ScheduleEnv.Jobs["vcjob/pg0"] = tmpJob
			}
			tt.args.taskInfo.Resreq = &api.Resource{}
			tt.args.taskInfo.Resreq.ScalarResources = make(map[v1.ResourceName]float64)
			tt.args.taskInfo.Resreq.ScalarResources[util.Ascend910bName] = util.NPUIndex10
			if err := sHandle.NodePredicate(tt.args.taskInfo, tt.args.nodeInfo); (err != nil) != tt.wantErr {
				t.Errorf("NodePredicate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type nPUNodeGetNewNPUNodeAnnotationTest struct {
	name            string
	usedTop         []int
	resourceName    string
	resourceNamePre string
	npuNode         *NPUNode
	want            string
	wantErr         bool
}

func buildNPUNodeGetNewNPUNodeAnnotationTest() []nPUNodeGetNewNPUNodeAnnotationTest {
	return []nPUNodeGetNewNPUNodeAnnotationTest{
		{
			name:    "01-GetNewNPUNodeAnnotation return error when npuNode is nil",
			npuNode: nil,
			wantErr: true,
		},
		{
			name:            "02-GetNewNPUNodeAnnotation return error when npuNode annotation is empty",
			npuNode:         &NPUNode{},
			usedTop:         []int{0},
			resourceName:    util.Ascend910,
			resourceNamePre: util.NPU910CardNamePre,
			wantErr:         true,
		},
		{
			name: "03-GetNewNPUNodeAnnotation return empty when npuNode annotation is empty",
			npuNode: &NPUNode{CommonNode: CommonNode{
				Annotation: map[string]string{util.Ascend910: ""}}},
			usedTop:         []int{0},
			resourceName:    util.Ascend910,
			resourceNamePre: util.NPU910CardNamePre,
			want:            "",
			wantErr:         false,
		},
		{
			name: "04-GetNewNPUNodeAnnotation return error when string to int error",
			npuNode: &NPUNode{CommonNode: CommonNode{
				Annotation: map[string]string{util.Ascend910: "Ascend910-s"}}},
			usedTop:         []int{0},
			resourceName:    util.Ascend910,
			resourceNamePre: util.NPU910CardNamePre,
			want:            "",
			wantErr:         true,
		},
		{
			name: "05-GetNewNPUNodeAnnotation return Ascend910-1 when get npu node annotation",
			npuNode: &NPUNode{CommonNode: CommonNode{
				Annotation: map[string]string{util.Ascend910: "Ascend910-0,Ascend910-1"}}},
			usedTop:         []int{0},
			resourceName:    util.Ascend910,
			resourceNamePre: util.NPU910CardNamePre,
			want:            "Ascend910-1",
			wantErr:         false,
		},
	}
}

type updateNPUNodeDeviceInfosTest struct {
	name string
	node NPUNode
	data k8s.NodeDeviceInfoWithID
}

func buildUpdateNPUNodeDeviceInfosTest() []updateNPUNodeDeviceInfosTest {
	return []updateNPUNodeDeviceInfosTest{
		{
			name: "01 force update by device info cm",
			node: NPUNode{CommonNode: CommonNode{devInfoUpdateTime: 0, Annotation: map[string]string{}}},
			data: k8s.NodeDeviceInfoWithID{NodeDeviceInfo: k8s.NodeDeviceInfo{UpdateTime: time.Now().Unix(),
				DeviceList: k8s.FakeDeviceList(),
			}},
		},
		{
			name: "02 update by device info cm and volcano cache",
			node: NPUNode{CommonNode: CommonNode{
				Idle:              map[v1.ResourceName]float64{util.NPU910CardName: util.NPUHexKilo},
				devInfoUpdateTime: time.Now().Unix(),
				Annotation:        map[string]string{util.NPU910CardName: "Ascend910-0"}}},
			data: k8s.NodeDeviceInfoWithID{NodeDeviceInfo: k8s.NodeDeviceInfo{UpdateTime: time.Now().Unix() +
				util.NPUIndex3, DeviceList: k8s.FakeDeviceList(),
			}},
		},
	}
}

func TestSyncAnnotation(t *testing.T) {
	t.Run("test syncAnnotation, node unhealth", func(t *testing.T) {
		cNode := NPUNode{CommonNode: CommonNode{}}
		nodeNew := &api.NodeInfo{
			Node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						util.NodeHealthyStatusKey: util.NodeUnHealthy,
					},
				},
			},
		}
		nodeInfo := k8s.NodeDNodeInfo{
			FaultDevList: []k8s.FaultDevList{
				{
					FaultCode:  []string{"test_code"},
					FaultLevel: "test_level",
				},
			},
		}
		switchInfo := k8s.SwitchFaultInfo{
			FaultCode:  []string{"test_code"},
			FaultLevel: "test_level",
		}
		cNode.syncAnnotation(nodeNew, nodeInfo, switchInfo)
		if len(nodeNew.Node.Annotations)+1 != len(cNode.Annotation) {
			t.Errorf("syncAnnotation is not equal")
		}
		if len(cNode.SwitchFaultCode) != len(switchInfo.FaultCode) || cNode.SwitchFaultLevel != switchInfo.FaultLevel {
			t.Errorf("syncswitchInfo is not equal")
		}
		if len(cNode.NodeFaultList) != len(nodeInfo.FaultDevList) {
			t.Errorf("syncnodeInfo is not equal")
		}
	})
}

func TestUpdateNPUNodeDeviceInfos(t *testing.T) {
	for _, tt := range buildUpdateNPUNodeDeviceInfosTest() {
		t.Run(tt.name, func(t *testing.T) {
			tt.node.updateNPUNodeDeviceInfos(tt.data)
		})
	}
}

type getNeedInitNodeListTest struct {
	name    string
	ssn     *framework.Session
	sHandle *ScheduleHandler
	want    []*api.NodeInfo
}

func buildGetNeedInitNodeListTest() []getNeedInitNodeListTest {
	// Prepare v1.Node objects for testing
	readyNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node-ready"},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{Type: v1.NodeReady, Status: v1.ConditionTrue},
			},
		},
	}
	notReadyNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node-notready"},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{Type: v1.NodeReady, Status: v1.ConditionFalse},
			},
		},
	}
	readyNode2 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node-ready2"},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{Type: v1.NodeReady, Status: v1.ConditionTrue},
			},
		},
	}

	// Build informer with pre-populated indexer for tests that need kubeClient
	fakeClient := fake.NewSimpleClientset()
	informerFactory := informers.NewSharedInformerFactory(fakeClient, 0)
	indexer := informerFactory.Core().V1().Nodes().Informer().GetIndexer()
	indexer.Add(readyNode)
	indexer.Add(notReadyNode)
	// node-ready2 is deliberately NOT added to the indexer (simulates deleted node)

	readyNodeForSsn := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node-both"},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{Type: v1.NodeReady, Status: v1.ConditionTrue},
			},
		},
	}
	notReadyNodeForSsn := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node-ssn-notready"},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{Type: v1.NodeReady, Status: v1.ConditionFalse},
			},
		},
	}

	// Session with mixed nodes (for mixed scenario test)
	ssnWithMixed := &framework.Session{
		NodeList: []*api.NodeInfo{
			api.NewNodeInfo(readyNodeForSsn),    // ready → included
			api.NewNodeInfo(notReadyNodeForSsn), // not ready → skipped
		},
		Nodes: map[string]*api.NodeInfo{
			"node-both":         api.NewNodeInfo(readyNodeForSsn),
			"node-ssn-notready": api.NewNodeInfo(notReadyNodeForSsn),
		},
	}

	return []getNeedInitNodeListTest{
		{
			name:    "01-sHandle is nil, return ssn.NodeList(nil)",
			ssn:     &framework.Session{},
			sHandle: nil,
			want:    nil,
		},
		{
			name: "02-KubeClient is nil, return ssn.NodeList directly",
			ssn: &framework.Session{
				NodeList: []*api.NodeInfo{api.NewNodeInfo(readyNode)},
			},
			sHandle: &ScheduleHandler{},
			want:    []*api.NodeInfo{api.NewNodeInfo(readyNode)},
		},
		{
			name: "03-KubeClient set, empty Nodes and empty NodeList, return empty",
			ssn:  &framework.Session{},
			sHandle: func() *ScheduleHandler {
				sh := &ScheduleHandler{}
				sh.FrameAttr.KubeClient = fakeClient
				sh.FrameAttr.informerFactory = informerFactory
				sh.Nodes = map[string]NPUNode{}
				return sh
			}(),
			want: []*api.NodeInfo{},
		},
		{
			name: "04-node deleted from cluster, not in informer indexer, skipped",
			ssn:  &framework.Session{},
			sHandle: func() *ScheduleHandler {
				sh := &ScheduleHandler{}
				sh.FrameAttr.KubeClient = fakeClient
				sh.FrameAttr.informerFactory = informerFactory
				sh.Nodes = map[string]NPUNode{"node-deleted": {}}
				return sh
			}(),
			want: []*api.NodeInfo{},
		},
		{
			name: "05-node not in ssn, found in informer and ready, recovered from informer",
			ssn:  &framework.Session{},
			sHandle: func() *ScheduleHandler {
				sh := &ScheduleHandler{}
				sh.FrameAttr.KubeClient = fakeClient
				sh.FrameAttr.informerFactory = informerFactory
				sh.Nodes = map[string]NPUNode{"node-ready": {}}
				return sh
			}(),
			want: []*api.NodeInfo{api.NewNodeInfo(readyNode)},
		},
		{
			name: "06-node not in ssn, found in informer but NotReady, skipped",
			ssn:  &framework.Session{},
			sHandle: func() *ScheduleHandler {
				sh := &ScheduleHandler{}
				sh.FrameAttr.KubeClient = fakeClient
				sh.FrameAttr.informerFactory = informerFactory
				sh.Nodes = map[string]NPUNode{"node-notready": {}}
				return sh
			}(),
			want: []*api.NodeInfo{},
		},
		{
			name: "07-node not in ssn and not in informer, skipped as deleted",
			ssn:  &framework.Session{},
			sHandle: func() *ScheduleHandler {
				sh := &ScheduleHandler{}
				sh.FrameAttr.KubeClient = fakeClient
				sh.FrameAttr.informerFactory = informerFactory
				sh.Nodes = map[string]NPUNode{"node-deleted": {}}
				return sh
			}(),
			want: []*api.NodeInfo{},
		},
		{
			name: "08-node in ssn.NodeList and ready, included",
			ssn: &framework.Session{
				NodeList: []*api.NodeInfo{api.NewNodeInfo(readyNode2)},
			},
			sHandle: func() *ScheduleHandler {
				sh := &ScheduleHandler{}
				sh.FrameAttr.KubeClient = fakeClient
				sh.FrameAttr.informerFactory = informerFactory
				sh.Nodes = map[string]NPUNode{}
				return sh
			}(),
			want: []*api.NodeInfo{api.NewNodeInfo(readyNode2)},
		},
		{
			name: "09-node in ssn.NodeList but NotReady, skipped",
			ssn: &framework.Session{
				NodeList: []*api.NodeInfo{api.NewNodeInfo(notReadyNodeForSsn)},
			},
			sHandle: func() *ScheduleHandler {
				sh := &ScheduleHandler{}
				sh.FrameAttr.KubeClient = fakeClient
				sh.FrameAttr.informerFactory = informerFactory
				sh.Nodes = map[string]NPUNode{}
				return sh
			}(),
			want: []*api.NodeInfo{},
		},
		{
			name: "10-node in both sHandle.Nodes and ssn, present in ssn.NodeList and ready, included by second loop",
			ssn: &framework.Session{
				NodeList: []*api.NodeInfo{api.NewNodeInfo(readyNodeForSsn)},
				Nodes:    map[string]*api.NodeInfo{"node-both": api.NewNodeInfo(readyNodeForSsn)},
			},
			sHandle: func() *ScheduleHandler {
				sh := &ScheduleHandler{}
				sh.FrameAttr.KubeClient = fakeClient
				sh.FrameAttr.informerFactory = informerFactory
				sh.Nodes = map[string]NPUNode{"node-both": {}}
				return sh
			}(),
			want: []*api.NodeInfo{api.NewNodeInfo(readyNodeForSsn)},
		},
		{
			name: "11-mixed: nodes recovered from informer + ready nodes from ssn.NodeList",
			ssn:  ssnWithMixed,
			sHandle: func() *ScheduleHandler {
				sh := &ScheduleHandler{}
				sh.FrameAttr.KubeClient = fakeClient
				sh.FrameAttr.informerFactory = informerFactory
				sh.Nodes = map[string]NPUNode{
					"node-ready":    {}, // ready, in informer, NOT in ssn → recovered
					"node-notready": {}, // not ready → skipped
					"node-deleted":  {}, // not in informer → skipped
					"node-both":     {}, // in ssn → handled by second loop
				}
				return sh
			}(),
			want: []*api.NodeInfo{
				api.NewNodeInfo(readyNode),       // node-ready from informer
				api.NewNodeInfo(readyNodeForSsn), // node-both from ssn.NodeList
			},
		},
		{
			name: "12-node in ssn.Nodes but NOT in ssn.NodeList, skipped by both loops",
			ssn: &framework.Session{
				NodeList: nil,
				Nodes:    map[string]*api.NodeInfo{"node-both": api.NewNodeInfo(readyNodeForSsn)},
			},
			sHandle: func() *ScheduleHandler {
				sh := &ScheduleHandler{}
				sh.FrameAttr.KubeClient = fakeClient
				sh.FrameAttr.informerFactory = informerFactory
				sh.Nodes = map[string]NPUNode{"node-both": {}}
				return sh
			}(),
			want: []*api.NodeInfo{},
		},
	}
}

func TestGetNeedInitNodeList(t *testing.T) {
	for _, tt := range buildGetNeedInitNodeListTest() {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sHandle.getNeedInitNodeList(tt.ssn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getNeedInitNodeList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateNPUNodeDpuInfos(t *testing.T) {
	t.Run("01-skip when dpu info update time is zero", func(t *testing.T) {
		node := &NPUNode{CommonNode: CommonNode{Annotation: map[string]string{}}}
		node.updateNPUNodeDpuInfos(k8s.DpuInfoWithNode{})
		if _, ok := node.Annotation[util.DpuInfoAnnoKey]; ok {
			t.Errorf("updateNPUNodeDpuInfos() should not set annotation when update time is zero")
		}
	})

	t.Run("02-set dpu info annotation when update time is valid", func(t *testing.T) {
		node := &NPUNode{CommonNode: CommonNode{Annotation: map[string]string{}}}
		dpuInfo := k8s.DpuInfoWithNode{
			DpuInfoCfg: k8s.DpuInfoCfg{
				DPUInfo: k8s.DPUInfoBody{
					DPUList: []k8s.DPUItem{{HcaName: "mlx5_0"}},
				},
				UpdateTime: 1234567890,
			},
		}
		node.updateNPUNodeDpuInfos(dpuInfo)
		dpuData, ok := node.Annotation[util.DpuInfoAnnoKey]
		if !ok {
			t.Errorf("updateNPUNodeDpuInfos() should set dpu info annotation")
		}
		if len(dpuData) == 0 {
			t.Errorf("updateNPUNodeDpuInfos() annotation should not be empty")
		}
	})

	t.Run("03-init annotation map when nil", func(t *testing.T) {
		node := &NPUNode{CommonNode: CommonNode{Annotation: nil}}
		node.updateNPUNodeDpuInfos(k8s.DpuInfoWithNode{
			DpuInfoCfg: k8s.DpuInfoCfg{UpdateTime: 100},
		})
		if node.Annotation == nil {
			t.Errorf("updateNPUNodeDpuInfos() should init annotation map")
		}
		if _, ok := node.Annotation[util.DpuInfoAnnoKey]; !ok {
			t.Errorf("updateNPUNodeDpuInfos() should set dpu info key")
		}
	})
}

const nested16TopoAnno = "[[[0,1],[2,3],[4,5],[6,7]],[[8,9],[10,11],[12,13],[14,15]]]"

func TestGetNPUCapacity(t *testing.T) {
	node910 := &api.NodeInfo{
		Capacity: &api.Resource{ScalarResources: map[v1.ResourceName]float64{
			v1.ResourceName(util.HwPreName + util.Ascend910): 8000,
		}},
	}
	if got := getNPUCapacity(node910); got != 8 {
		t.Errorf("getNPUCapacity(910) = %d, want 8", got)
	}

	nodeElastic := &api.NodeInfo{
		Capacity: &api.Resource{ScalarResources: map[v1.ResourceName]float64{
			v1.ResourceName(util.NPUCardName): 4000,
		}},
	}
	if got := getNPUCapacity(nodeElastic); got != 4 {
		t.Errorf("getNPUCapacity(elastic) = %d, want 4", got)
	}

	nodeEmpty := &api.NodeInfo{Capacity: &api.Resource{}}
	if got := getNPUCapacity(nodeEmpty); got != 0 {
		t.Errorf("getNPUCapacity(empty) = %d, want 0", got)
	}
}

func TestCardAnnoNpuPre(t *testing.T) {
	tests := []struct {
		key     string
		wantPre string
		wantOK  bool
	}{
		{"huawei.com/Ascend910-Unhealthy", util.NPU910CardNamePre, true},
		{"huawei.com/Ascend910b-NetworkUnhealthy", util.NPU910CardNamePre, true},
		{"huawei.com/Ascend310P-Unhealthy", util.NPU310PCardNamePre, true},
		{"huawei.com/Ascend310-NetworkUnhealthy", util.NPU310CardNamePre, true},
		{"huawei.com/npu-Unhealthy", util.NPUCardNamePre, true},
		{"huawei.com/unknown-key", "", false},
		{"totally-unrelated", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			pre, ok := cardAnnoNpuPre(tt.key)
			if pre != tt.wantPre || ok != tt.wantOK {
				t.Errorf("cardAnnoNpuPre(%q) = (%q, %v), want (%q, %v)", tt.key, pre, ok, tt.wantPre, tt.wantOK)
			}
		})
	}
}

func TestChipTopoOversize(t *testing.T) {
	// nested16 topo's max chip id is 15: a node owning card 16 is abnormal
	root := topo.ParseTopology(nested16TopoAnno)
	n := &NPUNode{CommonNode: CommonNode{Name: "n", ChipTopo: root}}
	if !n.chipTopoOversize(16) {
		t.Error("chipTopoOversize(16) should be true with max id 15")
	}
	if n.chipTopoOversize(15) {
		t.Error("chipTopoOversize(15) should be false with max id 15")
	}
	if n.chipTopoOversize(0) {
		t.Error("chipTopoOversize(0) should be false with max id 15")
	}
}

func TestParseChipTopology(t *testing.T) {
	// no topology annotation and no capacity -> BuildFlatTopology(0)="" ->
	// ParseTopology returns nil -> ChipTopo stays nil, method must not panic.
	n := &NPUNode{CommonNode: CommonNode{Name: "n", Annotation: map[string]string{}}}
	n.ParseChipTopology(&api.NodeInfo{Capacity: &api.Resource{}})
	if n.ChipTopo != nil {
		t.Errorf("no-annotation ParseChipTopology should leave ChipTopo nil, got %v", n.ChipTopo)
	}

	// explicit topo annotation -> real tree built, Raw preserved
	n2 := &NPUNode{CommonNode: CommonNode{Name: "n", Annotation: map[string]string{
		util.TopologyAnnoKey: nested16TopoAnno,
	}}}
	n2.ParseChipTopology(&api.NodeInfo{Capacity: &api.Resource{}})
	if n2.ChipTopo == nil {
		t.Fatal("explicit topo annotation should build a tree")
	}
	if n2.ChipTopo.Raw != nested16TopoAnno {
		t.Errorf("Raw = %q, want %q", n2.ChipTopo.Raw, nested16TopoAnno)
	}
	if got := n2.ChipTopo.MaxChipID(); got != 15 {
		t.Errorf("MaxChipID = %d, want 15", got)
	}
}

func TestParseChipTopologyCarvesFaults(t *testing.T) {
	// faulty-annotation cards (0,1) are excluded from the built tree's usable
	// pool, which the subsequent scheduling reads through SelectChips.
	n := &NPUNode{CommonNode: CommonNode{
		Name: "n",
		Annotation: map[string]string{
			util.TopologyAnnoKey: topo.BuildFlatTopology(8),
			util.HwPreName + util.Ascend910 + util.NPUUnhealthySuffix: "Ascend910-0,Ascend910-1",
		},
	}}
	n.ParseChipTopology(&api.NodeInfo{Capacity: &api.Resource{}})
	if n.ChipTopo == nil {
		t.Fatal("ParseChipTopology should build a tree")
	}
	got := n.ChipTopo.SelectChips(&util.Request{ReqNPUNum: 6, Mode: util.SoftScheduleMode})
	want := []int{2, 3, 4, 5, 6, 7}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("selection with faulty 0,1 = %v, want %v", got, want)
	}
}
