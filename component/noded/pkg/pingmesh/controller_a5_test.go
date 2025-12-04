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

/*
Package pingmesh is using for checking hccs network
*/
package pingmesh

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/client-go/kubernetes/fake"

	"ascend-common/api/slownet"
	"ascend-common/devmanager/common"
	"nodeD/pkg/kubeclient"
	"nodeD/pkg/pingmesh/executor"
	"nodeD/pkg/pingmesh/policygenerator"
	"nodeD/pkg/pingmesh/policygenerator/fullmesh"
	"nodeD/pkg/pingmesh/roceping"
	"nodeD/pkg/pingmesh/types"
	_ "nodeD/pkg/testtool"
)

const (
	testActivate = "off"
	testInterval = 1

	testFilePath = "test"
)

var fakeClient = fake.NewSimpleClientset()

func TestManagerParseRoCEPingConfig(t *testing.T) {
	convey.Convey("test Manager method parseRoCEPingConfig", t, func() {
		convey.Convey("01-should return err when roce cfg not exist", func() {
			data := map[string]string{}
			m := &Manager{}
			_, err := m.parseRoCEPingConfig(data)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("02-should return err when json unmarshal failed", func() {
			data := map[string]string{
				"roce": `{"activate": "off", "task_interval": x}`,
			}
			m := &Manager{}
			_, err := m.parseRoCEPingConfig(data)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("03-should return err when active value invalid", func() {
			data := map[string]string{
				"roce": `{"activate": "xxx", "task_interval": 1}`,
			}
			m := &Manager{}
			_, err := m.parseRoCEPingConfig(data)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("04-should return err when task_interval value invalid", func() {
			data := map[string]string{
				"roce": `{"activate": "off", "task_interval": 0}`,
			}
			m := &Manager{}
			_, err := m.parseRoCEPingConfig(data)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("05-should return success when all cfg is valid", func() {
			data := map[string]string{
				"roce": `{"activate": "off", "task_interval": 1}`,
			}
			m := &Manager{}
			_, err := m.parseRoCEPingConfig(data)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestGenerateJobUIDA5(t *testing.T) {
	config := &types.HccspingMeshConfig{}
	destAddrs := map[string]types.SuperDeviceIDs{
		"node1": map[string]string{"1": "111"},
	}
	pingFileHash := "abc"

	convey.Convey("Testing generateJobUIDA5 function", t, func() {
		convey.Convey("01-should return err when json Marshal failed", func() {
			patch := gomonkey.ApplyFuncReturn(json.Marshal, nil, errors.New("json Marshal failed"))
			defer patch.Reset()
			_, err := generateJobUIDA5(config, destAddrs, pingFileHash)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("02-should return err when sha256 hash Write failed", func() {
			patch := gomonkey.ApplyMethodReturn(sha256.New(), "Write", nil, errors.New("sha256 hash Write failed"))
			defer patch.Reset()
			_, err := generateJobUIDA5(config, destAddrs, pingFileHash)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("03-should return valid uid when all function called success", func() {
			_, err := generateJobUIDA5(config, destAddrs, pingFileHash)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestInitRoCEFileWatcher(t *testing.T) {
	patch := gomonkey.ApplyFunc(executor.New, func() (*executor.DevManager, error) {
		return &executor.DevManager{SuperPodId: 1}, nil
	})
	patchReset := mockGetDeviceTypeA5()
	defer patch.Reset()
	defer patchReset()
	m := NewManager(&Config{
		ResultMaxAge: DefaultResultMaxAge,
		KubeClient: &kubeclient.ClientK8s{
			ClientSet: fakeClient,
			NodeName:  fakeNode,
		},
	})
	convey.Convey("01--Testing InitRoCEFileWatcher, rocePingFileWatcher should not be nil", t, func() {
		m.pingManager = &roceping.PingManager{}
		patch4 := gomonkey.ApplyFuncReturn(slownet.GetRoCEPingListFilePath, testFilePath, nil)
		ctx, cancel := context.WithCancel(context.Background())
		m.InitRoCEFileWatcher(ctx)
		defer func() {
			patch4.Reset()
			cancel()
			m.rocePingFileWatcher = nil
		}()
		if reflect.ValueOf(m.rocePingFileWatcher).IsNil() {
			t.Errorf("02--Test InitRoCEFileWatcher failed")
		}
	})
}

func TestInitRoCEFileWatcherForNil(t *testing.T) {
	patch := gomonkey.ApplyFunc(executor.New, func() (*executor.DevManager, error) {
		return &executor.DevManager{SuperPodId: 1}, nil
	})
	defer patch.Reset()
	m := NewManager(&Config{
		ResultMaxAge: DefaultResultMaxAge,
		KubeClient: &kubeclient.ClientK8s{
			ClientSet: fakeClient,
			NodeName:  fakeNode,
		},
	})
	convey.Convey("03--Testing InitRoCEFileWatcher, rocePingFileWatcher should not be nil for not a5", t, func() {
		m.pingManager = &roceping.PingManager{}
		ctx, cancel := context.WithCancel(context.Background())
		m.InitRoCEFileWatcher(ctx)
		defer func() {
			cancel()
			m.rocePingFileWatcher = nil
		}()
		convey.ShouldNotBeNil(m)
		if !reflect.ValueOf(m.rocePingFileWatcher).IsNil() {
			t.Errorf("03--Test InitRoCEFileWatcher failed")
		}
	})
	convey.Convey("04--Testing InitRoCEFileWatcher, rocePingFileWatcher should be nil for path not exist", t, func() {
		m.pingManager = &roceping.PingManager{}
		patchReset := mockGetDeviceTypeA5()
		ctx, cancel := context.WithCancel(context.Background())
		m.InitRoCEFileWatcher(ctx)
		defer func() {
			patchReset()
			cancel()
		}()
		convey.ShouldNotBeNil(m)
		if !reflect.ValueOf(m.rocePingFileWatcher).IsNil() {
			t.Errorf("04--Test InitRoCEFileWatcher failed")
		}
	})
}

func TestInitRoCEFileWatcherForError(t *testing.T) {
	patch := gomonkey.ApplyFunc(executor.New, func() (*executor.DevManager, error) {
		return &executor.DevManager{SuperPodId: 1}, nil
	})
	defer patch.Reset()
	patchReset := mockGetDeviceTypeA5()
	defer patchReset()
	m := NewManager(&Config{
		ResultMaxAge: DefaultResultMaxAge,
		KubeClient: &kubeclient.ClientK8s{
			ClientSet: fakeClient,
			NodeName:  fakeNode,
		},
	})
	convey.Convey("05--Testing InitRoCEFileWatcher, rocePingFileWatcher should be nil for get file error", t, func() {
		m.pingManager = &roceping.PingManager{}
		patch4 := gomonkey.ApplyFuncReturn(slownet.GetRoCEPingListFilePath, testFilePath, errors.New("failed"))
		ctx, cancel := context.WithCancel(context.Background())
		m.InitRoCEFileWatcher(ctx)
		defer func() {
			patch4.Reset()
			cancel()
		}()
		convey.ShouldNotBeNil(m)
		if !reflect.ValueOf(m.rocePingFileWatcher).IsNil() {
			t.Errorf("02--Test InitRoCEFileWatcher failed")
		}
	})
	convey.Convey("06--Testing InitRoCEFileWatcher, rocePingFileWatcher should not be nil for add path error",
		t, func() {
			m.pingManager = &roceping.PingManager{}
			patch4 := gomonkey.ApplyFuncReturn(slownet.GetRoCEPingListFilePath, testFilePath, nil)
			patch5 := gomonkey.ApplyMethodReturn(&roceping.FileWatcherLoop{}, "AddListenPath", errors.New("failed"))
			ctx, cancel := context.WithCancel(context.Background())
			m.InitRoCEFileWatcher(ctx)
			defer func() {
				cancel()
				patch4.Reset()
				patch5.Reset()
			}()
			convey.ShouldNotBeNil(m)
			if reflect.ValueOf(m.rocePingFileWatcher).IsNil() {
				t.Errorf("04--Test InitRoCEFileWatcher failed")
			}
		})
}

func TestUpdateRoceConfigSuccess(t *testing.T) {
	convey.Convey("01--Testing UpdateRoceConfig call success", t, func() {
		patch := gomonkey.ApplyFunc(executor.New, func() (*executor.DevManager, error) {
			return &executor.DevManager{SuperPodId: 1}, nil
		})
		patchReset := mockGetDeviceTypeA5()
		patch3 := gomonkey.ApplyFuncReturn(roceping.NewPingManager, &roceping.PingManager{})
		m := NewManager(&Config{
			ResultMaxAge: DefaultResultMaxAge,
			KubeClient: &kubeclient.ClientK8s{
				ClientSet: fakeClient,
				NodeName:  fakeNode,
			},
		})
		m.currentRoCE = &types.HccspingMeshPolicy{
			Config: &types.HccspingMeshConfig{
				Activate:     testActivate,
				TaskInterval: testInterval,
			},
			Address: map[string]types.SuperDeviceIDs{
				"1": nil,
			},
		}

		count := 0
		patch4 := gomonkey.ApplyMethod(&roceping.PingManager{}, "UpdateConfig", func(_ *roceping.PingManager,
			cfg *types.HccspingMeshPolicy) {
			count++
		})
		defer func() {
			patch.Reset()
			patchReset()
			patch3.Reset()
			patch4.Reset()
		}()
		m.updateRoCEConfig()
		convey.So(count, convey.ShouldEqual, 1)
	})
}

func TestUpdateRoceConfigFailed01(t *testing.T) {
	patch := gomonkey.ApplyFunc(executor.New, func() (*executor.DevManager, error) {
		return &executor.DevManager{SuperPodId: 1}, nil
	})
	defer patch.Reset()
	patchReset := mockGetDeviceTypeA5()
	defer patchReset()
	m := NewManager(&Config{
		ResultMaxAge: DefaultResultMaxAge,
		KubeClient: &kubeclient.ClientK8s{
			ClientSet: fakeClient,
			NodeName:  fakeNode,
		},
	})
	count := 0
	patch4 := gomonkey.ApplyMethod(&roceping.PingManager{}, "UpdateConfig", func(_ *roceping.PingManager,
		cfg *types.HccspingMeshPolicy) {
		count++
	})
	defer patch4.Reset()
	convey.Convey("02--Testing UpdateRoceConfig call failed for pingmanager is nil", t, func() {
		count = 0
		m.updateRoCEConfig()
		convey.So(count, convey.ShouldEqual, 0)
	})
	convey.Convey("03--Testing UpdateRoceConfig call failed for Config is nil", t, func() {
		m.pingManager = &roceping.PingManager{}
		count = 0
		m.updateRoCEConfig()
		convey.So(count, convey.ShouldEqual, 0)
	})
	convey.Convey("04--Testing UpdateRoceConfig call failed for address is empty", t, func() {
		m.pingManager = &roceping.PingManager{}
		m.currentRoCE = &types.HccspingMeshPolicy{
			Config: &types.HccspingMeshConfig{
				Activate:     testActivate,
				TaskInterval: testInterval,
			},
			Address: map[string]types.SuperDeviceIDs{},
		}
		count = 0
		m.updateRoCEConfig()
		convey.So(count, convey.ShouldEqual, 0)
	})
}

func TestGeneratePingPolicy(t *testing.T) {
	convey.Convey("Testing GeneratePingPolicy", t, func() {
		c := &Manager{
			executor: nil,
			ipCmName: "config",
			nodeName: "node1",
			current:  &types.HccspingMeshPolicy{},
			currentRoCE: &types.HccspingMeshPolicy{
				Config: &types.HccspingMeshConfig{
					Activate: "on",
				},
			},
			superPodId:  "0",
			serverIndex: "0",
			rackId:      "0",
		}
		gen := fullmesh.New(c.nodeName, c.superPodId, c.serverIndex)
		c.policyFactory = policygenerator.NewFactory().Register(fullmesh.Rule, gen)
		roceGen := roceping.NewGenerator("node1", "0", "0")
		c.policyFactory = c.policyFactory.Register(roceping.Rule, roceGen)
		patchTime := gomonkey.ApplyFunc(time.Sleep, func(_ time.Duration) {
			fmt.Print("mock one second pass")
		})
		defer patchTime.Reset()
		convey.Convey("01--Testing GeneratePingPolicy failed because of timeout", func() {
			patch := gomonkey.ApplyMethod(&roceping.GeneratorImp{}, "Generate",
				func(_ *roceping.GeneratorImp, _ map[string]types.SuperDeviceIDs) map[string]types.DestinationAddress {
					return nil
				})
			defer patch.Reset()
			err := c.generatePingPolicy()
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "generate roce ping list info timeout")
		})
		convey.Convey("02--Testing GeneratePingPolicy success", func() {
			patch := gomonkey.ApplyMethod(&roceping.GeneratorImp{}, "Generate",
				func(_ *roceping.GeneratorImp, _ map[string]types.SuperDeviceIDs) map[string]types.DestinationAddress {
					return make(map[string]types.DestinationAddress)
				})
			defer patch.Reset()
			err := c.generatePingPolicy()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func mockGetDeviceTypeA5() func() {
	patch1 := gomonkey.ApplyMethod(&executor.DevManager{}, "GetDeviceType", func(_ *executor.DevManager) string {
		return common.Ascend910A5
	})
	patch2 := gomonkey.ApplyMethod(&roceping.PingManager{}, "GetDevType", func(_ *roceping.PingManager) string {
		return common.Ascend910A5
	})
	return func() {
		patch1.Reset()
		patch2.Reset()
	}
}
