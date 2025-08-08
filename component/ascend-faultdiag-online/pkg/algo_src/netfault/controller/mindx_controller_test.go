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
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

func TestStart(t *testing.T) {
	convey.Convey("TestStart", t, func() {
		convey.Convey("invalid input", func() {
			Start()
		})
	})
}

func TestReload(t *testing.T) {
	convey.Convey("TestReload", t, func() {
		convey.Convey("reload", func() {
			patch := gomonkey.ApplyFunc(reloadController, func(path string) {
				return
			})
			defer patch.Reset()
			Reload()
		})
	})
}

func TestStop(t *testing.T) {
	convey.Convey("stop", t, func() {
		convey.Convey("invalid input", func() {
			patch := gomonkey.ApplyFunc(stopController, func() {
				return
			})
			defer patch.Reset()
			Stop()
		})
	})
}
