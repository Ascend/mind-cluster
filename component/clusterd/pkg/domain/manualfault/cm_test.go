/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package manualfault test for manual separate npu info cache
package manualfault

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/conf"
	"clusterd/pkg/interface/kube"
)

// node1, dev1, code1
func getDemoFaultInfo1() FaultInfo {
	return FaultInfo{
		NodeName:    node1,
		DevName:     dev1,
		FaultCode:   code1,
		ReceiveTime: receiveTime1,
	}
}

// node1, dev1, code2
func getDemoFaultInfo2() FaultInfo {
	return FaultInfo{
		NodeName:    node1,
		DevName:     dev1,
		FaultCode:   code2,
		ReceiveTime: receiveTime2,
	}
}

// node1, dev2, code1
func getDemoFaultInfo3() FaultInfo {
	return FaultInfo{
		NodeName:    node1,
		DevName:     dev2,
		FaultCode:   code1,
		ReceiveTime: receiveTime1,
	}
}

// node2, dev2, code1
func getDemoFaultInfo4() FaultInfo {
	return FaultInfo{
		NodeName:    node2,
		DevName:     dev2,
		FaultCode:   code1,
		ReceiveTime: receiveTime1,
	}
}

func getDemoNodeInfo() map[string]nodeCmInfo {
	return map[string]nodeCmInfo{
		node1: {
			Total: []string{dev1},
			Detail: map[string][]devCmInfo{
				dev1: {
					{
						FaultCode:        code1,
						FaultLevel:       constant.ManuallySeparateNPU,
						LastSeparateTime: receiveTime1,
					},
				},
			},
		},
	}
}

func TestAddSeparateDev(t *testing.T) {
	conf.SetManualSeparatePolicy(validPolicy)
	convey.Convey("test func AddSeparateDev, add new node", t, testAddNewNode)
	convey.Convey("test func AddSeparateDev, add new dev", t, testAddNewDev)
	convey.Convey("test func AddSeparateDev, add new code", t, testAddNewCode)
	convey.Convey("test func AddSeparateDev, add same code", t, testAddSameCode)
}

func testAddNewNode() {
	InitFaultCmInfo()
	fault1 := getDemoFaultInfo1()
	FaultCmInfo.AddSeparateDev(fault1)
	convey.So(FaultCmInfo.Len(), convey.ShouldEqual, len1)
	info, ok := FaultCmInfo.manualInfo[fault1.NodeName]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{fault1.DevName})
	detail, ok := info.Detail[fault1.DevName]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(detail, convey.ShouldResemble, []devCmInfo{
		{
			FaultCode:        fault1.FaultCode,
			FaultLevel:       constant.ManuallySeparateNPU,
			LastSeparateTime: fault1.ReceiveTime,
		},
	})
}

func testAddNewDev() {
	InitFaultCmInfo()
	fault1 := getDemoFaultInfo1()
	fault3 := getDemoFaultInfo3()
	FaultCmInfo.AddSeparateDev(fault1)
	FaultCmInfo.AddSeparateDev(fault3)
	convey.So(FaultCmInfo.Len(), convey.ShouldEqual, len1)
	info, ok := FaultCmInfo.manualInfo[fault1.NodeName]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{fault1.DevName, fault3.DevName})

	detail, ok := info.Detail[fault1.DevName]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(detail, convey.ShouldResemble, []devCmInfo{
		{
			FaultCode:        fault1.FaultCode,
			FaultLevel:       constant.ManuallySeparateNPU,
			LastSeparateTime: fault1.ReceiveTime,
		},
	})

	detail, ok = info.Detail[fault3.DevName]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(detail, convey.ShouldResemble, []devCmInfo{
		{
			FaultCode:        fault3.FaultCode,
			FaultLevel:       constant.ManuallySeparateNPU,
			LastSeparateTime: fault3.ReceiveTime,
		},
	})
}

func testAddNewCode() {
	InitFaultCmInfo()
	fault1 := getDemoFaultInfo1()
	fault2 := getDemoFaultInfo2()
	FaultCmInfo.AddSeparateDev(fault1)
	FaultCmInfo.AddSeparateDev(fault2)
	convey.So(FaultCmInfo.Len(), convey.ShouldEqual, len1)
	info, ok := FaultCmInfo.manualInfo[fault1.NodeName]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{fault1.DevName})
	detail, ok := info.Detail[fault1.DevName]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(detail, convey.ShouldResemble, []devCmInfo{
		{
			FaultCode:        fault1.FaultCode,
			FaultLevel:       constant.ManuallySeparateNPU,
			LastSeparateTime: fault1.ReceiveTime,
		},
		{
			FaultCode:        fault2.FaultCode,
			FaultLevel:       constant.ManuallySeparateNPU,
			LastSeparateTime: fault2.ReceiveTime,
		},
	})
}

