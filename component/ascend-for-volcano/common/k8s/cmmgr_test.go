/*
Copyright(C)2020-2025. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package k8s is using for the k8s operation.
*/
package k8s

import (
	"encoding/json"
	"reflect"
	"testing"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/test"
)

func fakeClusterInfoCm[T any](cmPrefix string) *v1.ConfigMap {
	var info T
	tmpInfo := map[string]T{
		"node0": info,
	}
	bytes, _ := json.Marshal(tmpInfo)
	nodeName := cmPrefix + "node0"
	return test.FakeConfigmap(nodeName, util.MindXDlNameSpace, map[string]string{nodeName: string(bytes)})
}

func TestInitCmInformer(t *testing.T) {
	t.Run("01 k8s client is nil", func(t *testing.T) {
		InitCmInformer(nil, true)
	})
	t.Run("02 start device info informer success with clusterd", func(t *testing.T) {
		InitCmInformer(fake.NewSimpleClientset(), true)
	})
	needStartInformer = true
	t.Run("03 start device info informer success without clutserd", func(t *testing.T) {
		InitCmInformer(fake.NewSimpleClientset(), false)
	})
}

type UpdateConfigMapTestCase struct {
	name      string
	cmManager *ClusterInfoWitchCm
	obj       interface{}
	operator  string
	want      *ClusterInfoWitchCm
}

func buildUpdateConfigMapTestCases() []UpdateConfigMapTestCase {
	return []UpdateConfigMapTestCase{
		{
			name:      "01 will return empty when cm manager is nil",
			cmManager: nil,
			obj:       nil,
			want:      nil,
		},
		{
			name:      "02 will return empty when cm is nil",
			cmManager: &ClusterInfoWitchCm{},
			obj:       nil,
			want:      &ClusterInfoWitchCm{},
		},
		{
			name:      "03 will return cm mgr when cm is device info and operator is add",
			cmManager: &cmManager,
			obj:       FakeDeviceInfoCMDataByNode("node0", FakeDeviceList()),
			operator:  util.AddOperator,
			want:      &cmManager,
		},
		{
			name:      "04 will return cm mgr when cm is device info and operator is delete",
			cmManager: &cmManager,
			obj:       FakeDeviceInfoCMDataByNode("node0", FakeDeviceList()),
			operator:  util.DeleteOperator,
			want:      &cmManager,
		},
		{
			name:      "05 will return cm mgr when cm is noded info and operator is add",
			cmManager: &cmManager,
			obj:       test.FakeConfigmap(util.NodeDCmInfoNamePrefix+"node0", util.MindXDlNameSpace, FakeNodeInfos()),
			operator:  util.AddOperator,
			want:      &cmManager,
		},
		{
			name:      "06 will return cm mgr when cm is noded info and operator is delete",
			cmManager: &cmManager,
			obj:       test.FakeConfigmap(util.NodeDCmInfoNamePrefix+"node0", util.MindXDlNameSpace, FakeNodeInfos()),
			operator:  util.DeleteOperator,
			want:      &cmManager,
		},
		{
			name:      "06 will return cm mgr when cm is noded info and operator is add but cm date is nil",
			cmManager: &cmManager,
			obj:       test.FakeConfigmap(util.NodeDCmInfoNamePrefix+"node0", util.MindXDlNameSpace, nil),
			operator:  util.AddOperator,
			want:      &cmManager,
		},
	}
}

func TestUpdateConfigMap(t *testing.T) {
	for _, tt := range buildUpdateConfigMapTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			tt.cmManager.updateConfigMap(tt.obj, tt.operator)
			if !reflect.DeepEqual(tt.cmManager, tt.want) {
				t.Errorf("update cm failed, cm manager is different. cmMgr is %v want %v", tt.cmManager, tt.want)
			}
		})
	}
}

