// Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.

// Package logs test for common func about logs
package logs

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
)

var errTest = errors.New("test error")

const logFile = "./clusterd.log"

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		return
	}
	code := m.Run()
	teardown()
	fmt.Printf("exit_code = %v\n", code)
}

func setup() error {
	return initLog()
}

func initLog() error {
	logConfig := &hwlog.LogConfig{
		LogFileName: logFile,
		LogLevel:    0,
		MaxBackups:  2,
		FileMaxSize: 2,
		MaxAge:      365,
	}
	if err := hwlog.InitRunLogger(logConfig, context.Background()); err != nil {
		fmt.Printf("init hwlog failed, %v\n", err)
		return errors.New("init hwlog failed")
	}
	return nil
}

func teardown() {
	if err := os.Remove(logFile); err != nil {
		fmt.Printf("remove log file failed, %v\n", err)
		return
	}
}

func TestInitJobEventLogger(t *testing.T) {
	convey.Convey("test InitJobEventLogger failed, NewCustomLogger error", t, func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		p1 := gomonkey.ApplyFuncReturn(hwlog.NewCustomLogger, nil, errTest)
		defer p1.Reset()
		err := InitJobEventLogger(ctx)
		convey.So(err, convey.ShouldResemble, errTest)
		convey.So(JobEventLog, convey.ShouldBeNil)
	})

	convey.Convey("test InitJobEventLogger success", t, func() {
		mockCustomLog := &hwlog.CustomLogger{}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		p2 := gomonkey.ApplyFuncReturn(hwlog.NewCustomLogger, mockCustomLog, nil)
		defer p2.Reset()
		err := InitJobEventLogger(ctx)
		convey.So(err, convey.ShouldBeNil)
		convey.So(JobEventLog, convey.ShouldResemble, mockCustomLog)
	})
}

func TestRecordLog(t *testing.T) {
	const (
		user  = "user1"
		event = "modify data"
	)
	convey.Convey("test RecordLog, result is success", t, func() {
		RecordLog(user, event, constant.Success)
		log := getLastLine(t)
		convey.So(log, convey.ShouldContainSubstring, fmt.Sprintf("role[%s] %s %s", user, event, constant.Success))
	})
	convey.Convey("test RecordLog, result is failed", t, func() {
		RecordLog(user, event, constant.Failed)
		log := getLastLine(t)
		convey.So(log, convey.ShouldContainSubstring, fmt.Sprintf("role[%s] %s %s", user, event, constant.Failed))
	})
}

func getLastLine(t *testing.T) string {
	file, err := os.Open(logFile)
	if err != nil {
		t.Errorf("open log file failed, error: %v", err)
	}
	defer func(file *os.File) {
		if err = file.Close(); err != nil {
			t.Errorf("close log file failed, error: %v", err)
		}
	}(file)

	var lastLine string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lastLine = scanner.Text()
	}
	if err = scanner.Err(); err != nil {
		t.Errorf("scan file failed, error: %v", err)
	}
	return lastLine
}

func TestInitGrpcEventLogger(t *testing.T) {
	convey.Convey("test InitGrpcEventLogger failed, NewCustomLogger error", t, func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		p1 := gomonkey.ApplyFuncReturn(hwlog.NewCustomLogger, nil, errTest)
		defer p1.Reset()
		err := InitGrpcEventLogger(ctx)
		convey.So(err, convey.ShouldResemble, errTest)
		convey.So(GrpcEventLogger, convey.ShouldBeNil)
	})

	convey.Convey("test InitGrpcEventLogger success", t, func() {
		mockCustomLog := &hwlog.CustomLogger{}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		p2 := gomonkey.ApplyFuncReturn(hwlog.NewCustomLogger, mockCustomLog, nil)
		defer p2.Reset()
		err := InitGrpcEventLogger(ctx)
		convey.So(err, convey.ShouldBeNil)
		convey.So(GrpcEventLogger, convey.ShouldResemble, mockCustomLog)
	})
}
