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

// Package policy is used for processing superpod information
package policy

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
)

const logLineLength = 256

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

// TestReadConfigMap test for func readConfigMap
func TestReadConfigMap(t *testing.T) {
	configFilePath := "./super-pod-0.json"
	convey.Convey("Test ReadConfigMap", t, func() {
		convey.Convey("return nil when no configMap path", func() {
			ret := readConfigMap(configFilePath)
			convey.So(ret, convey.ShouldBeNil)
		})

		convey.Convey("return nil when ConfigMapPath file readAll fail", func() {
			patch := gomonkey.ApplyFunc(os.Open, func(path string) (*os.File, error) {
				return nil, nil
			})
			defer patch.Reset()
			ret := readConfigMap(configFilePath)
			convey.So(ret, convey.ShouldBeNil)
		})

		convey.Convey("return nil when ConfigMapPath file unmarshal fail", func() {
			configFile, err := os.Create(configFilePath)
			if err != nil {
				return
			}
			defer configFile.Close()
			defer configFile.Chmod(0600) //文件权限
			ret := readConfigMap(configFilePath)
			convey.So(ret, convey.ShouldBeNil)
			err = os.Remove(configFilePath)
			if err != nil {
				return
			}
		})

		convey.Convey("return nil when ConfigMapPath file unmarshal pass", func() {
			configFile, err := os.Create(configFilePath)
			if err != nil {
				return
			}
			defer configFile.Close()
			defer configFile.Chmod(0600) //文件权限
			_, err = io.WriteString(configFile, `{"SuperPodID":"0","RackMap":null}`)
			if err != nil {
				return
			}
			ret := readConfigMap(configFilePath)
			convey.So(ret, convey.ShouldNotBeNil)
			err = os.Remove(configFilePath)
			if err != nil {
				return
			}
		})
	})
}

// TestGetCurSuperPodInfoFromMapA3 test for func GetCurSuperPodInfoFromMapA3
func TestGetCurSuperPodInfoFromMapA3(t *testing.T) {
	convey.Convey("test GetSuperPodInfoFromMapA3", t, func() {
		convey.Convey("should return nil when input configmap nil", func() {
			fullMesh, linkPath := GetCurSuperPodInfoFromMapA3(nil)
			convey.So(fullMesh, convey.ShouldBeNil)
			convey.So(linkPath, convey.ShouldBeNil)
		})
		convey.Convey("should return nil when NodeDeviceMap format err", func() {
			configMap := &SuperPodInfo{
				NodeDeviceMap: nil,
			}
			fullMesh, linkPath := GetCurSuperPodInfoFromMapA3(configMap)
			convey.So(fullMesh, convey.ShouldBeNil)
			convey.So(linkPath, convey.ShouldBeNil)
		})
		convey.Convey("should return func value when valid", func() {
			configMap := &SuperPodInfo{
				NodeDeviceMap: map[string]*NodeDevice{},
			}
			mockFullMesh := []string{"mockFullMesh"}
			mockLinkPath := map[string][]string{"mockLinkPath": {"mockLinkPath"}}
			mockParseNodeDeviceMap := gomonkey.ApplyFuncReturn(parseNodeDeviceMap, mockFullMesh, mockLinkPath)
			defer mockParseNodeDeviceMap.Reset()
			fullMesh, linkPath := GetCurSuperPodInfoFromMapA3(configMap)
			convey.So(fullMesh, convey.ShouldResemble, mockFullMesh)
			convey.So(linkPath, convey.ShouldResemble, mockLinkPath)
		})
	})
}
