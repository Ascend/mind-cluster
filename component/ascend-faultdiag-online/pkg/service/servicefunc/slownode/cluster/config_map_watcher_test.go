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

// Package cluster a DT collection for func in config_map_watcher
package cluster

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"ascend-common/common-utils/hwlog"
)

const (
	num1          = 1
	num2          = 2
	num11         = 11
	num12         = 12
	logLineLength = 256
)

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

func TestGetConfigMapWatcher(t *testing.T) {
	cmWatcher := GetConfigMapWatcher()
	assert.NotNil(t, cmWatcher)
	assert.Equal(t, len(cmWatcher.watchers), 0)
}

func TestWatcher(t *testing.T) {
	patcher := gomonkey.NewPatches()
	defer patcher.Reset()
	cmWatcher := GetConfigMapWatcher()
	patcher.ApplyPrivateMethod(reflect.TypeOf(cmWatcher), "runInformer",
		func(*ConfigMapWatcher, context.Context, string, string) {})
	convey.Convey("test configMap watcher", t, func() {
		convey.Convey("test add cm watcher", func() {
			cmWatcher.WatchConfigMap("testNamespace1", "testcmName")
			// duplicated name space
			cmWatcher.WatchConfigMap("testNamespace1", "testcmName")
			convey.So(len(cmWatcher.watchers), convey.ShouldEqual, num1)
			// add different cm
			cmWatcher.WatchConfigMap("testNamespace2", "testcmName")
			convey.So(len(cmWatcher.watchers), convey.ShouldEqual, num2)
			// parallel adding 10 different cm
			for i := 0; i < 10; i++ {
				cmWatcher.WatchConfigMap(fmt.Sprintf("testNamespace%d", i), "newcmName")
			}
			convey.So(len(cmWatcher.watchers), convey.ShouldEqual, num12)
		})
		convey.Convey("test delete cmwatcher", func() {
			initLen := len(cmWatcher.watchers)
			cmWatcher.StopWatching("testNamespace1", "testcmName")
			convey.So(len(cmWatcher.watchers), convey.ShouldEqual, initLen-num1)
			// delete non exit cm
			cmWatcher.StopWatching("testNamespace1", "testcmName")
			convey.So(len(cmWatcher.watchers), convey.ShouldEqual, initLen-num1)
			// parallel stop
			for i := 0; i < 10; i++ {
				cmWatcher.StopWatching(fmt.Sprintf("testNamespace%d", i), "newcmName")
			}
			convey.So(len(cmWatcher.watchers), convey.ShouldEqual, initLen-num11)
		})
	})
}
