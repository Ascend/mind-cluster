// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault test for public fault processor
package publicfault

import (
	"sort"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/faultdomain"
	"clusterd/pkg/domain/publicfault"
)

const (
	testId        = "11937763019253715778"
	testTimeStamp = 1739866717
	testResource1 = "resource1"
	testResource2 = "resource2"
	testFaultCode = "010001001"
	testNodeName1 = "node1"
	testNodeName2 = "node2"
	testNodeName3 = "node3"
)

var (
	testCacheData = constant.PubFaultCache{
		FaultDevIds: []int32{0, 1},
		FaultId:     testId,
		FaultType:   constant.FaultTypeNPU,
		FaultCode:   testFaultCode,
		FaultLevel:  constant.SeparateNPU,
		FaultTime:   testTimeStamp,
		Assertion:   constant.AssertionOccur,
	}

	faultKey1 = testResource1 + testId
	faultKey2 = testResource2 + testId
)

func TestProcessor(t *testing.T) {
	convey.Convey("test func 'Process', public cache is nil", t, testNilCache)
	convey.Convey("test func 'Process', input type is invalid", t, testInputInvalid)
	convey.Convey("test func 'Process', public fault node name does not exist", t, testNodeNameInvalid)
	convey.Convey("test func 'Process', dp unhealthy card is different from public fault", t, testDiff)
	convey.Convey("test func 'Process', dp and public fault have common unhealthy card", t, testCommon)
}

func testNilCache() {
	resetFaultCache()
	ori := constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm]{
		AllConfigmap:    faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](oriDevInfo1),
		UpdateConfigmap: nil,
	}
	res := PubFaultProcessor.Process(ori)
	convey.So(res, convey.ShouldResemble, ori)
}

func testInputInvalid() {
	resetFaultCache()
	publicfault.PubFaultCache.AddPubFaultToCache(&testCacheData, testNodeName1, faultKey1)
	res := PubFaultProcessor.Process(nil)
	convey.So(res, convey.ShouldBeNil)
}

func testNodeNameInvalid() {
	resetFaultCache()
	publicfault.PubFaultCache.AddPubFaultToCache(&testCacheData, testNodeName3, faultKey1)
	content := constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm]{
		AllConfigmap:    faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](oriDevInfo1),
		UpdateConfigmap: nil,
	}
	exp := PubFaultProcessor.Process(content).(constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm])
	convey.So(content, convey.ShouldResemble, exp)
}

func testDiff() {
	resetFaultCache()
	publicfault.PubFaultCache.AddPubFaultToCache(&testCacheData, testNodeName1, faultKey1)
	content := constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm]{
		AllConfigmap:    faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](oriDevInfo1),
		UpdateConfigmap: nil,
	}
	resContent := PubFaultProcessor.Process(content).(constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm])
	sortDeviceFaultList(resContent.AllConfigmap)
	want := faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](expDeviceInfo1)
	sortDeviceFaultList(want)
	result := resContent.AllConfigmap
	convey.So(result, convey.ShouldResemble, want)
}

func testCommon() {
	resetFaultCache()
	const card5 = 5
	testCacheData.FaultDevIds = []int32{0, card5}
	publicfault.PubFaultCache.AddPubFaultToCache(&testCacheData, testNodeName2, faultKey2)
	content := constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm]{
		AllConfigmap:    faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](oriDevInfo2),
		UpdateConfigmap: nil,
	}
	resContent := PubFaultProcessor.Process(content).(constant.OneConfigmapContent[*constant.AdvanceDeviceFaultCm])
	hwlog.RunLog.Infof(util.ObjToString(resContent.AllConfigmap))
	hwlog.RunLog.Infof(
		util.ObjToString(faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](expDeviceInfo2)))
	sortDeviceFaultList(resContent.AllConfigmap)
	result := resContent.AllConfigmap
	want := faultdomain.GetAdvanceFaultCm[*constant.AdvanceDeviceFaultCm](expDeviceInfo2)
	sortDeviceFaultList(want)
	convey.So(result, convey.ShouldResemble, want)
}

func resetFaultCache() {
	for nodeName := range publicfault.PubFaultCache.GetPubFault() {
		delete(publicfault.PubFaultCache.GetPubFault(), nodeName)
	}
}

func sortDeviceFaultList(advanceFaultCm map[string]*constant.AdvanceDeviceFaultCm) {
	for _, advanceDeviceCm := range advanceFaultCm {
		for _, fault := range advanceDeviceCm.FaultDeviceList {
			sort.Slice(fault, func(i, j int) bool {
				return util.MakeDataHash(fault[i]) < util.MakeDataHash(fault[j])
			})
		}
		sort.Strings(advanceDeviceCm.CardUnHealthy)
		sort.Strings(advanceDeviceCm.NetworkUnhealthy)
		sort.Strings(advanceDeviceCm.Recovering)
		sort.Strings(advanceDeviceCm.AvailableDeviceList)
	}
}
