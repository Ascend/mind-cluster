// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package om a series of service function
package om

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"ascend-common/common-utils/hwlog"
)

func TestMain(m *testing.M) {
	if err := initLog(); err != nil {
		return
	}
	code := m.Run()
	fmt.Printf("exit_code = %v\n", code)
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
