// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault test for public fault service
package publicfault

import (
	"context"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/interface/grpc/pubfault"
)

var (
	pubFaultSvc *PubFaultService
	req         *pubfault.PublicFaultRequest
)

const (
	testId       = "11937763019444715778"
	testResource = "resource1"
	testNodeName = "node1"
)

func fakeService() *PubFaultService {
	ctx := context.Background()
	return NewPubFaultService(ctx)
}

func TestSendPublicFault(t *testing.T) {
	pubFaultSvc = fakeService()
	influence := &pubfault.PubFaultInfo{
		NodeName:  testNodeName,
		DeviceIds: []int32{0, 1},
	}
	fault := &pubfault.Fault{
		FaultId:       testId,
		FaultType:     constant.FaultTypeNPU,
		FaultCode:     testFaultCode,
		FaultTime:     testTimeStamp,
		Assertion:     constant.AssertionOccur,
		FaultLocation: nil,
		Influence:     []*pubfault.PubFaultInfo{influence},
		Description:   "",
	}
	req = &pubfault.PublicFaultRequest{
		Id:        testId,
		Timestamp: testTimeStamp,
		Version:   "1.0",
		Resource:  testResource,
		Faults:    []*pubfault.Fault{fault},
	}
	convey.Convey("test method 'SendPublicFault' success", t, testSendFault)
	convey.Convey("test method 'SendPublicFault' failed, req is nil", t, testSendFaultNilReq)
	convey.Convey("test method 'SendPublicFault' failed, limit error", t, testSendFaultErrLimit)
	convey.Convey("test method 'SendPublicFault' failed, check error", t, testSendFaultErrCheck)
}

func testSendFault() {
	p1 := gomonkey.ApplyFuncReturn(PubFaultCollector, nil)
	defer p1.Reset()
	resp, err := pubFaultSvc.SendPublicFault(context.Background(), req)
	expInfo := "public fault send successfully"
	convey.So(err, convey.ShouldBeNil)
	convey.So(resp.Code, convey.ShouldEqual, int32(common.OK))
	convey.So(resp.Info, convey.ShouldResemble, expInfo)
}

func testSendFaultNilReq() {
	resp, err := pubFaultSvc.SendPublicFault(context.Background(), nil)
	expInfo := "req is nil"
	convey.So(err, convey.ShouldBeNil)
	convey.So(resp.Code, convey.ShouldEqual, int32(common.InvalidReqParam))
	convey.So(resp.Info, convey.ShouldResemble, expInfo)
}

func testSendFaultErrLimit() {
	limitErr := errors.New("limiter work by resource failed")
	p1 := gomonkey.ApplyFuncReturn(PubFaultCollector, limitErr)
	defer p1.Reset()
	resp, err := pubFaultSvc.SendPublicFault(context.Background(), req)
	expInfo := "limiter work by resource failed"
	convey.So(err, convey.ShouldBeNil)
	convey.So(resp.Code, convey.ShouldEqual, int32(common.InvalidReqRate))
	convey.So(resp.Info, convey.ShouldResemble, expInfo)
}

func testSendFaultErrCheck() {
	p1 := gomonkey.ApplyFuncReturn(PubFaultCollector, testErr)
	defer p1.Reset()
	resp, err := pubFaultSvc.SendPublicFault(context.Background(), req)
	convey.So(err, convey.ShouldBeNil)
	convey.So(resp.Code, convey.ShouldEqual, int32(common.InvalidReqParam))
	convey.So(resp.Info, convey.ShouldResemble, testErr.Error())
}
