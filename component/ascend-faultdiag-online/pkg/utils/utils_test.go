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

// Package utils provides some DTs
package utils

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/utils/constants"
)

func TestToFloat64(t *testing.T) {
	type testCase struct {
		origin   any
		expected float64
	}

	var testCases = []testCase{
		{
			origin:   float32(123),
			expected: 123,
		},
		{
			origin:   float64(123),
			expected: 123,
		},
		{
			origin:   123,
			expected: 123,
		},
		{
			origin:   "123.3333",
			expected: 123.3333,
		},
	}
	var defaultValue = float64(123)
	for _, tc := range testCases {
		if got := ToFloat64(tc.origin, defaultValue); got != tc.expected {
			t.Errorf("ToFloat64(%v, %v) = %v, expected %v", tc.origin, defaultValue, got, tc.expected)
		}
	}

	// test default value
	varaibles := []any{"test", map[string]string{"test": "test"}, []string{"test"}}
	for _, v := range varaibles {
		if got := ToFloat64(v, defaultValue); got != defaultValue {
			t.Errorf("ToFloat64(%v, %v) = %v, expected %v", v, defaultValue, got, defaultValue)
		}
	}
}

func TestToString(t *testing.T) {

	type testCase struct {
		origin   any
		expected string
	}
	var testCases = []testCase{
		{
			origin:   "123",
			expected: "123",
		},
		{
			origin:   "",
			expected: "",
		},
		{
			origin:   123.00,
			expected: "",
		},
		{
			origin:   []string{"1", "2", "3"},
			expected: "",
		},
		{
			origin:   111,
			expected: "",
		},
	}

	for _, tc := range testCases {
		if got := ToString(tc.origin); got != tc.expected {
			t.Errorf("ToString(%v) = %v, expected %v", tc.origin, got, tc.expected)
		}
	}
}

func TestCopyInstance(t *testing.T) {
	type testCase struct {
		origin   any
		expected string
	}
	// src为nil
	var valueNil any
	instanceNil, errNil := CopyInstance(valueNil)
	assert.Error(t, errNil)
	assert.Equal(t, errNil.Error(), "src cannot be nil")
	assert.Nil(t, instanceNil)

	// src不是指针对象
	testCases := testCase{
		origin:   "name",
		expected: "123",
	}
	instance, err := CopyInstance(testCases)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "copy instance src is not ptr")
	assert.Nil(t, instance)

	// src是指针对象，但为空
	var testCasesOne *testCase
	instanceOne, errOne := CopyInstance(testCasesOne)
	assert.Error(t, errOne)
	assert.Equal(t, errOne.Error(), "src ptr cannot be nil")
	assert.Nil(t, instanceOne)

	// 正常复制实例
	testCasesTwo := &testCases
	instanceTwo, notErr := CopyInstance(testCasesTwo)
	assert.Nil(t, notErr)
	assert.Equal(t, instanceTwo, testCasesTwo)
}

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

const (
	loopbackIp = "127.0.0.1"
	mockEnvIp  = "192.168.1.100"
	mockNetIp  = "192.168.1.101"
	mockMask8  = 8
	mockMask24 = 24
	mockMask32 = 32
)

// TestGetNodeIpExistAndIpValid测试GetNodeIp函数环境变量存在且net.InterfaceAddrds返回值有效
func TestGetNodeIpExistAndIpValid(t *testing.T) {
	// 打桩os.Getenv返回值
	patch := gomonkey.NewPatches()
	defer patch.Reset()

	// 测试环境变量存在的情况
	patch.ApplyFunc(os.Getenv, func(key string) string {
		if key == constants.XdlIpField {
			return mockEnvIp
		}
		return ""
	})

	ip, err := GetNodeIp()
	assert.NoError(t, err)
	assert.Equal(t, mockEnvIp, ip)

	// 打桩net.InterfaceAddrs返回值有效
	patch.ApplyFunc(net.InterfaceAddrs, func() ([]net.Addr, error) {
		return []net.Addr{
			&net.IPNet{IP: net.ParseIP(loopbackIp), Mask: net.CIDRMask(mockMask8, mockMask32)},
			&net.IPNet{IP: net.ParseIP(mockNetIp), Mask: net.CIDRMask(mockMask24, mockMask32)},
		}, nil
	})

	ip, err = GetNodeIp()
	assert.NoError(t, err)
	assert.Equal(t, mockEnvIp, ip)
}

