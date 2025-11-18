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

// Package proxy for taskd proxy backend
package proxy

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"taskd/toolkit_backend/net"
	"taskd/toolkit_backend/net/common"
)

var initNetworkFunc = net.InitNetwork
var originalNewProxyInstance func(config *common.TaskNetConfig, log *hwlog.CustomLogger) error

func init() {
	originalNewProxyInstance = newProxyInstance
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

// MockNet mock net
type MockNet struct {
	InitNetworkResult *net.NetInstance
	InitNetworkErr    error
}

func (m *MockNet) InitNetwork(config *common.TaskNetConfig) (*net.NetInstance, error) {
	return m.InitNetworkResult, m.InitNetworkErr
}

func TestNewProxyInstance(t *testing.T) {
	mockNet := &MockNet{
		InitNetworkResult: &net.NetInstance{},
		InitNetworkErr:    nil,
	}
	originalInitNetwork := initNetworkFunc
	defer func() {
		initNetworkFunc = originalInitNetwork
	}()
	initNetworkFunc = func(config *common.TaskNetConfig, log *hwlog.CustomLogger) (*net.NetInstance, error) {
		return mockNet.InitNetwork(config)
	}
	proxyConfig := &common.TaskNetConfig{
		Pos: common.Position{
			ProcessRank: "-1",
		},
	}

	convey.Convey("test newProxyInstance", t, func() {
		convey.Convey("newProxyInstance failed", func() {
			mockNet.InitNetworkErr = errors.New("test newProxyInstance failed")
			proxyInstance = nil
			customLog := hwlog.SetCustomLogger(hwlog.RunLog)
			err := newProxyInstance(proxyConfig, customLog)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestProxyInit(t *testing.T) {
	mockNet := &MockNet{
		InitNetworkResult: &net.NetInstance{},
		InitNetworkErr:    nil,
	}
	originalInitNetwork := initNetworkFunc
	defer func() {
		initNetworkFunc = originalInitNetwork
	}()
	initNetworkFunc = func(config *common.TaskNetConfig, log *hwlog.CustomLogger) (*net.NetInstance, error) {
		return mockNet.InitNetwork(config)
	}
	customLog := hwlog.SetCustomLogger(hwlog.RunLog)
	proxy := &proxyClient{
		proxyInfo:   &proxyInfo{},
		proxyLogger: customLog,
	}
	proxyConfig := &common.TaskNetConfig{
		Pos: common.Position{
			ProcessRank: "-1",
		},
	}

	convey.Convey("test init func", t, func() {

		convey.Convey("valid config failed", func() {
			invalidConfig := &common.TaskNetConfig{
				Pos: common.Position{
					ProcessRank: "0",
				},
			}
			err := proxy.initNetwork(invalidConfig)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("initnetwork failed", func() {
			mockNet.InitNetworkErr = errors.New("mock initnetwork failed")
			err := proxy.initNetwork(proxyConfig)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestInitProxy(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	config := &common.TaskNetConfig{
		Pos: common.Position{
			Role:       common.ProxyRole,
			ServerRank: "0",
		},
		ListenAddr:   localHost,
		UpstreamAddr: localHost,
		EnableTls:    false,
		TlsConf:      nil,
	}

	convey.Convey("when proxy config is nil, then init proxy should return nil", t, func() {
		convey.ShouldNotBeNil(InitProxy(nil))
	})

	convey.Convey("when init log error, then init proxy should return nil", t, func() {
		patches.ApplyFunc(hwlog.InitRunLogger, func(config *hwlog.LogConfig, ctx context.Context) error {
			return fmt.Errorf("init log error")
		})

		convey.ShouldNotBeNil(InitProxy(config))
	})

	convey.Convey("when init proxy network error, then init proxy should return nil", t, func() {
		patches.ApplyFunc(hwlog.InitRunLogger, func(config *hwlog.LogConfig, ctx context.Context) error {
			return nil
		})

		patches.ApplyFunc(newProxyInstance, func(proxyConfig *common.TaskNetConfig) error {
			return fmt.Errorf("init instance error")
		})
		convey.ShouldNotBeNil(InitProxy(config))
	})

	convey.Convey("when no error, then init proxy should return nil", t, func() {
		patches.ApplyFunc(newProxyInstance, func(proxyConfig *common.TaskNetConfig) error {
			return nil
		})
		convey.ShouldBeNil(InitProxy(config))
	})
}

func TestDestroyProxy(t *testing.T) {
	convey.Convey("DestroyProxy should be called", t, func() {
		proxyInstance = &proxyClient{}
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		called := false
		patches.ApplyPrivateMethod(proxyInstance, "destroyNet", func(*proxyClient) {
			called = true
		})
		DestroyProxy()
		convey.ShouldBeTrue(called)

		proxyInstance = nil
		called = false
		DestroyProxy()
		convey.ShouldBeTrue(called)
	})
}
