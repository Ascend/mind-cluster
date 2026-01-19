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

// Package constant a package for constant
package constant

import (
	"context"
	"fmt"
	"testing"

	"ascend-common/common-utils/hwlog"
	"github.com/smartystreets/goconvey/convey"
)

const (
	someString = "qwertyui"
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
		return err
	}
	return nil
}

func TestNewProfilingExecRes(t *testing.T) {
	res := NewProfilingExecRes(On)
	convey.ShouldEqual(res.status, On)
	res = NewProfilingExecRes(Off)
	convey.ShouldEqual(res.status, Off)
	res = NewProfilingExecRes(Unknown)
	convey.ShouldEqual(res.status, Unknown)
	res = NewProfilingExecRes(Exp)
	convey.ShouldEqual(res.status, Exp)

	res = NewProfilingExecRes(someString)
	convey.ShouldEqual(res.status, Unknown)
}
