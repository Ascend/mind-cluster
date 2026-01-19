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

// Package pingmeshconfig for faultnetwork feature
package pingmeshconfig

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

// TestParseFaultNetworkInfoCM test case for func ParseFaultNetworkInfoCM
func TestParseFaultNetworkInfoCM(t *testing.T) {
	convey.Convey("test func ParseFaultNetworkInfoCM", t, func() {
		convey.Convey("should return err when obj is nil", func() {
			_, err := ParseFaultNetworkInfoCM(nil)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "not fault network of ras feature configmap")
		})
		convey.Convey("should return err when obj is not ConfigMap type", func() {
			_, err := ParseFaultNetworkInfoCM("")
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "not fault network of ras feature configmap")
		})
		convey.Convey("should return empty when config map data is invalid format", func() {
			cm := &v1.ConfigMap{
				Data: map[string]string{
					"k": "value",
				},
			}
			_, err := ParseFaultNetworkInfoCM(cm)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "unmarshal failed")
		})
		convey.Convey("obj is valid", func() {
			cm := &v1.ConfigMap{}
			cm.Name = "testCmName"
			configInfo := constant.ConfigPingMesh{}
			configInfo["global"] = &constant.HccspingMeshItem{Activate: "on"}
			cm.Data = map[string]string{"global": util.ObjToString(configInfo)}
			_, err := ParseFaultNetworkInfoCM(cm)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestDeepCopy test case for func DeepCopy
func TestDeepCopy(t *testing.T) {
	convey.Convey("test func DeepCopy", t, func() {
		convey.Convey("should return nil when info is nil", func() {
			info := DeepCopy(nil)
			convey.So(info, convey.ShouldBeNil)
		})
		convey.Convey("should return nil when json Marshal failed", func() {
			patch := gomonkey.ApplyFunc(json.Marshal, func(v any) ([]byte, error) {
				return nil, errors.New("can not marshal")
			})
			defer patch.Reset()
			info := constant.ConfigPingMesh{
				"global": &constant.HccspingMeshItem{
					Activate: "on",
				},
			}
			infoCopied := DeepCopy(info)
			convey.So(infoCopied, convey.ShouldBeNil)
		})
		convey.Convey("should return nil when json Unmarshal failed", func() {
			patch := gomonkey.ApplyFunc(json.Unmarshal, func(data []byte, v any) error {
				return errors.New("can not unmarshal")
			})
			defer patch.Reset()
			info := constant.ConfigPingMesh{
				"global": &constant.HccspingMeshItem{
					Activate: "on",
				},
			}
			infoCopied := DeepCopy(info)
			convey.So(infoCopied, convey.ShouldBeNil)
		})
		convey.Convey("should same with origin info when input is normal data", func() {
			info := constant.ConfigPingMesh{
				"global": &constant.HccspingMeshItem{
					Activate: "on",
				},
			}
			newInfo := DeepCopy(info)
			convey.So(info["global"].Activate, convey.ShouldEqual, newInfo["global"].Activate)
		})
	})
}
