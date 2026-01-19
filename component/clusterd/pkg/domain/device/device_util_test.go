// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package device a series of device test function
package device

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

var (
	testCmName           = "test-node-name"
	testDeviceCheckCode  = "aaa60c794e2dbec298a2f3c18ea64dea9a1fd2ccdb0cc577b8dfe2c3c5966965"
	testOneSafeStr       = 1000
	testTwoSafeStr       = 1001
	testTwoSafeStrLength = 2

	testDeviceKey1   = "key1"
	testDeviceValue1 = "value1"
	testDeviceKey2   = "key2"
	testDeviceValue2 = "value2"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

func TestParseDeviceInfoCM(t *testing.T) {
	convey.Convey("TestParseDeviceInfoCM", t, func() {
		convey.Convey("obj is nil", func() {
			_, err := ParseDeviceInfoCM(nil)
			convey.So(err.Error(), convey.ShouldEqual, "not device configmap")
		})
		convey.Convey("obj without DeviceInfoCfg key", func() {
			cm := &v1.ConfigMap{}
			cm.Name = testCmName
			_, err := ParseDeviceInfoCM(cm)
			convey.So(err.Error(), convey.ShouldEndWith, api.DeviceInfoCMDataKey)
		})
		convey.Convey("obj checkCode is not equal", func() {
			cm := &v1.ConfigMap{}
			cm.Name = testCmName
			devInfoCM := constant.DeviceInfoCM{}
			devInfoCM.CheckCode = ""
			devInfoCM.SuperPodID = 1
			devInfoCM.SuperPodID = 1
			devInfoCM.DeviceInfo = constant.DeviceInfoNoName{
				UpdateTime: 0,
			}
			cm.Data = map[string]string{}
			cm.Data[api.DeviceInfoCMDataKey] = util.ObjToString(devInfoCM)
			_, err := ParseDeviceInfoCM(cm)
			convey.So(err.Error(), convey.ShouldEqual, fmt.Sprintf("device info configmap %s is not valid", cm.Name))
		})
		convey.Convey("obj checkCode is equal", func() {
			cm := &v1.ConfigMap{}
			cm.Name = testCmName
			devInfoCM := constant.DeviceInfoCM{}
			devInfoCM.CheckCode = testDeviceCheckCode
			devInfoCM.SuperPodID = 1
			devInfoCM.SuperPodID = 1
			devInfoCM.DeviceInfo = constant.DeviceInfoNoName{
				UpdateTime: 0,
			}
			cm.Data = map[string]string{}
			cm.Data[api.DeviceInfoCMDataKey] = util.ObjToString(devInfoCM)
			_, err := ParseDeviceInfoCM(cm)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDeepCopy(t *testing.T) {
	convey.Convey("TestDeepCopy", t, func() {
		convey.Convey("info is nil", func() {
			deviceInfo := DeepCopy(nil)
			convey.So(deviceInfo, convey.ShouldEqual, nil)
		})
		convey.Convey("info is normal data", func() {
			device := &constant.DeviceInfo{}
			device.CmName = testCmName
			newDevice := DeepCopy(device)
			convey.So(device.CmName, convey.ShouldEqual, newDevice.CmName)
		})
	})
}

func TestGetSafeData(t *testing.T) {
	convey.Convey("TestGetSafeData", t, func() {
		convey.Convey("deviceInfos is nil", func() {
			arr := GetSafeData(nil)
			convey.So(len(arr), convey.ShouldEqual, 0)
		})
		convey.Convey("the length of deviceInfos is 1000", func() {
			deviceInfos := map[string]*constant.DeviceInfo{}
			for i := 0; i < testOneSafeStr; i++ {
				deviceInfos[strconv.Itoa(i)] = &constant.DeviceInfo{}
			}
			arr := GetSafeData(deviceInfos)
			convey.So(len(arr), convey.ShouldEqual, 1)
		})
		convey.Convey("the length of deviceInfos is 1001", func() {
			deviceInfos := map[string]*constant.DeviceInfo{}
			for i := 0; i < testTwoSafeStr; i++ {
				deviceInfos[strconv.Itoa(i)] = &constant.DeviceInfo{}
			}
			arr := GetSafeData(deviceInfos)
			convey.So(len(arr), convey.ShouldEqual, testTwoSafeStrLength)
		})
	})
}

func TestBusinessDataIsNotEqual(t *testing.T) {
	convey.Convey("TestBusinessDataIsEqual", t, func() {
		convey.Convey("data is nil", func() {
			result := constant.DeviceInfoBusinessDataIsNotEqual(nil, nil)
			convey.So(result, convey.ShouldEqual, false)
		})
		convey.Convey("oldDevInfo is nil and devInfo is not nil", func() {
			result := constant.DeviceInfoBusinessDataIsNotEqual(nil, &constant.DeviceInfo{})
			convey.So(result, convey.ShouldEqual, true)
		})
		convey.Convey("business data is not equal", func() {
			newData := getTestDeviceInfo(map[string]string{testDeviceKey1: testDeviceValue1})
			oldData := &constant.DeviceInfo{}
			result := constant.DeviceInfoBusinessDataIsNotEqual(oldData, newData)
			convey.So(result, convey.ShouldEqual, true)
		})
		convey.Convey("business data is equal, other data is not equal", func() {
			newData := getTestDeviceInfo(map[string]string{testDeviceKey1: testDeviceValue1})
			oldData := getTestDeviceInfo(map[string]string{testDeviceKey2: testDeviceValue2})
			result := constant.DeviceInfoBusinessDataIsNotEqual(oldData, newData)
			convey.So(result, convey.ShouldEqual, true)
		})
		convey.Convey("business data is equal, other data is equal", func() {
			newData := getTestDeviceInfo(map[string]string{testDeviceKey1: testDeviceValue1})
			oldData := getTestDeviceInfo(map[string]string{testDeviceKey1: testDeviceValue1})
			result := constant.DeviceInfoBusinessDataIsNotEqual(oldData, newData)
			convey.So(result, convey.ShouldEqual, false)
		})
	})
}

func getTestDeviceInfo(deviceList map[string]string) *constant.DeviceInfo {
	return &constant.DeviceInfo{
		DeviceInfoNoName: constant.DeviceInfoNoName{
			DeviceList: deviceList,
		},
	}
}
