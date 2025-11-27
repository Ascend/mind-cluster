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

// Package topology for generate topology of Rack
package topology

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

func TestTopoFileToStr(t *testing.T) {
	convey.Convey("test topoFileToStr", t, func() {
		convey.Convey("read file failed", func() {
			mock1 := gomonkey.ApplyFunc(os.ReadFile, func(_ string) ([]byte, error) {
				return nil, fmt.Errorf("fake error")
			})
			defer mock1.Reset()
			_, err := topoFileToStr("")
			convey.So(err, convey.ShouldNotBeNil)
		})
		mock2 := gomonkey.ApplyFunc(os.ReadFile, func(_ string) ([]byte, error) { return make([]byte, 0), nil })
		defer mock2.Reset()
		convey.Convey("json valid failed", func() {
			mock3 := gomonkey.ApplyFunc(json.Valid, func(_ []byte) bool { return false })
			defer mock3.Reset()
			_, err := topoFileToStr("")
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestToFile test case for ToFile
func TestToFile(t *testing.T) {
	var topoFilePath string
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
		topoFilePath = "/tmp/topology.json"
	} else {
		topoFilePath = filepath.Join(dir, "topology.json")
	}

	convey.Convey("test ToFile", t, func() {
		convey.Convey("test ToFile should be success", func() {
			mock1 := gomonkey.ApplyFunc(topoFileToStr, func(_ string) (string, error) {
				return "", nil
			})
			defer mock1.Reset()
			err := ToFile(topoFilePath, "")
			convey.So(err, convey.ShouldBeNil)
		})

	})
}
func TestGetFileHash(t *testing.T) {
	convey.Convey("test getFileHash err", t, func() {
		mockReadFile := gomonkey.ApplyFunc(os.ReadFile, func(_ string) ([]byte, error) {
			return []byte{}, errors.New("fake error")
		})
		defer mockReadFile.Reset()
		_, err := getFileHash("test1")
		convey.So(err, convey.ShouldNotBeNil)
	})

	convey.Convey("test getFileHash success", t, func() {
		mockReadFile := gomonkey.ApplyFunc(os.ReadFile, func(_ string) ([]byte, error) {
			return []byte{'a', 'b'}, nil
		})
		defer mockReadFile.Reset()
		_, err := getFileHash("test2")
		convey.So(err, convey.ShouldBeNil)
	})
}
