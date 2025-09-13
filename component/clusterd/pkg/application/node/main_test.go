// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package node main test for node
package node

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"ascend-common/common-utils/hwlog"
)

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		panic(err)
	}
	code := m.Run()
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
