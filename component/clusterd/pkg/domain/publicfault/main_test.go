// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault main test for public fault
package publicfault

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
)

const (
	testFilePath = "./testCfg.json"

	testId        = "11937763019444715778"
	testTimeStamp = 1739866717
	testResource1 = "resource1"
	testResource2 = "resource2"
	testFaultCode = "000000001"
	testNodeName1 = "node1"
	testNodeName2 = "node2"
)

var (
	testErr = errors.New("test error")

	testFaultInfo = api.PubFaultInfo{
		Id:        testId,
		TimeStamp: testTimeStamp,
		Version:   "v1.0",
		Resource:  testResource1,
		Faults:    []api.Fault{testFault},
	}
	testFault = api.Fault{
		FaultId:       testId,
		FaultType:     constant.FaultTypeNPU,
		FaultCode:     testFaultCode,
		FaultTime:     testTimeStamp,
		Assertion:     constant.AssertionOccur,
		FaultLocation: nil,
		Influence:     []api.Influence{testInfluence},
		Description:   "fault description",
	}
	testInfluence = api.Influence{
		NodeName:  testNodeName1,
		DeviceIds: []int32{0},
	}

	testCacheData = constant.PubFaultCache{
		FaultDevIds: []int32{0, 1},
		FaultId:     testId,
		FaultType:   constant.FaultTypeNPU,
		FaultCode:   testFaultCode,
		FaultTime:   testTimeStamp,
		Assertion:   constant.AssertionOccur,
	}
)

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		return
	}
	code := m.Run()
	teardown()
	fmt.Printf("exit_xode = %v\n", code)
}

func setup() error {
	return initLog()
}

func initLog() error {
	logConfig := &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := hwlog.InitRunLogger(logConfig, context.Background()); err != nil {
		fmt.Printf("init hwlog failed, %v\n", err)
		return errors.New("init hwlog failed")
	}
	return nil
}

func teardown() {
	err := os.Remove(testFilePath)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		fmt.Printf("remove file %s failed, %v\n", testFilePath, err)
	}
}
