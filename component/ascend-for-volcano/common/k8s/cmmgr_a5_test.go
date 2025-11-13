/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package k8s for the k8s operation
package k8s

import (
	"encoding/json"
	"testing"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/common/util"
)

const (
	testNodeName1         = "node1"
	testClusterdDpuCmName = util.DpuCmInfoNamePrefixByClusterd + "0"
)

func generateMockClusterdDpuCm() *v1.ConfigMap {
	dpuCmName := util.DpuCmInfoNamePrefixByDp + testNodeName1
	mockClusterdDpuCmData := map[string]DpuCMInfo{
		dpuCmName: {},
	}
	dataBytes, err := json.Marshal(mockClusterdDpuCmData)
	if err != nil {
		return nil
	}
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: testClusterdDpuCmName,
		},
		Data: map[string]string{testClusterdDpuCmName: string(dataBytes)},
	}
	return cm
}

func TestDealClusterDpuInfo(t *testing.T) {
	t.Run("should ignore configmap with wrong prefix", func(t *testing.T) {
		testDealClusterDpuInfoPrefix(t)
	})
	t.Run("should do nothing if getDataFromCM returns error", func(t *testing.T) {
		testDealClusterDpuInfoGetDataErr(t)
	})
	t.Run("should add or update DPU info on add/update operator", func(t *testing.T) {
		testDealClusterDpuInfoAddUpdate(t)
	})
	t.Run("should delete DPU info on delete operator", func(t *testing.T) {
		testDealClusterDpuInfoDelete(t)
	})
}

func testDealClusterDpuInfoPrefix(t *testing.T) {
	cmMgr := NewClusterInfoWitchCm()
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "wrong-prefix-cm",
		},
		Data: map[string]string{},
	}
	cmMgr.dpuInfosFromCm.Dpus = make(map[string]DpuCMInfo)
	cmMgr.dealClusterDpuInfo(cm, util.AddOperator)
	if len(cmMgr.dpuInfosFromCm.Dpus) != 0 {
		t.Fatalf("dealClusterDpuInfo should not add dpu info for wrong prefix")
	}
}
func testDealClusterDpuInfoGetDataErr(t *testing.T) {
	cmMgr := NewClusterInfoWitchCm()
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: testClusterdDpuCmName,
		},
		Data: nil, // will cause getDataFromCM to return error
	}
	cmMgr.dpuInfosFromCm.Dpus = make(map[string]DpuCMInfo)
	cmMgr.dealClusterDpuInfo(cm, util.AddOperator)
	if len(cmMgr.dpuInfosFromCm.Dpus) != 0 {
		t.Fatalf("dealClusterDpuInfo should not add dpu info when getDataFromCM fails")
	}
}

func testDealClusterDpuInfoAddUpdate(t *testing.T) {
	cmMgr := NewClusterInfoWitchCm()
	cm := generateMockClusterdDpuCm()
	cmMgr.dpuInfosFromCm.Dpus = make(map[string]DpuCMInfo)
	cmMgr.dealClusterDpuInfo(cm, util.AddOperator)
	info, ok := cmMgr.dpuInfosFromCm.Dpus[testNodeName1]
	if !ok {
		t.Fatalf("dealClusterDpuInfo should add DPU info for node")
	}
	// Check CacheUpdateTime is set (should be > 0)
	if info.CacheUpdateTime == 0 {
		t.Fatalf("CacheUpdateTime should be set on add")
	}
	// Test update
	oldTime := info.CacheUpdateTime
	time.Sleep(1 * time.Second)
	cmMgr.dealClusterDpuInfo(cm, util.UpdateOperator)
	info2 := cmMgr.dpuInfosFromCm.Dpus[testNodeName1]
	if info2.CacheUpdateTime <= oldTime {
		t.Fatalf("CacheUpdateTime should be updated on update")
	}
}

func testDealClusterDpuInfoDelete(t *testing.T) {
	cmMgr := NewClusterInfoWitchCm()
	cm := generateMockClusterdDpuCm()
	cmMgr.dpuInfosFromCm.Dpus = make(map[string]DpuCMInfo)
	// Add first
	cmMgr.dealClusterDpuInfo(cm, util.AddOperator)
	if len(cmMgr.dpuInfosFromCm.Dpus) != 1 {
		t.Fatalf("setup failed: dpu info not added")
	}
	// Now delete
	cmMgr.dealClusterDpuInfo(cm, util.DeleteOperator)
	if len(cmMgr.dpuInfosFromCm.Dpus) != 0 {
		t.Fatalf("dealClusterDpuInfo should delete dpu info on delete")
	}
}

func TestGetDpuInfosAndSetInformerStart(t *testing.T) {
	// Prepare a node list for test
	nodeList := []*api.NodeInfo{
		{Name: "node1"},
		{Name: "node2"},
		{Name: "node3"},
	}
	const expectedCount = 2

	// Prepare fake DPU info
	fakeDpuInfo1 := DpuCMInfo{CacheUpdateTime: 123}
	fakeDpuInfo2 := DpuCMInfo{CacheUpdateTime: 456}

	// Reset cmManager and its DPU map
	cmManager = NewClusterInfoWitchCm()
	cmManager.dpuInfosFromCm.Dpus["node1"] = fakeDpuInfo1
	cmManager.dpuInfosFromCm.Dpus["node2"] = fakeDpuInfo2

	// Call function under test
	result := GetDpuInfos(nodeList)

	// Check that only nodes with DPU info are returned
	if len(result) != expectedCount {
		t.Errorf("expected 2 dpu infos, got %d", len(result))
	}
	if _, ok := result["node1"]; !ok {
		t.Errorf("expected node1 in result")
	}
	if _, ok := result["node2"]; !ok {
		t.Errorf("expected node2 in result")
	}
	if _, ok := result["node3"]; ok {
		t.Errorf("did not expect node3 in result")
	}
}