func testAddSameCode() {
	InitFaultCmInfo()
	fault1 := getDemoFaultInfo1()
	fault2 := fault1
	fault2.ReceiveTime = receiveTime2
	FaultCmInfo.AddSeparateDev(fault1)
	FaultCmInfo.AddSeparateDev(fault2)
	convey.So(FaultCmInfo.Len(), convey.ShouldEqual, len1)
	info, ok := FaultCmInfo.manualInfo[fault1.NodeName]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{fault1.DevName})
	detail, ok := info.Detail[fault1.DevName]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(detail, convey.ShouldResemble, []devCmInfo{
		{
			FaultCode:        fault1.FaultCode,
			FaultLevel:       constant.ManuallySeparateNPU,
			LastSeparateTime: fault2.ReceiveTime,
		},
	})
}

func TestFaultCmInfo(t *testing.T) {
	convey.Convey("test FaultCmInfo method 'SetNodeInfo'", t, testSetNodeInfo)
	convey.Convey("test FaultCmInfo method 'DeepCopy'", t, testDeepCopy)
	convey.Convey("test FaultCmInfo method 'HasDevManualSep'", t, testHasDevManualSep)
}

func testSetNodeInfo() {
	InitFaultCmInfo()
	FaultCmInfo.SetNodeInfo(nil)
	convey.So(FaultCmInfo.Len(), convey.ShouldEqual, len0)
	FaultCmInfo.SetNodeInfo(getDemoNodeInfo())
	convey.So(FaultCmInfo.Len(), convey.ShouldEqual, len1)
}

func testDeepCopy() {
	cpInfo, err := FaultCmInfo.DeepCopy()
	convey.So(err, convey.ShouldBeNil)
	convey.So(len(cpInfo), convey.ShouldEqual, len1)

	p1 := gomonkey.ApplyFuncReturn(util.DeepCopy, testErr)
	defer p1.Reset()
	_, err = FaultCmInfo.DeepCopy()
	convey.So(err, convey.ShouldResemble, testErr)
}

func testHasDevManualSep() {
	res := FaultCmInfo.HasDevManualSep(node1, dev1)
	convey.So(res, convey.ShouldBeTrue)
	res = FaultCmInfo.HasDevManualSep(node1, dev2)
	convey.So(res, convey.ShouldBeFalse)
	res = FaultCmInfo.HasDevManualSep(node2, dev2)
	convey.So(res, convey.ShouldBeFalse)
}

func getDemoCm() *v1.ConfigMap {
	info := getDemoNodeInfo()
	data := convertNodeInfoToCmData(info)
	cm := &v1.ConfigMap{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Data:       data,
	}
	return cm
}

func TestDelete(t *testing.T) {
	convey.Convey("test FaultCmInfo method 'DeleteSeparateDev', delete dev", t, testDeleteDev)
	convey.Convey("test FaultCmInfo method 'DeleteDevCode', delete dev", t, testDeleteCode)
}

func testPrepareData() {
	InitFaultCmInfo()
	// node1, dev1, code1
	fault1 := getDemoFaultInfo1()
	FaultCmInfo.AddSeparateDev(fault1)
	// node1, dev1, code2
	fault2 := getDemoFaultInfo2()
	FaultCmInfo.AddSeparateDev(fault2)
	// node1, dev2, code1
	fault3 := getDemoFaultInfo3()
	FaultCmInfo.AddSeparateDev(fault3)
	// node2, dev2, code1
	fault4 := getDemoFaultInfo4()
	FaultCmInfo.AddSeparateDev(fault4)

	info, ok := FaultCmInfo.manualInfo[node1]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev1, dev2})
	detail, ok := info.Detail[dev1]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(detail, convey.ShouldResemble, []devCmInfo{
		{
			FaultCode:        fault1.FaultCode,
			FaultLevel:       constant.ManuallySeparateNPU,
			LastSeparateTime: fault1.ReceiveTime,
		},
		{
			FaultCode:        fault2.FaultCode,
			FaultLevel:       constant.ManuallySeparateNPU,
			LastSeparateTime: fault2.ReceiveTime,
		},
	})

	detail, ok = info.Detail[dev2]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(detail, convey.ShouldResemble, []devCmInfo{
		{
			FaultCode:        fault3.FaultCode,
			FaultLevel:       constant.ManuallySeparateNPU,
			LastSeparateTime: fault3.ReceiveTime,
		},
	})

	info, ok = FaultCmInfo.manualInfo[node2]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev2})
	detail, ok = info.Detail[dev2]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(detail, convey.ShouldResemble, []devCmInfo{
		{
			FaultCode:        fault4.FaultCode,
			FaultLevel:       constant.ManuallySeparateNPU,
			LastSeparateTime: fault4.ReceiveTime,
		},
	})
}

