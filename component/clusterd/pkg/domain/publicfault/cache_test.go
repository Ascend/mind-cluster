// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault test for public fault cache util
package publicfault

import (
	"errors"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

const (
	cacheLen      = 2
	notExistNode  = "not exist node"
	notExistFault = "not exist fault"
)

var (
	faultKey1 = testResource1 + testId
	faultKey2 = testResource2 + testId
)

func resetCache() {
	PubFaultCache = &cache{
		faultCache: make(map[string]map[string]*constant.PubFaultCache),
		mutex:      sync.Mutex{},
	}
}

func TestPubFaultCache(t *testing.T) {
	resetCache()
	defer resetCache()
	convey.Convey("test PubFaultCache method 'AddPubFaultToCache'", t, testAdd)
	convey.Convey("test PubFaultCache method 'GetPubFaultByNodeName'", t, testGet)
	convey.Convey("test PubFaultCache method 'FaultExisted'", t, testFaultExisted)
	convey.Convey("test PubFaultCache method 'DeepCopy'", t, testDeepCopy)
	convey.Convey("test PubFaultCache method 'DeleteOccurFault'", t, testDelete)
	convey.Convey("test PubFaultCache method 'GetPubFaultsForCM'", t, testGetPubFaultsForCM)
	convey.Convey("test PubFaultCache method 'LoadFaultToCache'", t, testLoadFaultToCache)
	convey.Convey("test PubFaultCache method 'GetPubFaultNum'", t, testGetPubFaultNum)
}

func testAdd() {
	// input is nil
	oldLen := len(PubFaultCache.faultCache)
	PubFaultCache.AddPubFaultToCache(nil, testNodeName1, faultKey1)
	newLen := len(PubFaultCache.faultCache)
	convey.So(oldLen, convey.ShouldEqual, newLen)

	// both node and fault not exist
	PubFaultCache.AddPubFaultToCache(&testCacheData, testNodeName1, faultKey1)
	convey.So(len(PubFaultCache.faultCache), convey.ShouldEqual, 1)
	convey.So(len(PubFaultCache.faultCache[testNodeName1]), convey.ShouldEqual, 1)

	PubFaultCache.AddPubFaultToCache(&testCacheData, testNodeName2, faultKey2)
	convey.So(len(PubFaultCache.faultCache), convey.ShouldEqual, cacheLen)

	// node exist but fault not exist
	PubFaultCache.AddPubFaultToCache(&testCacheData, testNodeName1, faultKey2)
	convey.So(len(PubFaultCache.faultCache[testNodeName1]), convey.ShouldEqual, cacheLen)

	// both node and fault exist
	PubFaultCache.AddPubFaultToCache(&testCacheData, testNodeName1, faultKey1)
	convey.So(len(PubFaultCache.faultCache), convey.ShouldEqual, cacheLen)

	// public fault number in cache exceeds the upper limit
	const maxPubFaultCacheNum = 50000
	p1 := gomonkey.ApplyMethodReturn(&cache{}, "GetPubFaultNum", maxPubFaultCacheNum)
	defer p1.Reset()
	resetCache()
	PubFaultCache.AddPubFaultToCache(&testCacheData, testNodeName1, faultKey1)
	convey.So(len(PubFaultCache.faultCache), convey.ShouldEqual, 0)
}

func testGet() {
	resetCache()
	PubFaultCache.AddPubFaultToCache(&testCacheData, testNodeName1, faultKey1)
	PubFaultCache.AddPubFaultToCache(&testCacheData, testNodeName2, faultKey2)
	PubFaultCache.AddPubFaultToCache(&testCacheData, testNodeName1, faultKey2)
	// node exist
	nodeFault, nodeExisted := PubFaultCache.GetPubFaultByNodeName(testNodeName1)
	convey.So(nodeExisted, convey.ShouldBeTrue)
	convey.So(nodeFault[faultKey1], convey.ShouldEqual, &testCacheData)

	// node not exist
	nodeFault, nodeExisted = PubFaultCache.GetPubFaultByNodeName(notExistNode)
	convey.So(nodeExisted, convey.ShouldBeFalse)
	convey.So(nodeFault, convey.ShouldBeNil)
}

func testFaultExisted() {
	// both node and fault exist
	existed, _ := PubFaultCache.FaultExisted(testNodeName1, faultKey1)
	convey.So(existed, convey.ShouldBeTrue)

	// node exist but fault not exist
	existed, _ = PubFaultCache.FaultExisted(testNodeName1, notExistFault)
	convey.So(existed, convey.ShouldBeFalse)

	// both node and fault does not exist
	existed, _ = PubFaultCache.FaultExisted(notExistNode, faultKey1)
	convey.So(existed, convey.ShouldBeFalse)
}

func testDeepCopy() {
	_, err := PubFaultCache.DeepCopy()
	convey.So(err, convey.ShouldBeNil)

	p1 := gomonkey.ApplyFuncReturn(util.DeepCopy, testErr)
	defer p1.Reset()
	_, err = PubFaultCache.DeepCopy()
	expErr := errors.New("deep copy public fault in cache failed")
	convey.So(err, convey.ShouldResemble, expErr)
}

func testDelete() {
	// both node and fault exist
	PubFaultCache.DeleteOccurFault(testNodeName1, faultKey1)
	convey.So(len(PubFaultCache.faultCache[testNodeName1]), convey.ShouldEqual, 1)

	// node exist but fault not exist
	PubFaultCache.DeleteOccurFault(testNodeName1, notExistFault)
	convey.So(len(PubFaultCache.faultCache[testNodeName1]), convey.ShouldEqual, 1)

	// both node and fault does not exist
	PubFaultCache.DeleteOccurFault(notExistNode, notExistFault)
	convey.So(len(PubFaultCache.faultCache), convey.ShouldEqual, cacheLen)
}

func testGetPubFaultsForCM() {
	resetCache()
	PubFaultCache.AddPubFaultToCache(&testCacheData, testNodeName1, faultKey1)
	_, pubFaultsNum := PubFaultCache.GetPubFaultsForCM()
	convey.So(pubFaultsNum, convey.ShouldEqual, 1)
}

func testLoadFaultToCache() {
	const (
		testFaultId   = "12345"
		testFaultId2  = "54321"
		testFaultCode = "123456789"
		testFaultTime = 1234567890
	)

	resetCache()
	faults := map[string][]constant.NodeFault{
		testNodeName1: {
			{
				FaultResource: constant.PublicFaultType,
				FaultDevIds:   []int32{0},
				FaultId:       testFaultId,
				FaultType:     constant.FaultTypeStorage,
				FaultCode:     testFaultCode,
				FaultLevel:    constant.SeparateNPU,
				FaultTime:     testFaultTime,
			},
			{
				FaultResource: constant.PublicFaultType,
				FaultDevIds:   []int32{1},
				FaultId:       testFaultId2,
				FaultType:     constant.FaultTypeNetwork,
				FaultCode:     testFaultCode,
				FaultLevel:    constant.SeparateNPU,
				FaultTime:     testFaultTime,
			},
		},
		testNodeName2: {{
			FaultResource: constant.PublicFaultType,
			FaultDevIds:   []int32{0},
			FaultId:       testFaultId,
			FaultType:     constant.FaultTypeNode,
			FaultCode:     testFaultCode,
			FaultLevel:    constant.SeparateNPU,
			FaultTime:     testFaultTime,
		}},
	}
	PubFaultCache.LoadFaultToCache(faults)
	for nodeName, nodeFault := range PubFaultCache.GetPubFault() {
		hwlog.RunLog.Info(nodeName)
		for faultKey, fault := range nodeFault {
			hwlog.RunLog.Info(faultKey)
			hwlog.RunLog.Infof("%#v", fault)
		}
	}
	convey.So(len(PubFaultCache.faultCache), convey.ShouldEqual, cacheLen)
}

func testGetPubFaultNum() {
	const expFaultNum = 3
	resetCache()
	PubFaultCache.AddPubFaultToCache(&testCacheData, testNodeName1, faultKey1)
	PubFaultCache.AddPubFaultToCache(&testCacheData, testNodeName2, faultKey2)
	PubFaultCache.AddPubFaultToCache(&testCacheData, testNodeName1, faultKey2)
	faultNum := PubFaultCache.GetPubFaultNum()
	convey.So(faultNum, convey.ShouldEqual, expFaultNum)
}
