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

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/plugin"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
)

var outBandTestErr = errors.New("outband test error")

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	err := hwlog.InitRunLogger(&hwLogConfig, context.Background())
	if err != nil {
		return
	}
}

func TestNewOutBandResetPlugin(t *testing.T) {
	convey.Convey("test NewOutBandResetPlugin", t, func() {
		convey.Convey("01-creates plugin with dmgr", func() {
			p := NewOutBandResetPlugin(&devmanager.DeviceManagerMock{})
			convey.So(p, convey.ShouldNotBeNil)
			convey.So(p.Name(), convey.ShouldEqual, "outbandReset")
		})
	})
}

func TestOutBandResetPlugin_CustomReset(t *testing.T) {
	convey.Convey("test OutBandResetPlugin CustomReset", t, func() {
		p := NewOutBandResetPlugin(&devmanager.DeviceManagerMock{})
		ctx := context.Background()
		convey.Convey("01-returns nil when resetErr is nil", func() {
			err := p.CustomReset(ctx, nil, nil)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("02-returns resetErr when cardType is not A3", func() {
			devs := []plugin.ResetDevice{{LogicID: 0, CardType: api.Ascend910B}}
			err := p.CustomReset(ctx, devs, outBandTestErr)
			convey.So(err, convey.ShouldEqual, outBandTestErr)
		})
	})
}

func TestOutBandResetPlugin_CustomReset_A3Success(t *testing.T) {
	convey.Convey("test OutBandResetPlugin CustomReset A3 success", t, func() {
		mockDmgr := &devmanager.DeviceManagerMock{}
		p := NewOutBandResetPlugin(mockDmgr)
		ctx := context.Background()
		devs := []plugin.ResetDevice{{LogicID: 0, CardType: api.Ascend910A3}}
		patch1 := gomonkey.ApplyMethodReturn(mockDmgr, "GetOutBandChannelState", nil)
		patch2 := gomonkey.ApplyMethodReturn(mockDmgr, "PreResetSoc", nil)
		patch3 := gomonkey.ApplyMethodReturn(mockDmgr, "SetDeviceResetOutBand", nil)
		patch4 := gomonkey.ApplyMethodReturn(mockDmgr, "RescanSoc", nil)
		patch5 := gomonkey.ApplyMethodReturn(mockDmgr, "GetDeviceBootStatus", common.BootStartFinish, nil)
		defer patch1.Reset()
		defer patch2.Reset()
		defer patch3.Reset()
		defer patch4.Reset()
		defer patch5.Reset()
		err := p.CustomReset(ctx, devs, outBandTestErr)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestOutBandResetPlugin_CustomReset_A3Fail(t *testing.T) {
	convey.Convey("test OutBandResetPlugin CustomReset A3 fail", t, func() {
		mockDmgr := &devmanager.DeviceManagerMock{}
		p := NewOutBandResetPlugin(mockDmgr)
		ctx := context.Background()
		devs := []plugin.ResetDevice{{LogicID: 0, CardType: api.Ascend910A3}}
		patch1 := gomonkey.ApplyMethodReturn(mockDmgr, "GetOutBandChannelState", outBandTestErr)
		defer patch1.Reset()
		err := p.CustomReset(ctx, devs, outBandTestErr)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestOutBandResetPlugin_CustomReset_A3BootTimeout(t *testing.T) {
	convey.Convey("test CustomReset A3 boot timeout", t, func() {
		mockDmgr := &devmanager.DeviceManagerMock{}
		p := NewOutBandResetPlugin(mockDmgr)
		ctx := context.Background()
		devs := []plugin.ResetDevice{{LogicID: 0, CardType: api.Ascend910A3}}
		patchWait := gomonkey.ApplyPrivateMethod(p, "waitRingResetComplete",
			func(_ *OutBandResetPlugin, _ []plugin.ResetDevice) error {
				return errors.New("boot timeout")
			})
		defer patchWait.Reset()
		patch1 := gomonkey.ApplyMethodReturn(mockDmgr, "GetOutBandChannelState", nil)
		patch2 := gomonkey.ApplyMethodReturn(mockDmgr, "PreResetSoc", nil)
		patch3 := gomonkey.ApplyMethodReturn(mockDmgr, "SetDeviceResetOutBand", nil)
		patch4 := gomonkey.ApplyMethodReturn(mockDmgr, "RescanSoc", nil)
		defer patch1.Reset()
		defer patch2.Reset()
		defer patch3.Reset()
		defer patch4.Reset()
		err := p.CustomReset(ctx, devs, outBandTestErr)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestOutBandResetPlugin_resetDeviceOutBand(t *testing.T) {
	convey.Convey("test OutBandResetPlugin resetDeviceOutBand", t, func() {
		p := NewOutBandResetPlugin(&devmanager.DeviceManagerMock{})
		convey.Convey("01-returns error when GetOutBandChannelState fails", func() {
			patch1 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetOutBandChannelState", outBandTestErr)
			defer patch1.Reset()
			err := p.resetDeviceOutBand(0)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-returns error when PreResetSoc fails", func() {
			patch1 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetOutBandChannelState", nil)
			patch2 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "PreResetSoc", outBandTestErr)
			defer patch1.Reset()
			defer patch2.Reset()
			err := p.resetDeviceOutBand(0)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("03-returns error when SetDeviceResetOutBand fails", func() {
			patch1 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetOutBandChannelState", nil)
			patch2 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "PreResetSoc", nil)
			patch3 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "SetDeviceResetOutBand", outBandTestErr)
			defer patch1.Reset()
			defer patch2.Reset()
			defer patch3.Reset()
			err := p.resetDeviceOutBand(0)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("04-returns error when RescanSoc fails", func() {
			patch1 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "GetOutBandChannelState", nil)
			patch2 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "PreResetSoc", nil)
			patch3 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "SetDeviceResetOutBand", nil)
			patch4 := gomonkey.ApplyMethodReturn(&devmanager.DeviceManagerMock{}, "RescanSoc", outBandTestErr)
			defer patch1.Reset()
			defer patch2.Reset()
			defer patch3.Reset()
			defer patch4.Reset()
			err := p.resetDeviceOutBand(0)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestOutBandResetPlugin_allDevicesBooted(t *testing.T) {
	convey.Convey("test allDevicesBooted", t, func() {
		mockDmgr := &devmanager.DeviceManagerMock{}
		p := NewOutBandResetPlugin(mockDmgr)
		convey.Convey("01-all booted returns true", func() {
			patch := gomonkey.ApplyMethodReturn(mockDmgr, "GetDeviceBootStatus", common.BootStartFinish, nil)
			defer patch.Reset()
			devs := []plugin.ResetDevice{{LogicID: 0}, {LogicID: 1}}
			convey.So(p.allDevicesBooted(devs), convey.ShouldBeTrue)
		})
		convey.Convey("02-one not booted returns false", func() {
			patch := gomonkey.ApplyMethodReturn(mockDmgr, "GetDeviceBootStatus", 0, nil)
			defer patch.Reset()
			devs := []plugin.ResetDevice{{LogicID: 0}}
			convey.So(p.allDevicesBooted(devs), convey.ShouldBeFalse)
		})
		convey.Convey("03-error returns false", func() {
			patch := gomonkey.ApplyMethodReturn(mockDmgr, "GetDeviceBootStatus", 0, outBandTestErr)
			defer patch.Reset()
			devs := []plugin.ResetDevice{{LogicID: 0}}
			convey.So(p.allDevicesBooted(devs), convey.ShouldBeFalse)
		})
	})
}
