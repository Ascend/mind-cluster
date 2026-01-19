// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault test for public fault util
package publicfault

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"

	"ascend-common/api"
	"clusterd/pkg/common/constant"
)

func TestParsePubFaultCM(t *testing.T) {
	faultData, err := json.Marshal(testFaultInfo)
	if err != nil {
		t.Error(err)
	}
	convey.Convey("test func ParsePubFaultCM success", t, func() {
		cm := v1.ConfigMap{
			Data: map[string]string{api.PubFaultCMDataKey: string(faultData)},
		}
		_, err = ParsePubFaultCM(&cm)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("test func ParsePubFaultCM failed, input type error", t, func() {
		_, err = ParsePubFaultCM(nil)
		convey.So(err, convey.ShouldResemble, errors.New("input is not a valid cm"))
	})
	convey.Convey("test func ParsePubFaultCM failed, cm key error", t, func() {
		cm := v1.ConfigMap{
			Data: map[string]string{"error key": string(faultData)},
		}
		_, err = ParsePubFaultCM(&cm)
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("test func ParsePubFaultCM failed, unmarshal error", t, func() {
		p1 := gomonkey.ApplyFuncReturn(json.Unmarshal, testErr)
		defer p1.Reset()
		cm := v1.ConfigMap{
			Data: map[string]string{api.PubFaultCMDataKey: string(faultData)},
		}
		_, err = ParsePubFaultCM(&cm)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestDeepCopy(t *testing.T) {
	convey.Convey("test func DeepCopy success", t, func() {
		res := DeepCopy(&testFaultInfo)
		convey.So(res, convey.ShouldResemble, &testFaultInfo)
	})
	convey.Convey("test func DeepCopy failed, input is nil", t, func() {
		res := DeepCopy(nil)
		convey.So(res, convey.ShouldBeNil)
	})
	convey.Convey("test func DeepCopy failed, marshal error", t, func() {
		p1 := gomonkey.ApplyFuncReturn(json.Marshal, nil, testErr)
		defer p1.Reset()
		res := DeepCopy(&testFaultInfo)
		convey.So(res, convey.ShouldBeNil)
	})
	convey.Convey("test func DeepCopy failed, unmarshal error", t, func() {
		p1 := gomonkey.ApplyFuncReturn(json.Unmarshal, testErr)
		defer p1.Reset()
		res := DeepCopy(&testFaultInfo)
		convey.So(res, convey.ShouldBeNil)
	})
}

func TestGetFaultLevelByCode(t *testing.T) {
	PubFaultCodeCfg.SeparateNPUCodes["001"] = struct{}{}
	PubFaultCodeCfg.SubHealthFaultCodes["002"] = struct{}{}
	PubFaultCodeCfg.NotHandleFaultCodes["003"] = struct{}{}
	convey.Convey("test func GetFaultLevelByCode, type is SeparateNPU", t, func() {
		level := GetFaultLevelByCode("001")
		convey.So(level, convey.ShouldEqual, constant.SeparateNPU)
	})
	convey.Convey("test func GetFaultLevelByCode, type is SubHealth", t, func() {
		level := GetFaultLevelByCode("002")
		convey.So(level, convey.ShouldEqual, constant.SubHealthFault)
	})
	convey.Convey("test func GetFaultLevelByCode, type is NotHandle", t, func() {
		level := GetFaultLevelByCode("003")
		convey.So(level, convey.ShouldEqual, constant.NotHandleFault)
	})
	convey.Convey("test func GetFaultLevelByCode, code is not defined", t, func() {
		level := GetFaultLevelByCode("")
		convey.So(level, convey.ShouldEqual, "")
	})
	PubFaultCodeCfg = pubFaultCodeCache{}
}
