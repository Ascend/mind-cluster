// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault test for public fault collector
package publicfault

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/api"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/node"
	"clusterd/pkg/domain/publicfault"
)

const (
	sec1     = 1 * time.Second
	diffTime = 5 * time.Second
)

var (
	faultInfo = api.PubFaultInfo{
		Id:        testId1,
		TimeStamp: testTimeStamp,
		Version:   validVersion,
		Resource:  testResource1,
		Faults:    []api.Fault{fault},
	}
	fault = api.Fault{
		FaultId:       testId1,
		FaultType:     constant.FaultTypeNPU,
		FaultCode:     testFaultCode,
		FaultTime:     testTimeStamp,
		Assertion:     constant.AssertionOccur,
		FaultLocation: nil,
		Influence:     []api.Influence{influence},
		Description:   "fault description",
	}
	influence = api.Influence{
		NodeName:  testNodeName1,
		DeviceIds: []int32{0},
	}

	faultCache *constant.PubFaultCache
)

func TestPubFaultCollector(t *testing.T) {
	patches := gomonkey.ApplyFuncReturn(publicfault.LoadPubFaultCfgFromFile, nil).
		ApplyFuncReturn(LimitByResource, nil).
		ApplyMethodReturn(&pubFaultInfoChecker{}, "CheckAndFlush", nil)
	defer patches.Reset()
	convey.Convey("test func PubFaultCollector success", t, testLoadCustomFile)
	convey.Convey("test func PubFaultCollector failed, limit error", t, testErrLimit)
	convey.Convey("test func PubFaultCollector failed, check error", t, testErrCheck)
}

func testLoadCustomFile() {
	resetInitOnce()
	err := PubFaultCollector(&faultInfo)
	convey.So(err, convey.ShouldBeNil)

	resetInitOnce()
	const time2 = 2
	output := []gomonkey.OutputCell{
		{Values: gomonkey.Params{testErr}, Times: time2},
		{Values: gomonkey.Params{nil}},
	}
	p1 := gomonkey.ApplyFuncSeq(publicfault.LoadPubFaultCfgFromFile, output)
	defer p1.Reset()
	err = PubFaultCollector(&faultInfo)
	convey.So(err, convey.ShouldBeNil)
}

func testErrLimit() {
	p2 := gomonkey.ApplyFuncReturn(LimitByResource, testErr)
	defer p2.Reset()
	err := PubFaultCollector(&faultInfo)
	expErr := errors.New("limiter work by resource failed")
	convey.So(err, convey.ShouldResemble, expErr)
}

func testErrCheck() {
	p3 := gomonkey.ApplyMethodReturn(&pubFaultInfoChecker{}, "CheckAndFlush", testErr)
	defer p3.Reset()
	err := PubFaultCollector(&faultInfo)
	expErr := fmt.Errorf("check public fault info failed, error: %v", testErr)
	convey.So(err, convey.ShouldResemble, expErr)
}

func resetInitOnce() {
	pubFaultInitOnce = sync.Once{}
}

func TestGetNodeName(t *testing.T) {
	inf := api.Influence{
		NodeSN:    testNodeSN1,
		DeviceIds: []int32{0},
	}
	nodeInfo := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNodeName1,
			Annotations: map[string]string{
				"product-serial-number": testNodeSN1,
			},
		},
	}
	node.SaveNodeToCache(nodeInfo)
	convey.Convey("test func getNodeName when node name does not exist, get sn success", t, func() {
		res := getNodeName(inf)
		convey.So(res, convey.ShouldEqual, testNodeName1)
	})
	convey.Convey("test func getNodeName when node name does not exist, get sn failed", t, func() {
		inf.NodeSN = testNodeSN2
		res := getNodeName(inf)
		convey.So(res, convey.ShouldEqual, "")
	})
}

func TestDealFault(t *testing.T) {
	faultCache = convertPubFaultInfoToCache(fault, influence)
	convey.Convey("test func dealFault when fault assertion is occur", t, testOccur)
	convey.Convey("test func dealFault when fault assertion is recover, fault existed", t, testFaultExisted)
	convey.Convey("test func dealFault when fault assertion is recover. first recover, second occur", t, testRecoverOccur)
	convey.Convey("test func dealFault when fault assertion is recover. first occur, second recover", t, testOccurRecover)
	convey.Convey("test func dealFault when fault assertion is once", t, testOnce)
}

