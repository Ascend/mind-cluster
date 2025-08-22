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

// Package utils is a DT collection for func in network.go
package utils

import (
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"ascend-faultdiag-online/pkg/utils/constants"
)

const (
	loopbackIp = "127.0.0.1"
	mockEnvIp  = "192.168.1.100"
	mockNetIp  = "192.168.1.101"
	mockMask8  = 8
	mockMask24 = 24
	mockMask32 = 32
)

func TestGetNodeIp(t *testing.T) {
	convey.Convey("test GetNodeIp", t, func() {
		testGetNodeIpWithEnv()
		testGetNodeIpWithoutEnv()
		testGetNodeIpWithError()
		testGetNodeIpWithInvalidIp()
	})
}

func testGetNodeIpWithEnv() {
	convey.Convey("test GetNodeIp with env", func() {
		patch := gomonkey.ApplyFunc(os.Getenv, func(key string) string {
			if key == constants.XdlIpField {
				return mockEnvIp
			}
			return ""
		})
		defer patch.Reset()
		ip, err := GetNodeIp()
		convey.So(err, convey.ShouldBeNil)
		convey.So(ip, convey.ShouldEqual, mockEnvIp)
		// net.InterfaceAddrs() is valid and got the same ip
		patch.ApplyFunc(net.InterfaceAddrs, func() ([]net.Addr, error) {
			return []net.Addr{
				&net.IPNet{IP: net.ParseIP(loopbackIp), Mask: net.CIDRMask(mockMask8, mockMask32)},
				&net.IPNet{IP: net.ParseIP(mockNetIp), Mask: net.CIDRMask(mockMask24, mockMask32)},
			}, nil
		})
		ip, err = GetNodeIp()
		convey.So(err, convey.ShouldBeNil)
		convey.So(ip, convey.ShouldEqual, mockEnvIp)
	})
}

func testGetNodeIpWithoutEnv() {
	convey.Convey("test testGetNodeIpWithout env", func() {
		patch := gomonkey.ApplyFunc(os.Getenv, func(key string) string {
			return ""
		})
		patch.ApplyFunc(net.InterfaceAddrs, func() ([]net.Addr, error) {
			return []net.Addr{
				&net.IPNet{IP: net.ParseIP(loopbackIp), Mask: net.CIDRMask(mockMask8, mockMask32)},
				&net.IPNet{IP: net.ParseIP(mockNetIp), Mask: net.CIDRMask(mockMask24, mockMask32)},
			}, nil
		})
		defer patch.Reset()
		ip, err := GetNodeIp()
		convey.So(err, convey.ShouldBeNil)
		convey.So(ip, convey.ShouldEqual, mockNetIp)
	})
}

func testGetNodeIpWithError() {
	convey.Convey("test testGetNodeIp with error", func() {
		patch := gomonkey.ApplyFunc(os.Getenv, func(key string) string {
			return ""
		})
		patch.ApplyFunc(net.InterfaceAddrs, func() ([]net.Addr, error) {
			return []net.Addr{
				&net.IPNet{IP: net.ParseIP(loopbackIp), Mask: net.CIDRMask(mockMask8, mockMask32)},
			}, fmt.Errorf("no valid IP")
		})
		defer patch.Reset()
		ip, err := GetNodeIp()
		convey.So(err.Error(), convey.ShouldEqual, "no valid IP")
		convey.So(ip, convey.ShouldBeEmpty)
	})
}

func testGetNodeIpWithInvalidIp() {
	convey.Convey("test testGetNodeIp with invalid ip", func() {
		patch := gomonkey.ApplyFunc(os.Getenv, func(key string) string {
			return ""
		})
		patch.ApplyFunc(net.InterfaceAddrs, func() ([]net.Addr, error) {
			return []net.Addr{
				&net.IPNet{IP: net.ParseIP(loopbackIp), Mask: net.CIDRMask(mockMask8, mockMask32)},
			}, nil
		})
		defer patch.Reset()
		ip, err := GetNodeIp()
		convey.So(err.Error(), convey.ShouldEqual, "no valid IP address found")
		convey.So(ip, convey.ShouldBeEmpty)
	})
}

func TestGetClusterIp(t *testing.T) {
	convey.Convey("test GetClusterIp", t, func() {
		// env exists
		var ip = "127.0.0.1"
		patch := gomonkey.ApplyFunc(os.Getenv, func(key string) string {
			if key == constants.PodIP {
				return ip
			}
			return ""
		})
		defer convey.So(GetClusterIp(), convey.ShouldEqual, ip)
		patch.Reset()
		// env not exist
		convey.So(GetClusterIp(), convey.ShouldEqual, "")
	})
}
