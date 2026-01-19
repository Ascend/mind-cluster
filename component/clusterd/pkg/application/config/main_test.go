// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package config main test for config
package config

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/logs"
)

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		return
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

	logger, err := hwlog.NewCustomLogger(logConfig, context.Background())
	if err != nil {
		hwlog.RunLog.Errorf("GrpcEventLog init failed, error is %v", err)
		return errors.New("init logs failed")
	}
	logs.GrpcEventLogger = logger
	return nil
}
