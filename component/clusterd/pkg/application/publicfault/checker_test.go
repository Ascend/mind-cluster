// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault test for public fault checker
package publicfault

import (
	"encoding/json"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/publicfault"
	"clusterd/pkg/domain/statistics"
)

const (
	testId          = "11937763019444715778"
	testTimeStamp   = 1739866717000
	testResource    = "resource1"
	testFaultCode   = "000000001"
	testNodeName    = "node1"
	testNodeSN      = "node1-sn"
	testDescription = "fault description"

	testErrId         = "err id"
	testErrTimeStamp  = 123
	testErrTimeStamp2 = 1234567890000
	testErrFaultCode  = "err code"
	testErrAssertion  = "err assertion"
	testErrNodeName   = "err name~~"
	testErrNodeSN     = "err sn~~"
	testErrDevId      = 66
)

var (
	testPubFaultInfo *api.PubFaultInfo
	testFault        *api.Fault
	testInfluence    *api.Influence
)

var (
	faultInfoValid = api.PubFaultInfo{
		Id:        testId,
		TimeStamp: testTimeStamp,
		Version:   validVersion,
		Resource:  testResource,
		Faults:    []api.Fault{faultValid},
	}
	faultValid = api.Fault{
		FaultId:       testId,
		FaultType:     constant.FaultTypeNPU,
		FaultCode:     testFaultCode,
		FaultTime:     testTimeStamp,
		Assertion:     constant.AssertionOccur,
		FaultLocation: nil,
		Influence:     []api.Influence{influenceValid},
		Description:   "fault description",
	}
	influenceValid = api.Influence{
		NodeName:  testNodeName,
		DeviceIds: []int32{0},
	}
)

