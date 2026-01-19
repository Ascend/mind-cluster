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
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

const (
	count0 = 0
	count1 = 1
)

func createSymbolicLink(t *testing.T) (originPath, symLinkPath string) {
	var fileMode0755 os.FileMode = 0755
	tmpDir := t.TempDir()
	originalPath := filepath.Join(tmpDir, "origin")
	err := os.MkdirAll(originalPath, fileMode0755)
	assert.Nil(t, err)
	symlinkPath := filepath.Join(tmpDir, "symbolic")
	// create a symlink
	err = os.Symlink(originalPath, symlinkPath)
	assert.Nil(t, err)
	return originalPath, symlinkPath
}

func TestStart(t *testing.T) {
	callStartCount := count0
	patch := gomonkey.ApplyFunc(startController, func(path string) {
		callStartCount++
	})
	defer patch.Reset()
	convey.Convey("TestStart", t, func() {
		convey.Convey("invalid path", func() {
			_, symlinkPath := createSymbolicLink(t)
			varPatch := gomonkey.ApplyGlobalVar(&clusterLevelPath, symlinkPath)
			Start()
			convey.So(callStartCount, convey.ShouldEqual, count0)
			varPatch.Reset()
		})
		convey.Convey("invalid input", func() {
			callStartCount = count0
			Start()
			convey.So(callStartCount, convey.ShouldEqual, count1)
		})
	})
}

func TestReload(t *testing.T) {
	callStartCount := count0
	patch := gomonkey.ApplyFunc(reloadController, func(path string) {
		callStartCount++
	})
	defer patch.Reset()
	convey.Convey("TestReload", t, func() {
		convey.Convey("invalid path", func() {
			_, symlinkPath := createSymbolicLink(t)
			varPatch := gomonkey.ApplyGlobalVar(&clusterLevelPath, symlinkPath)
			Reload()
			convey.So(callStartCount, convey.ShouldEqual, count0)
			varPatch.Reset()
		})
		convey.Convey("reload", func() {
			callStartCount = count0
			Reload()
			convey.So(callStartCount, convey.ShouldEqual, count1)
		})
	})
}

func TestStop(t *testing.T) {
	convey.Convey("stop", t, func() {
		convey.Convey("invalid input", func() {
			patch := gomonkey.ApplyFunc(stopController, func() {})
			defer patch.Reset()
			Stop()
		})
	})
}

func TestRegisterDetectionCallback(t *testing.T) {
	convey.Convey("Test RegisterDetectionCallback", t, func() {
		convey.Convey("should return when input is nil", func() {
			RegisterDetectionCallback(nil)
			convey.So(callbackFunc, convey.ShouldBeNil)
		})

		convey.Convey("should set callbackFunc when input is valid", func() {
			var callCount int
			callback := func(string) {
				callCount++
			}
			RegisterDetectionCallback(callback)
			convey.So(callbackFunc, convey.ShouldNotBeNil)
			callbackFunc("test")
			convey.So(callCount, convey.ShouldEqual, count1)
		})
	})
}
