// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package node a series of node function
package node

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"

	"ascend-common/api"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
)

var (
	testCmName           = "test-node-name"
	testNodeCheckCode    = "4c97cddcb947bd707778eb50b0986a69768afc2ef3e4f351db0b92e9d07d1fed"
	testOneSafeStr       = 2000
	testTwoSafeStr       = 2001
	testTwoSafeStrLength = 2
)

func TestParseNodeInfoCM(t *testing.T) {
	convey.Convey("TestParseNodeInfoCM", t, func() {
		convey.Convey("obj is nil", func() {
			_, err := ParseNodeInfoCM(nil)
			convey.So(err.Error(), convey.ShouldEqual, "not node info configmap")
		})
		convey.Convey("obj without NodeInfo key, should return err", func() {
			cm := &v1.ConfigMap{}
			cm.Name = testCmName
			_, err := ParseNodeInfoCM(cm)
			convey.So(err.Error(), convey.ShouldEndWith, api.NodeInfoCMDataKey)
		})
		convey.Convey("obj checkCode is not equal, should return err", func() {
			cm := &v1.ConfigMap{}
			cm.Name = testCmName
			nodeInfoCM := constant.NodeInfoCM{}
			nodeInfoCM.CheckCode = ""
			nodeInfoCM.NodeInfo = constant.NodeInfoNoName{}
			cm.Data = map[string]string{}
			cm.Data[api.NodeInfoCMDataKey] = util.ObjToString(nodeInfoCM)
			_, err := ParseNodeInfoCM(cm)
			convey.So(err.Error(), convey.ShouldEqual, fmt.Sprintf("node info configmap %s is not valid", cm.Name))
		})
		convey.Convey("obj checkCode is equal, err should be nil", func() {
			cm := &v1.ConfigMap{}
			cm.Name = testCmName
			nodeInfoCM := constant.NodeInfoCM{}
			nodeInfoCM.CheckCode = testNodeCheckCode
			nodeInfoCM.NodeInfo = constant.NodeInfoNoName{}
			cm.Data = map[string]string{}
			cm.Data[api.NodeInfoCMDataKey] = util.ObjToString(nodeInfoCM)
			_, err := ParseNodeInfoCM(cm)
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
			node := &constant.NodeInfo{}
			node.CmName = testCmName
			newNode := DeepCopy(node)
			convey.So(newNode.CmName, convey.ShouldEqual, node.CmName)
		})
	})
}

func TestGetSafeData(t *testing.T) {
	convey.Convey("TestGetSafeData", t, func() {
		convey.Convey("nodeInfos is nil", func() {
			arr := GetSafeData(nil)
			convey.So(len(arr), convey.ShouldEqual, 0)
		})
		convey.Convey("the length of nodeInfos is 2000", func() {
			nodeInfos := map[string]*constant.NodeInfo{}
			for i := 0; i < testOneSafeStr; i++ {
				nodeInfos[strconv.Itoa(i)] = &constant.NodeInfo{}
			}
			arr := GetSafeData(nodeInfos)
			convey.So(len(arr), convey.ShouldEqual, 1)
		})
		convey.Convey("the length of deviceInfos is 2001", func() {
			nodeInfos := map[string]*constant.NodeInfo{}
			for i := 0; i < testTwoSafeStr; i++ {
				nodeInfos[strconv.Itoa(i)] = &constant.NodeInfo{}
			}
			arr := GetSafeData(nodeInfos)
			convey.So(len(arr), convey.ShouldEqual, testTwoSafeStrLength)
		})
	})
}

func TestBusinessDataIsNotEqual(t *testing.T) {
	convey.Convey("Test BusinessDataIsNotEqual", t, func() {
		convey.Convey("both oldNodeInfo and newNodeInfo are nil", func() {
			result := constant.NodeInfoBusinessDataIsNotEqual(nil, nil)
			convey.So(result, convey.ShouldEqual, false)
		})
		convey.Convey("oldNodeInfo is nil,newNodeInfo is not nil", func() {
			newData := getTestNodeInfo("", nil)
			result := constant.NodeInfoBusinessDataIsNotEqual(nil, newData)

			convey.So(result, convey.ShouldEqual, true)
		})
		convey.Convey("oldNodeInfo and newNodeInfo are not equal", func() {
			newData := getTestNodeInfo("unhealthy", nil)
			oldData := getTestNodeInfo("healthy", nil)
			result := constant.NodeInfoBusinessDataIsNotEqual(newData, oldData)
			convey.So(result, convey.ShouldEqual, true)
		})
		convey.Convey("oldNodeInfo and newNodeInfo are equal", func() {
			newData := getTestNodeInfo("unhealthy", nil)
			oldData := getTestNodeInfo("unhealthy", nil)
			result := constant.NodeInfoBusinessDataIsNotEqual(newData, oldData)
			convey.So(result, convey.ShouldEqual, false)
		})
	})
}

func getTestNodeInfo(status string, faultList []*constant.FaultDev) *constant.NodeInfo {
	return &constant.NodeInfo{
		NodeInfoNoName: constant.NodeInfoNoName{
			NodeStatus:   status,
			FaultDevList: faultList,
		},
	}
}
