/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

package builtin

import (
	"context"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"

	"Ascend-device-plugin/pkg/kubeclient"
	"Ascend-device-plugin/pkg/plugin"
	"ascend-common/common-utils/hwlog"
)

var resetRecordTestErr = errors.New("reset record test error")

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	err := hwlog.InitRunLogger(&hwLogConfig, context.Background())
	if err != nil {
		return
	}
}

func TestNewResetRecordPlugin(t *testing.T) {
	convey.Convey("test NewResetRecordPlugin", t, func() {
		convey.Convey("01-creates plugin with client", func() {
			mockNodeName := gomonkey.ApplyFuncReturn(kubeclient.GetNodeNameFromEnv, "test-node", nil)
			defer mockNodeName.Reset()
			p := NewResetRecordPlugin(&kubeclient.ClientK8s{})
			convey.So(p, convey.ShouldNotBeNil)
			convey.So(p.Name(), convey.ShouldEqual, "resetRecord")
		})
	})
}

func TestResetRecordPlugin_PreReset(t *testing.T) {
	convey.Convey("test ResetRecordPlugin PreReset", t, func() {
		p := &ResetRecordPlugin{client: &kubeclient.ClientK8s{}, nodeName: "test-node"}
		ctx := context.Background()
		convey.Convey("01-executes without error when event created successfully", func() {
			mockCreate := gomonkey.ApplyMethodReturn(p.client, "CreateEvent", &v1.Event{}, nil)
			defer mockCreate.Reset()
			p.PreReset(ctx, []plugin.ResetDevice{{LogicID: 0}})
		})
		convey.Convey("02-executes without error when CreateEvent fails", func() {
			mockCreate := gomonkey.ApplyMethodReturn(p.client, "CreateEvent", (*v1.Event)(nil), resetRecordTestErr)
			defer mockCreate.Reset()
			p.PreReset(ctx, []plugin.ResetDevice{{LogicID: 0}})
		})
	})
}

func TestResetRecordPlugin_AfterReset(t *testing.T) {
	convey.Convey("test ResetRecordPlugin AfterReset", t, func() {
		p := &ResetRecordPlugin{client: &kubeclient.ClientK8s{}, nodeName: "test-node"}
		ctx := context.Background()
		convey.Convey("01-executes without error when resetErr is nil and event created successfully", func() {
			mockCreate := gomonkey.ApplyMethodReturn(p.client, "CreateEvent", &v1.Event{}, nil)
			defer mockCreate.Reset()
			p.AfterReset(ctx, []plugin.ResetDevice{{LogicID: 0}}, nil)
		})
		convey.Convey("02-executes without error when resetErr is not nil and event created successfully", func() {
			mockCreate := gomonkey.ApplyMethodReturn(p.client, "CreateEvent", &v1.Event{}, nil)
			defer mockCreate.Reset()
			p.AfterReset(ctx, []plugin.ResetDevice{{LogicID: 0}}, resetRecordTestErr)
		})
	})
}

func TestFormatDeviceList(t *testing.T) {
	convey.Convey("test formatDeviceList", t, func() {
		convey.Convey("01-returns empty string for empty list", func() {
			result := formatDeviceList(nil)
			convey.So(result, convey.ShouldEqual, "")
		})
		convey.Convey("02-returns comma-separated logicIDs", func() {
			devs := []plugin.ResetDevice{{LogicID: 1}, {LogicID: 2}, {LogicID: 3}}
			result := formatDeviceList(devs)
			convey.So(result, convey.ShouldEqual, "1,2,3")
		})
	})
}

func TestGetFaultDevID(t *testing.T) {
	convey.Convey("test getFaultDevID", t, func() {
		convey.Convey("01-returns fault dev LogicID when present", func() {
			devs := []plugin.ResetDevice{
				{LogicID: 0, IsFaultDev: false},
				{LogicID: 1, IsFaultDev: true},
				{LogicID: 2, IsFaultDev: false},
			}
			convey.So(getFaultDevID(devs), convey.ShouldEqual, 1)
		})
		convey.Convey("02-returns -1 when no fault dev", func() {
			devs := []plugin.ResetDevice{{LogicID: 0}, {LogicID: 1}}
			convey.So(getFaultDevID(devs), convey.ShouldEqual, invalidFaultDevID)
		})
		convey.Convey("03-returns first fault dev when multiple", func() {
			devs := []plugin.ResetDevice{
				{LogicID: 0, IsFaultDev: true},
				{LogicID: 1, IsFaultDev: true},
			}
			convey.So(getFaultDevID(devs), convey.ShouldEqual, 0)
		})
	})
}

func TestGetFaultTokensLeft(t *testing.T) {
	convey.Convey("test getFaultTokensLeft", t, func() {
		convey.Convey("01-returns tokens from fault dev", func() {
			devs := []plugin.ResetDevice{
				{LogicID: 0},
				{LogicID: 1, IsFaultDev: true, TokensLeft: 2},
			}
			convey.So(getFaultTokensLeft(devs), convey.ShouldEqual, 2)
		})
		convey.Convey("02-returns -1 when no fault dev", func() {
			devs := []plugin.ResetDevice{{LogicID: 0}}
			convey.So(getFaultTokensLeft(devs), convey.ShouldEqual, invalidTokensLeft)
		})
		convey.Convey("03-returns zero when tokens exhausted", func() {
			devs := []plugin.ResetDevice{
				{LogicID: 0, IsFaultDev: true, TokensLeft: 0},
			}
			convey.So(getFaultTokensLeft(devs), convey.ShouldEqual, 0)
		})
	})
}
