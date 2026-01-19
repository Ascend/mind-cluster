// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package statistics test for statistic funcs about fault
package statistics

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/interface/kube"
)

const (
	faultsStr   = "faults"
	faultNumStr = "faultNum"

	testCMName = "test-cm-name"
	testCMNS   = "default"
)

var testCM = &v1.ConfigMap{}

func TestUpdateFaultToCM(t *testing.T) {
	convey.Convey("test func 'UpdateFaultToCM', exceedsLimiter is true", t, testUpdateWhenTrue)
	convey.Convey("test func 'UpdateFaultToCM', exceedsLimiter is false", t, testUpdateWhenFalse)
}

func testUpdateWhenTrue() {
	p1 := gomonkey.ApplyFuncReturn(kube.PatchCMData, nil, nil)
	defer p1.Reset()
	err := UpdateFaultToCM(faultsStr, faultNumStr, true)
	convey.So(err, convey.ShouldBeNil)

	p2 := gomonkey.ApplyFuncReturn(kube.PatchCMData, nil, testErr)
	defer p2.Reset()
	err = UpdateFaultToCM(faultsStr, faultNumStr, true)
	expErr := fmt.Errorf("patch cm <%s> data failed", constant.StatisticFaultCMName)
	convey.So(err, convey.ShouldResemble, expErr)
}

func testUpdateWhenFalse() {
	p1 := gomonkey.ApplyFuncReturn(kube.UpdateOrCreateConfigMap, nil)
	defer p1.Reset()
	err := UpdateFaultToCM(faultsStr, faultNumStr, false)
	convey.So(err, convey.ShouldBeNil)

	p2 := gomonkey.ApplyFuncReturn(kube.UpdateOrCreateConfigMap, testErr)
	defer p2.Reset()
	err = UpdateFaultToCM(faultsStr, faultNumStr, false)
	expErr := fmt.Errorf("update or create cm <%s> failed", constant.StatisticFaultCMName)
	convey.So(err, convey.ShouldResemble, expErr)
}

func TestLoadFaultFromCM(t *testing.T) {
	constructTestCM()
	p1 := gomonkey.ApplyFuncReturn(kube.GetConfigMap, testCM, nil)
	defer p1.Reset()
	convey.Convey("test func 'LoadFaultFromCM' success", t, testLoadFault)
	convey.Convey("test func 'LoadFaultFromCM' failed, label key does not exist", t, testLoadFaultNoKey)
	convey.Convey("test func 'LoadFaultFromCM' failed, unmarshal error", t, testLoadFaultErrUnmarshal)
}

func testLoadFault() {
	err := LoadFaultFromCM()
	convey.So(err, convey.ShouldBeNil)

	p2 := gomonkey.ApplyFuncReturn(kube.GetConfigMap, testCM, testErr)
	defer p2.Reset()
	err = LoadFaultFromCM()
	expErr := fmt.Errorf("get cm <%s> failed", constant.StatisticFaultCMName)
	convey.So(err, convey.ShouldResemble, expErr)
}

func testLoadFaultNoKey() {
	p2 := gomonkey.ApplyFuncReturn(kube.GetConfigMap, &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testCMName,
			Namespace: testCMNS,
		},
		Data: nil,
	}, nil)
	defer p2.Reset()
	err := LoadFaultFromCM()
	expErr := fmt.Errorf("statistic fault cm <%s> has no key '%s'",
		constant.StatisticFaultCMName, constant.StatisticPubFaultKey)
	convey.So(err, convey.ShouldResemble, expErr)
}

func testLoadFaultErrUnmarshal() {
	p3 := gomonkey.ApplyFuncReturn(json.Unmarshal, testErr)
	defer p3.Reset()
	err := LoadFaultFromCM()
	expErr := fmt.Errorf("unmarshal node fault from cm <%s> failed", constant.StatisticFaultCMName)
	convey.So(err, convey.ShouldResemble, expErr)
}

func constructTestCM() {
	const (
		testFaultId   = "12345"
		testFaultCode = "123456789"
		testFaultTime = 1234567890
		testNodeName  = "node1"
	)
	cmData := make(map[string]string, 1)
	faults := make(map[string][]constant.NodeFault)
	nodeFault := constant.NodeFault{
		FaultResource: constant.FaultTypeNetwork,
		FaultDevIds:   []int32{0},
		FaultId:       testFaultId,
		FaultType:     constant.FaultTypeStorage,
		FaultCode:     testFaultCode,
		FaultLevel:    constant.SeparateNPU,
		FaultTime:     testFaultTime,
	}
	faults[testNodeName] = []constant.NodeFault{nodeFault}
	cmData[constant.StatisticPubFaultKey] = util.ObjToString(faults)
	testCM = &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testCMName,
			Namespace: testCMNS,
		},
		Data: cmData,
	}
}
