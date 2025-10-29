/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

package fileutils

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestCheckPath(t *testing.T) {
	convey.Convey("test CheckPath", t, func() {
		// normal path
		path := "/etc/faultdiag-online/config/config.yaml"
		absPath, err := CheckPath(path)
		convey.So(err, convey.ShouldBeNil)
		convey.So(absPath, convey.ShouldEqual, path)
		path = "/tmp/../../etc/passwd"
		absPath, err = CheckPath(path)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "path traversal detected")
		convey.So(absPath, convey.ShouldBeEmpty)
		// non-existed path
		path = "/dsadsa/dsdsad123123/dsadascsafsaeda/wseqweqw"
		absPath, err = CheckPath(path)
		convey.So(err, convey.ShouldBeNil)
		convey.So(absPath, convey.ShouldEqual, path)
		// non-existed path
		path = "/dsadsa/dsdsad123123/dsadascsafsaeda/wseqweqw/"
		absPath, err = CheckPath(path)
		convey.So(err, convey.ShouldBeNil)
		convey.So(absPath, convey.ShouldNotBeEmpty)
		// supported path
		path = "./file.go"
		absPath, err = CheckPath(path)
		convey.So(err, convey.ShouldBeNil)
		convey.So(absPath, convey.ShouldNotBeEmpty)
	})
}
