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
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"Ascend-device-plugin/pkg/kubeclient"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	err := hwlog.InitRunLogger(&hwLogConfig, context.Background())
	if err != nil {
		return
	}
}

func TestInitPluginManager(t *testing.T) {
	convey.Convey("test InitPluginManager", t, func() {
		convey.Convey("01-initializes with built-in plugins", func() {
			mockNodeName := gomonkey.ApplyFuncReturn(kubeclient.GetNodeNameFromEnv, "test-node", nil)
			defer mockNodeName.Reset()
			pm, err := InitPluginManager(&devmanager.DeviceManagerMock{}, &kubeclient.ClientK8s{})
			convey.So(err, convey.ShouldBeNil)
			convey.So(pm, convey.ShouldNotBeNil)
			defer pm.Stop()
			_, ok1 := pm.GetPlugin("outbandReset")
			convey.So(ok1, convey.ShouldBeTrue)
			_, ok2 := pm.GetPlugin("resetRecord")
			convey.So(ok2, convey.ShouldBeTrue)
		})
		convey.Convey("02-default config only enables outbandReset", func() {
			mockNodeName := gomonkey.ApplyFuncReturn(kubeclient.GetNodeNameFromEnv, "test-node", nil)
			defer mockNodeName.Reset()
			pm, err := InitPluginManager(&devmanager.DeviceManagerMock{}, &kubeclient.ClientK8s{})
			convey.So(err, convey.ShouldBeNil)
			defer pm.Stop()
			pre, custom, after := pm.GetHookChains()
			convey.So(len(pre), convey.ShouldEqual, 1)
			convey.So(len(custom), convey.ShouldEqual, 1)
			convey.So(len(after), convey.ShouldEqual, 1)
			convey.So(pre[0].Name(), convey.ShouldEqual, "outbandReset")
			convey.So(custom[0].Name(), convey.ShouldEqual, "outbandReset")
			convey.So(after[0].Name(), convey.ShouldEqual, "outbandReset")
		})
	})
}

func TestInitPluginManager_PluginCount(t *testing.T) {
	convey.Convey("test InitPluginManager plugin count", t, func() {
		convey.Convey("01-registers exactly 2 built-in plugins", func() {
			mockNodeName := gomonkey.ApplyFuncReturn(kubeclient.GetNodeNameFromEnv, "test-node", nil)
			defer mockNodeName.Reset()
			pm, err := InitPluginManager(&devmanager.DeviceManagerMock{}, &kubeclient.ClientK8s{})
			convey.So(err, convey.ShouldBeNil)
			defer pm.Stop()
			convey.So(len(pm.Plugins), convey.ShouldEqual, 2)
		})
	})
}

func TestInitPluginManager_NilDmgr(t *testing.T) {
	convey.Convey("test InitPluginManager nil dmgr", t, func() {
		convey.Convey("01-creates plugin manager with nil dmgr", func() {
			mockNodeName := gomonkey.ApplyFuncReturn(kubeclient.GetNodeNameFromEnv, "test-node", nil)
			defer mockNodeName.Reset()
			pm, err := InitPluginManager(nil, &kubeclient.ClientK8s{})
			convey.So(err, convey.ShouldBeNil)
			convey.So(pm, convey.ShouldNotBeNil)
			defer pm.Stop()
		})
	})
}

func TestInitPluginManager_NilKubeClient(t *testing.T) {
	convey.Convey("test InitPluginManager nil kubeClient", t, func() {
		convey.Convey("01-creates plugin manager with nil kubeClient", func() {
			mockNodeName := gomonkey.ApplyFuncReturn(kubeclient.GetNodeNameFromEnv, "test-node", nil)
			defer mockNodeName.Reset()
			pm, err := InitPluginManager(&devmanager.DeviceManagerMock{}, nil)
			convey.So(err, convey.ShouldBeNil)
			convey.So(pm, convey.ShouldNotBeNil)
			defer pm.Stop()
		})
	})
}
