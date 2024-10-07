/* Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package common a series of common function
package common

import (
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

// TestWriteToFile test of WriteToFile
func TestWriteToFile(t *testing.T) {
	convey.Convey("Test WriteToFile", t, func() {
		convey.Convey("Test WriteToFile success", func() {
			path := ResetInfoDir + "/default.test/reset.json"
			crErr := WriteToFile("test", path)
			convey.So(crErr, convey.ShouldBeNil)
			path1 := ResetInfoDir + "/default.test/restartType"
			crErr1 := WriteToFile("test", path1)
			convey.So(crErr1, convey.ShouldBeNil)
			rmErr := RemoveFileAndDir("default", "test")
			convey.So(rmErr, convey.ShouldBeNil)
		})
	})
}

// TestGenResetDirName test of GenResetDirName
func TestGenResetDirName(t *testing.T) {
	convey.Convey("Test GenResetDirName", t, func() {
		convey.Convey("Test GenResetDirName success", func() {
			name := GenResetDirName("default", "test")
			convey.ShouldEqual(name, ResetInfoDir+"/default.test")
		})
	})
}

// TestGenResetFileName test of GenResetFileName
func TestGenResetFileName(t *testing.T) {
	convey.Convey("Test GenResetFileName", t, func() {
		convey.Convey("Test GenResetFileName success", func() {
			name := GenResetFileName("default", "test")
			convey.ShouldEqual(name, ResetInfoDir+"/default.test/reset.json")
		})
	})
}

// TestGenResetTypeFileName test of GenResetTypeFileName
func TestGenResetTypeFileName(t *testing.T) {
	convey.Convey("Test GenResetTypeFileName", t, func() {
		convey.Convey("Test GenResetTypeFileName success", func() {
			name := GenResetTypeFileName("default", "test")
			convey.ShouldEqual(name, ResetInfoDir+"/default.test/resetType")
		})
	})
}
