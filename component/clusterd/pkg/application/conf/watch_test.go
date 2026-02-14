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

// Package conf test for global config watcher
package conf

import (
	"context"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/conf"
	"clusterd/pkg/interface/kube"
)

const (
	defaultFaultWindowHours = 24
	defaultFaultThreshold   = 3
	defaultFaultFreeHours   = 48
	testCase1               = `
enabled: true
separate:
  fault_window_hours: 24
  fault_threshold: 3
release:
  fault_free_hours: 48
`
	invalidTestCase1 = `
enabled: true
separate:
  fault_window_hours: 10000
  fault_threshold: 3
release:
  fault_free_hours: 48
`
)

func getDemoCm() *v1.ConfigMap {
	data := map[string]string{constant.ManuallySeparateNPUConfigKey: testCase1}
	cm := &v1.ConfigMap{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Data:       data,
	}
	return cm
}

func TestTryLoad(t *testing.T) {
	convey.Convey("test func TryLoadGlobalConfig success", t, testTryLoad)
	convey.Convey("test func TryLoadGlobalConfig failed, get cm failed", t, testTryLoadErrGetCm)
}

func testTryLoad() {
	resetGlobalConfig()
	p1 := gomonkey.ApplyFuncReturn(kube.GetConfigMap, getDemoCm(), nil)
	defer p1.Reset()
	TryLoadGlobalConfig()
	convey.So(conf.GetManualEnabled(), convey.ShouldBeTrue)
	convey.So(conf.GetSeparateWindow(), convey.ShouldEqual, int64(defaultFaultWindowHours*constant.HoursToMilliseconds))
	convey.So(conf.GetSeparateThreshold(), convey.ShouldEqual, int64(defaultFaultThreshold))
	convey.So(conf.GetReleaseDuration(), convey.ShouldEqual, int64(defaultFaultFreeHours*constant.HoursToMilliseconds))
}

func testTryLoadErrGetCm() {
	resetGlobalConfig()
	p1 := gomonkey.ApplyFuncReturn(kube.GetConfigMap, nil, testErr)
	defer p1.Reset()
	TryLoadGlobalConfig()
	convey.So(conf.GetManualEnabled(), convey.ShouldBeFalse)
	convey.So(conf.GetSeparateWindow(), convey.ShouldEqual, 0)
	convey.So(conf.GetSeparateThreshold(), convey.ShouldEqual, 0)
	convey.So(conf.GetReleaseDuration(), convey.ShouldEqual, 0)
}

func TestLoadGlobalConfig(t *testing.T) {
	convey.Convey("test func loadGlobalConfig failed, data is nil", t, testLoadNilData)
	convey.Convey("test func loadGlobalConfig failed, key is not found", t, testLoadNotExistKey)
	convey.Convey("test func loadGlobalConfig failed, unmarshal failed", t, testLoadUnmarshalErr)
	convey.Convey("test func loadGlobalConfig failed, check failed", t, testLoadCheckErr)
}

func testLoadNilData() {
	resetGlobalConfig()
	cm := getDemoCm()
	cm.Data = nil
	loadGlobalConfig(cm)
	convey.So(conf.GetManualEnabled(), convey.ShouldBeFalse)
	convey.So(conf.GetSeparateWindow(), convey.ShouldEqual, 0)
	convey.So(conf.GetSeparateThreshold(), convey.ShouldEqual, 0)
	convey.So(conf.GetReleaseDuration(), convey.ShouldEqual, 0)
}

func testLoadNotExistKey() {
	resetGlobalConfig()
	cm := getDemoCm()
	cm.Data = map[string]string{"abc": "def"}
	loadGlobalConfig(cm)
	convey.So(conf.GetManualEnabled(), convey.ShouldBeFalse)
	convey.So(conf.GetSeparateWindow(), convey.ShouldEqual, 0)
	convey.So(conf.GetSeparateThreshold(), convey.ShouldEqual, 0)
	convey.So(conf.GetReleaseDuration(), convey.ShouldEqual, 0)
}

func testLoadUnmarshalErr() {
	resetGlobalConfig()
	cm := getDemoCm()
	cm.Data = map[string]string{constant.ManuallySeparateNPUConfigKey: "invalid yaml"}
	loadGlobalConfig(cm)
	convey.So(conf.GetManualEnabled(), convey.ShouldBeFalse)
	convey.So(conf.GetSeparateWindow(), convey.ShouldEqual, 0)
	convey.So(conf.GetSeparateThreshold(), convey.ShouldEqual, 0)
	convey.So(conf.GetReleaseDuration(), convey.ShouldEqual, 0)
}

func testLoadCheckErr() {
	resetGlobalConfig()
	cm := getDemoCm()
	cm.Data = map[string]string{constant.ManuallySeparateNPUConfigKey: invalidTestCase1}
	loadGlobalConfig(cm)
	convey.So(conf.GetManualEnabled(), convey.ShouldBeFalse)
	convey.So(conf.GetSeparateWindow(), convey.ShouldEqual, 0)
	convey.So(conf.GetSeparateThreshold(), convey.ShouldEqual, 0)
	convey.So(conf.GetReleaseDuration(), convey.ShouldEqual, 0)
}

func resetGlobalConfig() {
	conf.SetManualSeparatePolicy(conf.ManuallySeparatePolicy{})
}

func TestWatch(t *testing.T) {
	const processInterval = 500 * time.Millisecond
	convey.Convey("test func WatchGlobalConfig success", t, func() {
		var hasExecuted bool
		var p1 = gomonkey.ApplyFunc(TryLoadGlobalConfig, func() {
			hasExecuted = true
			return
		})
		defer p1.Reset()
		ctx, cancel := context.WithCancel(context.TODO())
		go WatchGlobalConfig(ctx)
		time.Sleep(processInterval)
		cancel()
		convey.So(hasExecuted, convey.ShouldBeFalse)
	})
}
