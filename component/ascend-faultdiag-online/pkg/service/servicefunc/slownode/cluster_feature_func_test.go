/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, convey.Software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package slownode a DT collection for slownode cluster feature func
package slownode

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

const (
	testDir      = "testdata"
	testFileName = "testfile.json"
)

func TestClusterWriteFile(t *testing.T) {
	convey.Convey("test clusterWriteFile func", t, func() {
		// make sure the test directory is clean before running tests
		convey.So(os.RemoveAll(testDir), convey.ShouldBeNil)
		defer os.RemoveAll(testDir)

		testData := map[string]any{
			"key": "value",
			"num": 123,
		}

		testCreateDirectoryAndWriteFile(testData)
		testDirectoryAlreadyExists(testData)
		testSymlinkPath(testData)
		testJsonMarshalFailure()
	})
}

func testCreateDirectoryAndWriteFile(testData map[string]any) {
	convey.Convey("dir is not existed", func() {
		targetDir := filepath.Join(testDir, "newdir")
		targetFile := filepath.Join(targetDir, testFileName)

		convey.Convey("create file and write successfully", func() {
			clusterWriteFile(targetDir, testFileName, testData)

			_, err := os.Stat(targetDir)
			convey.So(err, convey.ShouldBeNil)

			_, err = os.Stat(targetFile)
			convey.So(err, convey.ShouldBeNil)

			content, err := os.ReadFile(targetFile)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(content), convey.ShouldBeGreaterThan, 0)
		})
	})
}

func testDirectoryAlreadyExists(testData map[string]any) {
	convey.Convey("dir exists", func() {
		targetDir := filepath.Join(testDir, "existingdir")
		targetFile := filepath.Join(targetDir, testFileName)

		// pre-create the directory
		convey.So(os.MkdirAll(targetDir, os.ModePerm), convey.ShouldBeNil)

		convey.Convey("write data", func() {
			clusterWriteFile(targetDir, testFileName, testData)

			// verify the file exists and is written correctly
			_, err := os.Stat(targetFile)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func testSymlinkPath(testData map[string]any) {
	convey.Convey("file is symlink", func() {
		targetDir := filepath.Join(testDir, "realdir")
		symlinkDir := filepath.Join(testDir, "symlinkdir")

		// create the target directory and symlink
		convey.So(os.MkdirAll(targetDir, os.ModePerm), convey.ShouldBeNil)
		convey.So(os.Symlink(targetDir, symlinkDir), convey.ShouldBeNil)

		convey.Convey("write failed", func() {
			clusterWriteFile(symlinkDir, testFileName, testData)
			_, err := os.Stat(filepath.Join(symlinkDir, testFileName))
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(os.IsNotExist(err), convey.ShouldBeTrue)
		})
	})
}

func testJsonMarshalFailure() {
	convey.Convey("testfile.json marshal failed", func() {
		invalidData := map[string]any{
			"channel": make(chan int), // channels cannot be marshaled to JSON
		}

		convey.Convey("write failed", func() {

			targetDir := filepath.Join(testDir, "marshalerrordir")
			clusterWriteFile(targetDir, testFileName, invalidData)
			_, err := os.Stat(targetDir)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(os.IsNotExist(err), convey.ShouldBeTrue)
		})
	})
}
