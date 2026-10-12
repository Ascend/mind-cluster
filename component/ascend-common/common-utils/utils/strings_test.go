/* Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package utils provides the util func
package utils

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

const byteLength = 32

func TestReplacePrefix(t *testing.T) {
	convey.Convey("relative path", t, func() {
		path := ReplacePrefix("./testdata/cert/ca.crt", "****")
		convey.So(path, convey.ShouldEqual, "****testdata/cert/ca.crt")
	})
	convey.Convey("abconvey.Solute path", t, func() {
		path := ReplacePrefix("/testdata/cert/ca.crt", "****")
		convey.So(path, convey.ShouldEqual, "****estdata/cert/ca.crt")
	})
	convey.Convey("path length less than 2", t, func() {
		path := ReplacePrefix("/", "****")
		convey.So(path, convey.ShouldEqual, "****")
	})
	convey.Convey("empty string", t, func() {
		path := ReplacePrefix("", "****")
		convey.So(path, convey.ShouldEqual, "****")
	})

}

func TestMaskPrefix(t *testing.T) {
	convey.Convey("relative path", t, func() {
		path := MaskPrefix("./testdata/cert/ca.crt")
		convey.So(path, convey.ShouldEqual, "****testdata/cert/ca.crt")
	})
	convey.Convey("abconvey.Solute path", t, func() {
		path := MaskPrefix("/testdata/cert/ca.crt")
		convey.So(path, convey.ShouldEqual, "****estdata/cert/ca.crt")
	})
	convey.Convey("path length less than 2", t, func() {
		path := MaskPrefix("/")
		convey.So(path, convey.ShouldEqual, "****")
	})
	convey.Convey("empty string", t, func() {
		path := MaskPrefix("")
		convey.So(path, convey.ShouldEqual, "****")
	})

}

func TestGetSha256Code(t *testing.T) {
	convey.Convey("test sha256", t, func() {
		hashes := GetSha256Code([]byte("this is a test sentence"))
		convey.So(len(hashes), convey.ShouldEqual, byteLength)
	})
}

func TestIsDigitString(t *testing.T) {
	convey.Convey("test IsDigitString", t, func() {
		convey.Convey("case IsDigitString is true", func() {
			str := "123"
			convey.ShouldBeTrue(IsDigitString(str))
		})
		convey.Convey("case IsDigitString is false", func() {
			str := "123a"
			convey.ShouldBeFalse(IsDigitString(str))
		})
	})
}

// TestIsNPUResourceName test IsNPUResourceName
func TestIsNPUResourceName(t *testing.T) {
	convey.Convey("IsNPUResourceName", t, func() {
		convey.Convey("01-npu resource names should return true", func() {
			npuNames := []string{
				"huawei.com/npu",
				"huawei.com/npu-core",
				"huawei.com/Ascend910",
				"huawei.com/Ascend310",
				"huawei.com/Ascend310P",
				"huawei.com/Ascend910-2c",
				"huawei.com/Ascend910-16c",
			}
			for _, name := range npuNames {
				convey.So(IsNPUResourceName(name), convey.ShouldBeTrue)
			}
		})
		convey.Convey("02-same prefix non-npu resource names should return false", func() {
			nonNPUNames := []string{
				"huawei.com/dpu.resource.name",
				"huawei.com/ub_rdma_1",
				"cpu",
				"memory",
				"huawei.com/",
				"",
				"huawei.com/Ascend910B",
				"huawei.com/Ascend910A3",
				"huawei.com/Ascend910X",
				"huawei.com/Ascend910B-4c",
				"huawei.com/Ascend310-2c",
				"huawei.com/Ascend910-c",
				"huawei.com/Ascend910-2cpu",
			}
			for _, name := range nonNPUNames {
				convey.So(IsNPUResourceName(name), convey.ShouldBeFalse)
			}
		})
	})
}

// TestIsStaticCutResourceName test IsStaticCutResourceName
func TestIsStaticCutResourceName(t *testing.T) {
	convey.Convey("IsStaticCutResourceName", t, func() {
		convey.Convey("01-static segmentation resource names should return true", func() {
			staticCutNames := []string{
				"huawei.com/Ascend910-2c",
				"huawei.com/Ascend910-3c",
				"huawei.com/Ascend910-6c",
				"huawei.com/Ascend910-8c",
				"huawei.com/Ascend910-10c",
				"huawei.com/Ascend910-12c",
				"huawei.com/Ascend910-16c",
				"huawei.com/Ascend310P-1c",
				"huawei.com/Ascend310P-2c",
				"huawei.com/Ascend310P-4c",
			}
			for _, name := range staticCutNames {
				convey.So(IsStaticCutResourceName(name), convey.ShouldBeTrue)
			}
		})
		convey.Convey("02-non-static segmentation resource names should return false", func() {
			nonStaticCutNames := []string{
				"huawei.com/npu",
				"huawei.com/npu-core",
				"huawei.com/Ascend910",
				"huawei.com/Ascend910B",
				"huawei.com/Ascend310P",
				"huawei.com/Ascend910B-4c",
				"huawei.com/Ascend310-2c",
				"huawei.com/Ascend910-c",
				"huawei.com/Ascend910-1c",
				"huawei.com/Ascend910-7c",
				"huawei.com/Ascend310P-8c",
				"huawei.com/Ascend910-2cpu",
				"huawei.com/dpu.resource.name",
				"cpu",
				"",
			}
			for _, name := range nonStaticCutNames {
				convey.So(IsStaticCutResourceName(name), convey.ShouldBeFalse)
			}
		})
	})
}