func testDeleteDev() {
	testPrepareData()
	FaultCmInfo.DeleteSeparateDev(node1, dev1)
	info, ok := FaultCmInfo.manualInfo[node1]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev2})
	_, ok = info.Detail[dev1]
	convey.So(ok, convey.ShouldBeFalse)
	_, ok = info.Detail[dev2]
	convey.So(ok, convey.ShouldBeTrue)

	FaultCmInfo.DeleteSeparateDev(node1, dev2)
	_, ok = FaultCmInfo.manualInfo[node1]
	convey.So(ok, convey.ShouldBeFalse)
	_, ok = FaultCmInfo.manualInfo[node2]
	convey.So(ok, convey.ShouldBeTrue)

	FaultCmInfo.DeleteSeparateDev(node2, dev2)
	_, ok = FaultCmInfo.manualInfo[node2]
	convey.So(ok, convey.ShouldBeFalse)
	convey.So(FaultCmInfo.Len(), convey.ShouldEqual, len0)

	FaultCmInfo.DeleteSeparateDev(node3, dev2)
	convey.So(FaultCmInfo.Len(), convey.ShouldEqual, len0)
}

func testDeleteCode() {
	testPrepareData()
	info1, err := FaultCmInfo.DeepCopy()
	convey.So(err, convey.ShouldBeNil)
	FaultCmInfo.DeleteDevCode(node3, dev3, code1)
	info2, err := FaultCmInfo.DeepCopy()
	convey.So(err, convey.ShouldBeNil)
	convey.So(reflect.DeepEqual(info1, info2), convey.ShouldBeTrue)

	info1, err = FaultCmInfo.DeepCopy()
	FaultCmInfo.DeleteDevCode(node1, dev3, code1)
	info2, err = FaultCmInfo.DeepCopy()
	convey.So(err, convey.ShouldBeNil)
	convey.So(reflect.DeepEqual(info1, info2), convey.ShouldBeTrue)

	FaultCmInfo.DeleteDevCode(node1, dev1, code1)
	info, ok := FaultCmInfo.manualInfo[node1]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev1, dev2})
	detail, ok := info.Detail[dev1]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(detail, convey.ShouldResemble, []devCmInfo{
		{
			FaultCode:        code2,
			FaultLevel:       constant.ManuallySeparateNPU,
			LastSeparateTime: receiveTime2,
		},
	})

	FaultCmInfo.DeleteDevCode(node1, dev1, code2)
	info, ok = FaultCmInfo.manualInfo[node1]
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(info.Total, convey.ShouldResemble, []string{dev2})
	_, ok = info.Detail[dev1]
	convey.So(ok, convey.ShouldBeFalse)

	FaultCmInfo.DeleteDevCode(node1, dev2, code1)
	info, ok = FaultCmInfo.manualInfo[node1]
	convey.So(ok, convey.ShouldBeFalse)
	_, ok = FaultCmInfo.manualInfo[node2]
	convey.So(ok, convey.ShouldBeTrue)

	FaultCmInfo.DeleteDevCode(node2, dev2, code1)
	info, ok = FaultCmInfo.manualInfo[node2]
	convey.So(ok, convey.ShouldBeFalse)
}

func TestGetSepNPUByCurrentCmInfo(t *testing.T) {
	convey.Convey("test func 'GetSepNPUByCurrentCmInfo', cm is nil", t, testGetSepNPUByNilCm)
	convey.Convey("test func 'GetSepNPUByCurrentCmInfo', data is nil", t, testGetSepNPUNilData)
	convey.Convey("test func 'GetSepNPUByCurrentCmInfo', get success", t, testGetSepNPU)
	convey.Convey("test func 'GetSepNPUByCurrentCmInfo', parse error", t, testGetSepNPUErrParse)
}

func testGetSepNPUByNilCm() {
	devs := GetSepNPUByCurrentCmInfo(nil)
	convey.So(len(devs), convey.ShouldEqual, len0)
}

func testGetSepNPU() {
	cm := getDemoCm()
	devs := GetSepNPUByCurrentCmInfo(cm)
	convey.So(len(devs), convey.ShouldEqual, len1)
	convey.So(devs, convey.ShouldResemble, map[string][]string{node1: {dev1}})
}

func testGetSepNPUNilData() {
	cm := getDemoCm()
	cm.Data = nil
	devs := GetSepNPUByCurrentCmInfo(cm)
	convey.So(len(devs), convey.ShouldEqual, len0)
}

func testGetSepNPUErrParse() {
	cm := getDemoCm()
	cm.Data = map[string]string{"123": "456"}
	devs := GetSepNPUByCurrentCmInfo(cm)
	convey.So(len(devs), convey.ShouldEqual, len0)
}

