// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault main test for public fault
package publicfault

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"ascend-common/common-utils/hwlog"
)

var (
	testErr = errors.New("test error")

	faultKey1 string
	faultKey2 string
	faultKey3 string
	faultKey4 string
)

const (
	testFilePath = "./testCfg.json"

	testId1         = "11937763019253715778"
	testId2         = "11937763019253715800"
	testTimeStamp   = 1739866717000
	testResource1   = "resource1"
	testResource2   = "resource2"
	testFaultCode   = "000000001"
	testNodeName1   = "node1"
	testNodeName2   = "node2"
	testNodeSN1     = "node1-sn"
	testNodeSN2     = "node2-sn"
	testDescription = "fault description"
)

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		return
	}
	faultKey1 = testResource1 + testId1
	faultKey2 = testResource2 + testId2
	faultKey3 = testResource1 + testId2
	faultKey4 = testResource2 + testId1
	code := m.Run()
	teardown()
	fmt.Printf("exit_code = %v\n", code)
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