// TestGetNodeIpNotExistAndIpValid测试GetNodeIp函数环境变量不存在且net.InterfaceAddrds返回值有效
func TestGetNodeIpNotExistAndIpValid(t *testing.T) {
	// 打桩os.Getenv返回值
	patch := gomonkey.NewPatches()
	defer patch.Reset()

	// 测试环境变量不存在的情况
	patch.ApplyFunc(os.Getenv, func(key string) string {
		return ""
	})

	// 打桩net.InterfaceAddrs返回值有效
	patch.ApplyFunc(net.InterfaceAddrs, func() ([]net.Addr, error) {
		return []net.Addr{
			&net.IPNet{IP: net.ParseIP(loopbackIp), Mask: net.CIDRMask(mockMask8, mockMask32)},
			&net.IPNet{IP: net.ParseIP(mockNetIp), Mask: net.CIDRMask(mockMask24, mockMask32)},
		}, nil
	})

	ip, err := GetNodeIp()
	assert.Equal(t, err, nil)
	assert.Equal(t, mockNetIp, ip)
}

// TestGetNodeIpExistAndInterfaceErr测试GetNodeIp函数环境变量不存在且net.InterfaceAddrds返回值无效
func TestGetNodeIpExistAndInterfaceErr(t *testing.T) {
	// 打桩os.Getenv返回值
	patch := gomonkey.NewPatches()
	defer patch.Reset()

	// 测试环境变量存在的情况
	patch.ApplyFunc(os.Getenv, func(key string) string {
		return ""
	})

	// 打桩net.InterfaceAddrs返回值无效
	patch.ApplyFunc(net.InterfaceAddrs, func() ([]net.Addr, error) {
		return []net.Addr{
			&net.IPNet{IP: net.ParseIP(loopbackIp), Mask: net.CIDRMask(mockMask8, mockMask32)},
		}, fmt.Errorf("no valid IP")
	})

	ip, err := GetNodeIp()
	assert.Error(t, err)
	assert.Equal(t, "", ip)
	assert.Contains(t, err.Error(), "no valid IP")
}

// TestGetNodeIpExistAndIpInvalid测试GetNodeIp函数环境变量不存在且net.InterfaceAddrds返回值无效
func TestGetNodeIpExistAndIpInvalid(t *testing.T) {
	// 打桩os.Getenv返回值
	patch := gomonkey.NewPatches()
	defer patch.Reset()

	// 测试环境变量存在的情况
	patch.ApplyFunc(os.Getenv, func(key string) string {
		return ""
	})

	// 打桩net.InterfaceAddrs返回值无效
	patch.ApplyFunc(net.InterfaceAddrs, func() ([]net.Addr, error) {
		return []net.Addr{
			&net.IPNet{IP: net.ParseIP(loopbackIp), Mask: net.CIDRMask(mockMask8, mockMask32)},
		}, nil
	})

	ip, err := GetNodeIp()
	assert.Error(t, err)
	assert.Equal(t, "", ip)
	assert.Contains(t, err.Error(), "no valid IP address found")
}

// TestGetClusterIpWithEnvExist测试GetCluserIP函数环境变量存在
func TestGetClusterIpWithEnvExist(t *testing.T) {
	// 打桩os.Getenv返回值
	patch := gomonkey.NewPatches()
	defer patch.Reset()

	// 测试环境变量存在的情况
	patch.ApplyFunc(os.Getenv, func(key string) string {
		if key == constants.PodIP {
			return mockEnvIp
		}
		return ""
	})

	ip := GetClusterIp()
	assert.Equal(t, mockEnvIp, ip)
}

// TestGetClusterIpWithEnvExist测试GetCluserIP函数环境变量不存在
func TestGetClusterIpWithEnvNotExist(t *testing.T) {
	// 打桩os.Getenv返回值
	patch := gomonkey.NewPatches()
	defer patch.Reset()

	// 测试环境变量存在的情况
	patch.ApplyFunc(os.Getenv, func(key string) string {
		return ""
	})

	ip := GetClusterIp()
	assert.Equal(t, "", ip)
}

func TestWriteAndReadUniqueId(t *testing.T) {
	convey.Convey("test write and isRestarted", t, func() {
		// no file, no restarted
		convey.So(IsRestarted(), convey.ShouldBeFalse)

		// write start info
		WriteStartInfo()
		convey.So(IsRestarted(), convey.ShouldBeTrue)
		// wait the restartInterval
		time.Sleep(constants.RestartInterval * time.Millisecond)
		convey.So(IsRestarted(), convey.ShouldBeFalse)
	})
}