// test case
var (
	validFaultInfoTC = []struct {
		description string
		faultInfo   api.PubFaultInfo
	}{
		{"valid fault info", faultInfoValid},
	}

	errFaultInfoTC = []struct {
		description string
		faultInfo   api.PubFaultInfo
	}{
		{"id not exist", api.PubFaultInfo{
			TimeStamp: testTimeStamp,
			Version:   validVersion,
			Resource:  testResource,
			Faults:    []api.Fault{faultValid},
		}},
		{"id invalid", api.PubFaultInfo{
			Id:        testErrId,
			TimeStamp: testTimeStamp,
			Version:   validVersion,
			Resource:  testResource,
			Faults:    []api.Fault{faultValid},
		}},
		{"timestamp not exist", api.PubFaultInfo{
			Id:       testId,
			Version:  validVersion,
			Resource: testResource,
			Faults:   []api.Fault{faultValid},
		}},
		{"timestamp invalid, length error", api.PubFaultInfo{
			Id:        testId,
			TimeStamp: testErrTimeStamp,
			Version:   validVersion,
			Resource:  testResource,
			Faults:    []api.Fault{faultValid},
		}},
		{"timestamp invalid, size error", api.PubFaultInfo{
			Id:        testId,
			TimeStamp: testErrTimeStamp2,
			Version:   validVersion,
			Resource:  testResource,
			Faults:    []api.Fault{faultValid},
		}},
		{"version not exist", api.PubFaultInfo{
			Id:        testId,
			TimeStamp: testTimeStamp,
			Resource:  testResource,
			Faults:    []api.Fault{faultValid},
		}},
		{"version invalid", api.PubFaultInfo{
			Id:        testId,
			TimeStamp: testTimeStamp,
			Version:   "err version",
			Resource:  testResource,
			Faults:    []api.Fault{faultValid},
		}},
		{"resource not exist", api.PubFaultInfo{
			Id:        testId,
			TimeStamp: testTimeStamp,
			Version:   validVersion,
			Faults:    []api.Fault{faultValid},
		}},
		{"resource invalid", api.PubFaultInfo{
			Id:        testId,
			TimeStamp: testTimeStamp,
			Version:   validVersion,
			Resource:  "error resource",
			Faults:    []api.Fault{faultValid},
		}},
		{"faults not exist", api.PubFaultInfo{
			Id:        testId,
			TimeStamp: testTimeStamp,
			Version:   validVersion,
			Resource:  testResource,
		}},
		{"faults length invalid", api.PubFaultInfo{
			Id:        testId,
			TimeStamp: testTimeStamp,
			Version:   validVersion,
			Resource:  testResource,
			Faults:    nil,
		}},
	}

	validFaultTC = []struct {
		description string
		fault       api.Fault
	}{
		{"fault location not exist", api.Fault{
			FaultId:     testId,
			FaultType:   constant.FaultTypeNPU,
			FaultCode:   testFaultCode,
			FaultTime:   testTimeStamp,
			Assertion:   constant.AssertionOccur,
			Influence:   []api.Influence{influenceValid},
			Description: "fault description",
		}},
		{"description not exist", api.Fault{
			FaultId:       testId,
			FaultType:     constant.FaultTypeNPU,
			FaultCode:     testFaultCode,
			FaultTime:     testTimeStamp,
			Assertion:     constant.AssertionOccur,
			FaultLocation: nil,
			Influence:     []api.Influence{influenceValid},
		}},
	}

	errFaultTC = []struct {
		description string
		fault       api.Fault
	}{
		{"id not exist", api.Fault{
			FaultType:     constant.FaultTypeNPU,
			FaultCode:     testFaultCode,
			FaultTime:     testTimeStamp,
			Assertion:     constant.AssertionOccur,
			FaultLocation: nil,
			Influence:     []api.Influence{influenceValid},
			Description:   testDescription,
		}},
		{"id invalid", api.Fault{
			FaultId:       testErrId,
			FaultType:     constant.FaultTypeNPU,
			FaultCode:     testFaultCode,
			FaultTime:     testTimeStamp,
			Assertion:     constant.AssertionOccur,
			FaultLocation: nil,
			Influence:     []api.Influence{influenceValid},
			Description:   testDescription,
		}},
		{"fault type not exist", api.Fault{
			FaultId:       testId,
			FaultCode:     testFaultCode,
			FaultTime:     testTimeStamp,
			Assertion:     constant.AssertionOccur,
			FaultLocation: nil,
			Influence:     []api.Influence{influenceValid},
			Description:   testDescription,
		}},
		{"fault type invalid", api.Fault{
			FaultId:       testId,
			FaultType:     testErrId,
			FaultCode:     testFaultCode,
			FaultTime:     testTimeStamp,
			Assertion:     constant.AssertionOccur,
			FaultLocation: nil,
			Influence:     []api.Influence{influenceValid},
			Description:   testDescription,
		}},
		{"fault code not exist", api.Fault{
			FaultId:       testId,
			FaultType:     constant.FaultTypeNPU,
			FaultTime:     testTimeStamp,
			Assertion:     constant.AssertionOccur,
			FaultLocation: nil,
			Influence:     []api.Influence{influenceValid},
			Description:   testDescription,
		}},
		{"fault code invalid", api.Fault{
			FaultId:       testId,
			FaultType:     constant.FaultTypeNPU,
			FaultCode:     testErrFaultCode,
			FaultTime:     testTimeStamp,
			Assertion:     constant.AssertionOccur,
			FaultLocation: nil,
			Influence:     []api.Influence{influenceValid},
			Description:   testDescription,
		}},
		{"fault time not exist", api.Fault{
			FaultId:       testId,
			FaultType:     constant.FaultTypeNPU,
			FaultCode:     testFaultCode,
			Assertion:     constant.AssertionOccur,
			FaultLocation: nil,
			Influence:     []api.Influence{influenceValid},
			Description:   testDescription,
		}},
		{"fault time invalid, length error", api.Fault{
			FaultId:       testId,
			FaultType:     constant.FaultTypeNPU,
			FaultCode:     testFaultCode,
			FaultTime:     testErrTimeStamp,
			Assertion:     constant.AssertionOccur,
			FaultLocation: nil,
			Influence:     []api.Influence{influenceValid},
			Description:   testDescription,
		}},
		{"fault time invalid, size error", api.Fault{
			FaultId:       testId,
			FaultType:     constant.FaultTypeNPU,
			FaultCode:     testFaultCode,
			FaultTime:     testErrTimeStamp2,
			Assertion:     constant.AssertionOccur,
			FaultLocation: nil,
			Influence:     []api.Influence{influenceValid},
			Description:   testDescription,
		}},
		{"fault assertion not exist", api.Fault{
			FaultId:       testId,
			FaultType:     constant.FaultTypeNPU,
			FaultCode:     testFaultCode,
			FaultTime:     testTimeStamp,
			FaultLocation: nil,
			Influence:     []api.Influence{influenceValid},
			Description:   testDescription,
		}},
		{"fault assertion invalid", api.Fault{
			FaultId:       testId,
			FaultType:     constant.FaultTypeNPU,
			FaultCode:     testFaultCode,
			FaultTime:     testTimeStamp,
			Assertion:     testErrAssertion,
			FaultLocation: nil,
			Influence:     []api.Influence{influenceValid},
			Description:   testDescription,
		}},
		{"fault location invalid", api.Fault{
			FaultId:   testId,
			FaultType: constant.FaultTypeNPU,
			FaultCode: testFaultCode,
			FaultTime: testTimeStamp,
			Assertion: constant.AssertionOccur,
			FaultLocation: map[string]string{"1": "1", "2": "2", "3": "3", "4": "4", "5": "5", "6": "6", "7": "7",
				"8": "8", "9": "9", "10": "10", "11": "11", "12": "12"},
			Influence:   []api.Influence{influenceValid},
			Description: testDescription,
		}},
		{"influence invalid", api.Fault{
			FaultId:       testId,
			FaultType:     constant.FaultTypeNPU,
			FaultCode:     testFaultCode,
			FaultTime:     testTimeStamp,
			Assertion:     constant.AssertionOccur,
			FaultLocation: nil,
			Influence:     nil,
			Description:   testDescription,
		}},
	}

	validInfluenceTC = []struct {
		description string
		influence   api.Influence
	}{
		{"node name not exist", api.Influence{
			NodeSN:    testNodeSN,
			DeviceIds: []int32{0},
		}},
		{"node sn not exist", api.Influence{
			NodeSN:    testNodeSN,
			DeviceIds: []int32{0},
		}},
	}

	errInfluenceTC = []struct {
		description string
		influence   api.Influence
	}{
		{"both node name and sn not exist", api.Influence{
			DeviceIds: []int32{0},
		}},
		{"node name not exist, node sn does not exist in cache", api.Influence{
			NodeSN:    testErrNodeSN,
			DeviceIds: []int32{0},
		}},
		{"node sn not exist, node name invalid", api.Influence{
			NodeName:  testErrNodeName,
			DeviceIds: []int32{0},
		}},
		{"device id not exist", api.Influence{
			NodeName: testNodeName,
		}},
		{"device id value invalid", api.Influence{
			NodeName:  testNodeName,
			DeviceIds: []int32{testErrDevId},
		}},
		{"device id value invalid, duplicate ids", api.Influence{
			NodeName:  testNodeName,
			DeviceIds: []int32{0, 1, 1},
		}},
	}
)

