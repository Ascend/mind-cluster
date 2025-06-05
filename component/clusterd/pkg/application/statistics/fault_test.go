// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package statistics test for statistic funcs about fault
package statistics

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/publicfault"
	"clusterd/pkg/domain/statistics"
	"clusterd/pkg/interface/kube"
)

func TestStatisticFault(t *testing.T) {
	convey.Convey("test StatisticFault method 'Notify'", t, testNotify)
	convey.Convey("test StatisticFault method 'UpdateFault'", t, testUpdateFault)
	convey.Convey("test StatisticFault method 'LoadFaultData'", t, testLoadFaultData)
}

func testNotify() {
	StatisticFault.Notify()
	convey.So(len(StatisticFault.updateChan), convey.ShouldEqual, 1)
}

func testUpdateFault() {
	const waitGoroutineFinishedTime = 200 * time.Millisecond

	// ctx stop
	StatisticFault.updateChan = make(chan struct{}, 1)
	ctx, cancel := context.WithCancel(context.Background())
	haveStopped := atomic.Bool{}
	go func() {
		StatisticFault.UpdateFault(ctx)
		haveStopped.Store(true)
	}()
	cancel()
	time.Sleep(waitGoroutineFinishedTime)
	convey.So(haveStopped.Load(), convey.ShouldBeTrue)

	// notify updateChan
	StatisticFault.Notify()
	p1 := gomonkey.ApplyFuncReturn(statistics.UpdateFaultToCM, testErr)
	defer p1.Reset()

	ctx, cancel = context.WithCancel(context.Background())
	go func() {
		StatisticFault.UpdateFault(ctx)
	}()
	time.Sleep(waitGoroutineFinishedTime)
	cancel()
	time.Sleep(waitGoroutineFinishedTime)
	convey.So(len(StatisticFault.updateChan), convey.ShouldEqual, 0)
}

func testLoadFaultData() {
	const (
		testFaultId   = "12345"
		testFaultCode = "123456789"
		testFaultTime = 1234567890
		testNodeName  = "node1"
		testCMName    = "test-cm-name"
		testCMNS      = "default"
	)
	cmData := make(map[string]string, 1)
	faults := make(map[string][]constant.NodeFault)
	nodeFault := constant.NodeFault{
		FaultResource: constant.PublicFaultType,
		FaultDevIds:   []int32{0},
		FaultId:       testFaultId,
		FaultType:     constant.FaultTypeStorage,
		FaultCode:     testFaultCode,
		FaultLevel:    constant.SeparateNPU,
		FaultTime:     testFaultTime,
	}
	faults[testNodeName] = []constant.NodeFault{nodeFault}
	cmData[constant.StatisticPubFaultKey] = util.ObjToString(faults)
	cm := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testCMName,
			Namespace: testCMNS,
		},
		Data: cmData,
	}

	p1 := gomonkey.ApplyFuncReturn(kube.GetConfigMap, cm, nil)
	defer p1.Reset()
	StatisticFault.LoadFaultData()
	convey.So(len(publicfault.PubFaultCache.GetPubFault()), convey.ShouldEqual, 1)
}