func buildUpdateConfigMapClusterTestCases01() []UpdateConfigMapTestCase {
	return []UpdateConfigMapTestCase{
		{
			name:      "01 will return empty when cm manager is nil",
			cmManager: nil,
			obj:       nil,
			want:      nil,
		},
		{
			name:      "02 will return empty when cm is nil",
			cmManager: &ClusterInfoWitchCm{},
			obj:       nil,
			want:      &ClusterInfoWitchCm{},
		},
		{
			name:      "03 obj is cluster device info add test",
			cmManager: &cmManager,
			obj:       fakeClusterInfoCm[NodeDeviceInfoWithID](util.ClusterDeviceInfo),
			operator:  util.AddOperator,
			want:      &cmManager,
		},
		{
			name:      "04 obj is cluster device info delete test",
			cmManager: &cmManager,
			obj:       fakeClusterInfoCm[NodeDeviceInfoWithID](util.ClusterDeviceInfo),
			operator:  util.DeleteOperator,
			want:      &cmManager,
		},
	}
}

func buildUpdateConfigMapClusterTestCases02() []UpdateConfigMapTestCase {
	return []UpdateConfigMapTestCase{
		{
			name:      "05 obj is cluster noded info delete test",
			cmManager: &cmManager,
			obj:       fakeClusterInfoCm[NodeDNodeInfo](util.ClusterNodeInfo),
			operator:  util.DeleteOperator,
			want:      &cmManager,
		},
		{
			name:      "06 obj is cluster noded info add test",
			cmManager: &cmManager,
			obj:       fakeClusterInfoCm[NodeDNodeInfo](util.ClusterNodeInfo),
			operator:  util.AddOperator,
			want:      &cmManager,
		},
		{
			name:      "07 obj is cluster switch info delete test",
			cmManager: &cmManager,
			obj:       fakeClusterInfoCm[SwitchFaultInfo](util.ClusterSwitchInfo),
			operator:  util.DeleteOperator,
			want:      &cmManager,
		},
		{
			name:      "08 obj is cluster switch info add test",
			cmManager: &cmManager,
			obj:       fakeClusterInfoCm[SwitchFaultInfo](util.ClusterSwitchInfo),
			operator:  util.AddOperator,
			want:      &cmManager,
		},
	}
}

func TestUpdateConfigMapCluster(t *testing.T) {
	tests := append(buildUpdateConfigMapClusterTestCases01(), buildUpdateConfigMapClusterTestCases02()...)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cmManager.updateConfigMapCluster(tt.obj, tt.operator)
			if !reflect.DeepEqual(tt.cmManager, tt.want) {
				t.Errorf("update cm failed, cm manager is different. cmMgr is %v want %v", tt.cmManager, tt.want)
			}
		})
	}
}

func TestGetCmInfos(t *testing.T) {
	nodeList := []*api.NodeInfo{{Name: testName}}
	t.Run("GetSwitchInfos test, get empty switch info", func(t *testing.T) {
		if got := GetSwitchInfos(nodeList); !reflect.DeepEqual(got, map[string]SwitchFaultInfo{testName: {}}) {
			t.Errorf("GetSwitchInfos() = %v, want %v", got, map[string]SwitchFaultInfo{testName: {}})
		}
	})

	t.Run("GetNodeDInfos test, get empty nodeD info", func(t *testing.T) {
		if got := GetNodeDInfos(nodeList); !reflect.DeepEqual(got, map[string]NodeDNodeInfo{testName: {}}) {
			t.Errorf("GetSwitchInfos() = %v, want %v", got, map[string]NodeDNodeInfo{testName: {}})
		}
	})
	tmpDeviceList := NodeDeviceInfo{DeviceList: make(map[string]string)}
	tmpDeviceInfos := map[string]NodeDeviceInfoWithID{testName: {NodeDeviceInfo: tmpDeviceList}}
	needStartInformer = false
	t.Run("Get device info test, get empty device info with use clusterD", func(t *testing.T) {
		if got := GetDeviceInfosAndSetInformerStart(nodeList, true); !reflect.DeepEqual(got, tmpDeviceInfos) {
			t.Errorf("Get device info = %v, want %v", got, tmpDeviceInfos)
		}
	})

	t.Run("Get device info test, get empty device info without use clusterD", func(t *testing.T) {
		if got := GetDeviceInfosAndSetInformerStart(nodeList, false); !reflect.DeepEqual(got, tmpDeviceInfos) {
			t.Errorf("Get device info = %v, want %v", got, tmpDeviceInfos)
		}
	})
}
