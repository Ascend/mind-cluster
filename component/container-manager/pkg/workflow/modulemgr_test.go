/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package workflow test for interface Module
package workflow

import (
	"context"
	"errors"
	"fmt"
	"syscall"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
)

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		return
	}
	code := m.Run()
	fmt.Printf("exit_code = %v\n", code)
}

func setup() error {
	if err := initLog(); err != nil {
		return err
	}
	newMockModule()
	return nil
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

var mockModule = &MockModule{}

// MockModule mock module
type MockModule struct {
	workCalled     bool
	shutdownCalled bool
	workDuration   time.Duration
}

const workDuration = 100 * time.Millisecond

func newMockModule() *MockModule {
	return &MockModule{
		workDuration: workDuration,
	}
}

// Name module name
func (mm *MockModule) Name() string {
	return "mock module"
}

// Init module init
func (mm *MockModule) Init() error {
	hwlog.RunLog.Infof("init module <%s> success", mm.Name())
	return nil
}

// Work module work
func (mm *MockModule) Work(ctx context.Context) {
	mm.workCalled = true
	select {
	case <-time.After(mm.workDuration):
	case <-ctx.Done():
	}
}

// ShutDown module shutdown
func (mm *MockModule) ShutDown() {
	mm.shutdownCalled = true
}

func TestModuleMgr(t *testing.T) {
	const (
		waitWorkExecTime = 10 * time.Millisecond
		waitShutdownTime = 50 * time.Millisecond
	)
	moduleMgr := NewModuleMgr()
	convey.Convey("test method 'Register' success", t, func() {
		moduleMgr.Register(mockModule)
		convey.So(len(moduleMgr.modules), convey.ShouldEqual, 1)
	})
	convey.Convey("test method 'Init' success", t, func() {
		err := moduleMgr.Init()
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("test method 'Work' success", t, func() {
		moduleMgr.Work(context.Background())
		time.Sleep(waitWorkExecTime)
		if !mockModule.workCalled {
			t.Errorf("mockModule Work() was not called")
		}
	})
	convey.Convey("test method 'ShutDown' success", t, func() {
		moduleMgr.Register(mockModule)
		go moduleMgr.ShutDown()
		time.Sleep(waitWorkExecTime)
		err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		if err != nil {
			t.Errorf("kill process failed, %v", err)
		}
		time.Sleep(waitShutdownTime)
		if !mockModule.shutdownCalled {
			t.Errorf("mockModule ShutDown() was not called")
		}
	})
}
