// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault test for public fault cache util
package publicfault

import (
	"errors"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

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

func TestPubFaultCache(t *testing.T) {
	resetCache()
	defer resetCache()
	convey.Convey("test PubFaultCache method 'AddPubFaultToCache'", t, testAdd)
	convey.Convey("test PubFaultCache method 'GetPubFaultByNodeName'", t, testGet)
	convey.Convey("test PubFaultCache method 'FaultExisted'", t, testFaultExisted)
	convey.Convey("test PubFaultCache method 'DeepCopy'", t, testDeepCopy)
	convey.Convey("test PubFaultCache method 'DeleteOccurFault'", t, testDelete)
}

func testAdd() {
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
}

func resetCache() {
	PubFaultCache = &pubFaultCache{
		faultCache: make(map[string]map[string]*constant.PubFaultCache),
		mutex:      sync.Mutex{},
	}
}

func testGet() {
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
