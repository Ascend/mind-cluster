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

// Package utils provides some common utils
package utils

import (
	"os"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"ascend-faultdiag-online/pkg/algo_src/slownode/parse/common/constants"
)

const (
	testDir                  = "testdata"
	fileMode0755 os.FileMode = 0755
	fileMode0644 os.FileMode = 0644
)

func TestFileValidator(t *testing.T) {
	if err := os.MkdirAll(testDir, fileMode0755); err != nil {
		t.Fatalf("create temp test dir failed: %v", err)
	}
	defer os.RemoveAll(testDir)

	convey.Convey("test FileValidator func", t, func() {
		testValidFileExactlyMaxSize()
		testValidFileSmallerThanMaxSize()
		testInvalidFileExceedsMaxSize()
		testEmptyFile()
		testNonExistentFile()
		testDirectoryInsteadOfFile()
	})
}

func testValidFileExactlyMaxSize() {
	convey.Convey("file size equals to the max size", func() {
		content := make([]byte, constants.FileMaxSize)
		filePath := createTestFile("valid_exact_size.txt", content)

		convey.Convey("pass", func() {
			err := FileValidator(filePath)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func testValidFileSmallerThanMaxSize() {
	convey.Convey("file size is less than the max size", func() {
		content := make([]byte, constants.FileMaxSize-1)
		filePath := createTestFile("valid_small.txt", content)

		convey.Convey("pass", func() {
			err := FileValidator(filePath)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func testInvalidFileExceedsMaxSize() {
	convey.Convey("file size is greater than the max size", func() {
		content := make([]byte, constants.FileMaxSize+1)
		filePath := createTestFile("invalid_large.txt", content)

		convey.Convey("got error", func() {
			err := FileValidator(filePath)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "file size:")
		})
	})
}

func testEmptyFile() {
	convey.Convey("file is empty", func() {
		filePath := createTestFile("empty.txt", []byte{})

		convey.Convey("pass", func() {
			err := FileValidator(filePath)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func testNonExistentFile() {
	convey.Convey("file does not exist", func() {
		filePath := "nonexistent_file.txt"

		convey.Convey("got error", func() {
			err := FileValidator(filePath)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "no such file or directory")
		})
	})
}

func testDirectoryInsteadOfFile() {
	convey.Convey("file path is a directory not file", func() {
		dirPath := testDir + "/test_dir"
		convey.So(os.Mkdir(dirPath, fileMode0755), convey.ShouldBeNil)

		convey.Convey("got error", func() {
			err := FileValidator(dirPath)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "is a directory")
		})
	})
}

func createTestFile(name string, content []byte) string {
	filePath := testDir + "/" + name
	convey.So(os.WriteFile(filePath, content, fileMode0644), convey.ShouldBeNil)
	return filePath
}
