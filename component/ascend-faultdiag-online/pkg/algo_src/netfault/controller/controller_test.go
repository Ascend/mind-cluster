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

// Package controller
package controller

import (
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

func TestStartController(t *testing.T) {
	convey.Convey("TestStartController", t, func() {
		convey.Convey("empty path return", func() {
			clusterPath := `/cluster`
			startController(clusterPath)
		})
		convey.Convey("path not exist return", func() {
			clusterPath := `/tmp/clusterxxx`
			startController(clusterPath)
		})
		convey.Convey("path not exist", func() {
			patch0 := gomonkey.ApplyFuncReturn(os.Stat, nil, nil)
			defer patch0.Reset()
			patch1 := gomonkey.ApplyFunc(startSuperPodsDetectionAsync, func(path string) {
				return
			})
			defer patch1.Reset()
			startController("/tmp")
		})
	})
}

func TestStopController(t *testing.T) {
	convey.Convey("TestStopController", t, func() {
		convey.Convey("no parameters", func() {
			patch0 := gomonkey.ApplyMethod(reflect.TypeOf(controllerExitCond), "Wait", func(_ *sync.Cond) {
				return
			})
			defer patch0.Reset()
			stopController()
		})
	})
}

func TestReloadController(t *testing.T) {
	convey.Convey("TestReloadController", t, func() {
		convey.Convey("patch stop", func() {
			patch0 := gomonkey.ApplyFunc(stopController, func() {
				return
			})
			defer patch0.Reset()
			reloadController("/cluster")
		})
	})
}
