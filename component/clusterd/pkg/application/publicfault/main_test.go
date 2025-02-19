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

var testErr = errors.New("test error")

const testFilePath = "./testCfg.json"

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