func testOccur() {
	resetFaultCache()

	// occur fault, and fault does not exist before
	dealFault(constant.AssertionOccur, testNodeName1, faultKey1, faultCache)
	convey.So(publicfault.PubFaultCache.GetPubFaultNum(), convey.ShouldEqual, 1)

	// occur fault, and fault existed before
	dealFault(constant.AssertionOccur, testNodeName1, faultKey1, faultCache)
	convey.So(publicfault.PubFaultCache.GetPubFaultNum(), convey.ShouldEqual, 1)
}

func testFaultExisted() {
	resetFaultCache()
	resetQueueCache()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		PubFaultNeedDelete.DealDelete(ctx)
	}()
	defer cancel()

	// fault existed ->
	publicfault.PubFaultCache.AddPubFaultToCache(faultCache, testNodeName1, faultKey1)
	convey.So(publicfault.PubFaultCache.GetPubFaultNum(), convey.ShouldEqual, 1)
	time.Sleep(diffTime)

	// recover fault, and fault existed before ->
	dealFault(constant.AssertionRecover, testNodeName1, faultKey1, faultCache)
	// will be deleted right now
	convey.So(publicfault.PubFaultCache.GetPubFaultNum(), convey.ShouldEqual, 0)
}

func testRecoverOccur() {
	resetFaultCache()
	resetQueueCache()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		PubFaultNeedDelete.DealDelete(ctx)
	}()
	defer cancel()

	// fault recover ->
	dealFault(constant.AssertionRecover, testNodeName1, faultKey1, faultCache)
	// add to needDeleteQueue ->
	convey.So(PubFaultNeedDelete.Len(), convey.ShouldEqual, 1)
	time.Sleep(sec1)
	// fault occur ->
	dealFault(constant.AssertionOccur, testNodeName1, faultKey1, faultCache)
	// add to PubFaultCache ->
	convey.So(publicfault.PubFaultCache.GetPubFaultNum(), convey.ShouldEqual, 1)
	// will be deleted after 5s
	time.Sleep(diffTime)
	convey.So(publicfault.PubFaultCache.GetPubFaultNum(), convey.ShouldEqual, 0)
}

func testOccurRecover() {
	resetFaultCache()
	resetQueueCache()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		PubFaultNeedDelete.DealDelete(ctx)
	}()
	defer cancel()

	// fault occur ->
	dealFault(constant.AssertionOccur, testNodeName1, faultKey1, faultCache)
	// add to PubFaultCache ->
	convey.So(publicfault.PubFaultCache.GetPubFaultNum(), convey.ShouldEqual, 1)
	// fault recover ->
	dealFault(constant.AssertionRecover, testNodeName1, faultKey1, faultCache)

	// after 1s, fault still exists ->
	time.Sleep(sec1)
	convey.So(publicfault.PubFaultCache.GetPubFaultNum(), convey.ShouldEqual, 1)
	// will be deleted after 5s
	time.Sleep(diffTime)
	convey.So(publicfault.PubFaultCache.GetPubFaultNum(), convey.ShouldEqual, 0)
}

func testOnce() {
	resetFaultCache()
	resetQueueCache()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		PubFaultNeedDelete.DealDelete(ctx)
	}()
	defer cancel()

	// fault once ->
	dealFault(constant.AssertionOnce, testNodeName1, faultKey1, faultCache)
	// add to PubFaultCache ->
	convey.So(publicfault.PubFaultCache.GetPubFaultNum(), convey.ShouldEqual, 1)

	// after 1s, fault still exists ->
	time.Sleep(sec1)
	convey.So(publicfault.PubFaultCache.GetPubFaultNum(), convey.ShouldEqual, 1)
	// will be deleted after 5s
	time.Sleep(diffTime)
	convey.So(publicfault.PubFaultCache.GetPubFaultNum(), convey.ShouldEqual, 0)
}

func resetFaultCache() {
	for nodeName := range publicfault.PubFaultCache.GetPubFault() {
		delete(publicfault.PubFaultCache.GetPubFault(), nodeName)
	}
}
