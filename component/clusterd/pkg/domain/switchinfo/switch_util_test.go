// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.
// Package switchinfo for

package switchinfo

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
)

const (
	moreLength = 2
)

func TestMain(m *testing.M) {
	hwLogConfig := &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(hwLogConfig, nil)
	m.Run()
}

// TestParseSwitchInfoCM test
func TestParseSwitchInfoCM(t *testing.T) {
	convey.Convey("test parse switch info", t, func() {
		config := v1.ConfigMap{
			Data: map[string]string{api.SwitchInfoCMDataKey: "invalid"},
		}
		_, err := ParseSwitchInfoCM(&config)
		convey.So(err, convey.ShouldNotBeNil)
		config = v1.ConfigMap{
			Data: map[string]string{"invalid": "invalid"},
		}
		_, err = ParseSwitchInfoCM(&config)
		convey.So(err, convey.ShouldNotBeNil)
		swit := constant.SwitchFaultInfo{
			FaultCode:  []string{},
			FaultLevel: "FaultLevel",
			UpdateTime: 0,
			NodeStatus: "Healthy",
		}
		bytes, err := json.Marshal(swit)
		convey.So(err, convey.ShouldBeNil)
		config = v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: constant.SwitchInfoPrefix + "testName",
			},
			Data: map[string]string{api.SwitchInfoCMDataKey: string(bytes)},
		}
		_, err = ParseSwitchInfoCM(&config)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestDeepCopy test deep copy
func TestDeepCopy(t *testing.T) {
	convey.Convey("test deep copy", t, func() {
		FaultLevel := "NotHandle"
		info := constant.SwitchInfo{
			SwitchFaultInfo: constant.SwitchFaultInfo{
				FaultCode:  []string{},
				FaultLevel: FaultLevel,
				UpdateTime: time.Now().Unix(),
				NodeStatus: "Healthy",
			},
			CmName: "",
		}
		copied, err := DeepCopy(&info)
		convey.So(err, convey.ShouldBeNil)
		convey.So(copied.FaultLevel == FaultLevel, convey.ShouldBeTrue)
	})
}

// TestGetSafeData test get safe data
func TestGetSafeData(t *testing.T) {
	convey.Convey("Test Get Safe Datas", t, func() {
		switchInfos := map[string]*constant.SwitchInfo{}
		res := GetSafeData(switchInfos)
		convey.So(len(res) == 0, convey.ShouldBeTrue)
		switchInfos = map[string]*constant.SwitchInfo{"nodeName1": &constant.SwitchInfo{}}
		res = GetSafeData(switchInfos)
		convey.So(len(res) == len(switchInfos), convey.ShouldBeTrue)
		switchInfos = map[string]*constant.SwitchInfo{}
		for i := 0; i <= safeSwitchSize; i++ {
			switchInfos["nodeName"+strconv.Itoa(i)] = &constant.SwitchInfo{}
		}
		res = GetSafeData(switchInfos)
		convey.So(len(res) == moreLength, convey.ShouldBeTrue)
	})
}

// TestBusinessDataIsNotEqual Test business data is not equal
func TestBusinessDataIsNotEqual(t *testing.T) {
	convey.Convey("Test Get Safe Datas", t, func() {
		oldSwitch := constant.SwitchInfo{
			SwitchFaultInfo: constant.SwitchFaultInfo{
				FaultCode:  []string{},
				FaultLevel: "FaultLevel",
				NodeStatus: "Unhealthy",
			},
		}
		newSwitch := oldSwitch
		notSame := constant.SwitchInfoBusinessDataIsNotEqual(&oldSwitch, &newSwitch)
		convey.So(notSame, convey.ShouldBeFalse)
		oldSwitch.NodeStatus = "Healthy"
		notSame = constant.SwitchInfoBusinessDataIsNotEqual(&oldSwitch, &newSwitch)
		convey.So(notSame, convey.ShouldBeTrue)
	})
}