func TestDeleteManualCm(t *testing.T) {
	convey.Convey("test func 'DeleteManualCm' success, delete success", t, testDelete)
	convey.Convey("test func 'DeleteManualCm' failed, delete error", t, testDeleteErr)
}

func testDelete() {
	LastCmInfo = getDemoNodeInfo()
	p1 := gomonkey.ApplyFuncReturn(kube.DeleteConfigMap, nil)
	defer p1.Reset()
	DeleteManualCm()
	convey.So(len(LastCmInfo), convey.ShouldEqual, len0)
}

func testDeleteErr() {
	LastCmInfo = getDemoNodeInfo()
	p1 := gomonkey.ApplyFuncReturn(kube.DeleteConfigMap, testErr)
	defer p1.Reset()
	DeleteManualCm()
	convey.So(len(LastCmInfo), convey.ShouldEqual, len(getDemoNodeInfo()))
}

func TestTryGetManualCm(t *testing.T) {
	convey.Convey("test func 'TryGetManualCm' success", t, testTryGetManualCm)
	convey.Convey("test func 'TryGetManualCm' failed, get cm error", t, testTryErrGetCm)
}

func testTryGetManualCm() {
	p1 := gomonkey.ApplyFuncReturn(kube.GetConfigMap, getDemoCm(), nil)
	defer p1.Reset()
	cm, err := TryGetManualCm()
	convey.So(err, convey.ShouldBeNil)
	convey.So(cm, convey.ShouldResemble, getDemoCm())
}

func testTryErrGetCm() {
	p1 := gomonkey.ApplyFuncReturn(kube.GetConfigMap, nil, testErr)
	defer p1.Reset()
	cm, err := TryGetManualCm()
	convey.So(err, convey.ShouldResemble, testErr)
	convey.So(cm, convey.ShouldBeNil)
}

func TestUpdateOrCreateManualCm(t *testing.T) {
	convey.Convey("test func 'UpdateOrCreateManualCm' success", t, testUpdateOrCreateManualCm)
	convey.Convey("test func 'UpdateOrCreateManualCm' success, marshal error", t, testMarshalErr)
	convey.Convey("test func 'UpdateOrCreateManualCm' failed, update error", t, testUpdateErr)
	convey.Convey("test func 'UpdateOrCreateManualCm' failed, deep copy error", t, testDeepCpErr)
	convey.Convey("test func 'UpdateOrCreateManualCm' success, manual info is nil", t, testNilInfo)
}

func testUpdateOrCreateManualCm() {
	p1 := gomonkey.ApplyFuncReturn(kube.UpdateOrCreateConfigMap, nil)
	defer p1.Reset()
	LastCmInfo = make(map[string]nodeCmInfo)
	nodeInfo := getDemoNodeInfo()
	FaultCmInfo.SetNodeInfo(nodeInfo)
	UpdateOrCreateManualCm()
	convey.So(reflect.DeepEqual(LastCmInfo, nodeInfo), convey.ShouldBeTrue)
}

func testMarshalErr() {
	p1 := gomonkey.ApplyFuncReturn(json.Marshal, nil, testErr).
		ApplyFuncReturn(kube.UpdateOrCreateConfigMap, nil)
	defer p1.Reset()
	LastCmInfo = make(map[string]nodeCmInfo)
	nodeInfo := getDemoNodeInfo()
	FaultCmInfo.SetNodeInfo(nodeInfo)
	UpdateOrCreateManualCm()
	convey.So(len(LastCmInfo), convey.ShouldEqual, len0)
}

func testUpdateErr() {
	p1 := gomonkey.ApplyFuncReturn(kube.UpdateOrCreateConfigMap, testErr)
	defer p1.Reset()
	LastCmInfo = make(map[string]nodeCmInfo)
	nodeInfo := getDemoNodeInfo()
	FaultCmInfo.SetNodeInfo(nodeInfo)
	UpdateOrCreateManualCm()
	convey.So(reflect.DeepEqual(LastCmInfo, nodeInfo), convey.ShouldBeFalse)
}

func testDeepCpErr() {
	p1 := gomonkey.ApplyFuncReturn(util.DeepCopy, testErr)
	defer p1.Reset()
	LastCmInfo = make(map[string]nodeCmInfo)
	nodeInfo := getDemoNodeInfo()
	FaultCmInfo.SetNodeInfo(nodeInfo)
	UpdateOrCreateManualCm()
	convey.So(len(LastCmInfo), convey.ShouldEqual, len0)
}

func testNilInfo() {
	LastCmInfo = make(map[string]nodeCmInfo)
	FaultCmInfo.SetNodeInfo(nil)
	p1 := gomonkey.ApplyFuncReturn(kube.DeleteConfigMap, nil)
	defer p1.Reset()
	UpdateOrCreateManualCm()
	convey.So(len(LastCmInfo), convey.ShouldEqual, len0)
}
