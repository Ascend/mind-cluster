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

// Package config is used for file reading and writing, as well as data processing.
package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/algo_src/slownode/jobdetectionmanager"
)

const (
	testDir       = "testdata"
	testFileName  = "testfile.txt"
	logLineLength = 256
)

func init() {
	config := hwlog.LogConfig{
		OnlyToStdout:  true,
		MaxLineLength: logLineLength,
	}
	err := hwlog.InitRunLogger(&config, context.TODO())
	if err != nil {
		fmt.Println(err)
	}
}

func TestCheckExistDirectoryOrFile(t *testing.T) {
	var fileMode0644 os.FileMode = 0644
	var fileMode0755 os.FileMode = 0755
	// create test dir and file
	err := os.MkdirAll(testDir, fileMode0755)
	assert.Nil(t, err)
	testFilePath := filepath.Join(testDir, testFileName)
	err = os.WriteFile(testFilePath, []byte("test"), fileMode0644)
	assert.Nil(t, err)
	defer os.RemoveAll(testDir)

	patches := gomonkey.ApplyFunc(jobdetectionmanager.GetDetectionLoopStatusClusterLevel, func(string) bool { return true })
	defer patches.Reset()
	patches.ApplyFunc(jobdetectionmanager.GetDetectionLoopStatusNodeLevel, func(string) bool { return true })

	convey.Convey("test CheckExistDirectoryOrFile func", t, func() {
		testFileExists(testFilePath)
		testDirectoryExists(testDir)
		testSymlinkDetection(testFilePath)
	})
}

func testFileExists(testFilePath string) {
	convey.Convey("file exists", func() {
		convey.Convey("fileOrDir=false, got result: true", func() {
			result := CheckExistDirectoryOrFile(testFilePath, false, "cluster", "testJob")
			convey.So(result, convey.ShouldBeTrue)
		})

		convey.Convey("fileOrDir=true got result: false", func() {
			result := CheckExistDirectoryOrFile(testFilePath, true, "cluster", "testJob")
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func testDirectoryExists(testDir string) {
	convey.Convey("dir exists", func() {
		convey.Convey("fileOrDir=true, got result: true", func() {
			result := CheckExistDirectoryOrFile(testDir, true, "node", "testJob")
			convey.So(result, convey.ShouldBeTrue)
		})

		convey.Convey("fileOrDir=false, got result: false", func() {
			result := CheckExistDirectoryOrFile(testDir, false, "node", "testJob")
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func testSymlinkDetection(testFilePath string) {
	convey.Convey("test symlink", func() {
		symlinkPath := filepath.Join(testDir, "symlink.txt")
		convey.So(os.Symlink(testFilePath, symlinkPath), convey.ShouldBeNil)
		defer os.Remove(symlinkPath)

		convey.Convey("fileOrDir=false, got result: true", func() {
			result := CheckExistDirectoryOrFile(symlinkPath, false, "cluster", "testJob")
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}