func TestPubFaultInfoChecker(t *testing.T) {
	publicfault.PubFaultResource = []string{testResource}
	publicfault.PubFaultCodeCfg.SeparateNPUCodes[testFaultCode] = struct{}{}
	convey.Convey("test fault info success", t, testValidFaultInfo)
	convey.Convey("test fault info failed", t, testInvalidFaultInfo)
	publicfault.PubFaultResource = []string{}

	convey.Convey("test fault success", t, testValidFault)
	convey.Convey("test fault failed", t, testInvalidFault)

	statistics.GetNodeSNAndNameCache()[testNodeSN] = testNodeName
	convey.Convey("test influence success", t, testValidInfluence)
	delete(statistics.GetNodeSNAndNameCache(), testNodeSN)
	convey.Convey("test influence failed", t, testInvalidInfluence)
}

func testValidFaultInfo() {
	for _, testCase := range validFaultInfoTC {
		getTestFaultInfo(testCase.faultInfo)
		err := NewPubFaultInfoChecker(testPubFaultInfo).Check()
		convey.So(err, convey.ShouldBeNil)
	}
}

func testInvalidFaultInfo() {
	for _, testCase := range errFaultInfoTC {
		getTestFaultInfo(testCase.faultInfo)
		err := NewPubFaultInfoChecker(testPubFaultInfo).Check()
		hwlog.RunLog.Error(err)
		convey.So(err, convey.ShouldNotBeNil)
	}
}

func getTestFaultInfo(pubFaultInfo api.PubFaultInfo) {
	data, err := json.Marshal(pubFaultInfo)
	convey.So(err, convey.ShouldBeNil)
	err = json.Unmarshal(data, &testPubFaultInfo)
	convey.So(err, convey.ShouldBeNil)
}

func testValidFault() {
	for _, testCase := range validFaultTC {
		getTestFault(testCase.fault)
		var checker = faultChecker{fault: testFault}
		err := checker.check()
		convey.So(err, convey.ShouldBeNil)
	}
}

func testInvalidFault() {
	for _, testCase := range errFaultTC {
		getTestFault(testCase.fault)
		var checker = faultChecker{fault: testFault}
		err := checker.check()
		hwlog.RunLog.Error(err)
		convey.So(err, convey.ShouldNotBeNil)
	}
}

func getTestFault(fault api.Fault) {
	data, err := json.Marshal(fault)
	convey.So(err, convey.ShouldBeNil)
	err = json.Unmarshal(data, &testFault)
	convey.So(err, convey.ShouldBeNil)
}

func testValidInfluence() {
	for _, testCase := range validInfluenceTC {
		getTestInfluence(testCase.influence)
		var checker = influenceChecker{influence: testInfluence}
		err := checker.check()
		convey.So(err, convey.ShouldBeNil)
	}
}

func testInvalidInfluence() {
	for _, testCase := range errInfluenceTC {
		getTestInfluence(testCase.influence)
		var checker = influenceChecker{influence: testInfluence}
		err := checker.check()
		hwlog.RunLog.Error(err)
		convey.So(err, convey.ShouldNotBeNil)
	}
}

func getTestInfluence(influence api.Influence) {
	data, err := json.Marshal(influence)
	convey.So(err, convey.ShouldBeNil)
	err = json.Unmarshal(data, &testInfluence)
	convey.So(err, convey.ShouldBeNil)
}
