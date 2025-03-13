// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package kube a series of kube test function
package kube

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/api"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/superpod"
)

var (
	testTwoDeviceFunc = 2
	testTwoNodeFunc   = 2
)

func TestStopInformer(t *testing.T) {
	convey.Convey("TestStopInformer", t, func() {
		convey.So(StopInformer, convey.ShouldNotPanic)
	})
}

func TestCleanFuncs(t *testing.T) {
	convey.Convey("TestCleanFuncs", t, func() {
		CleanFuncs()
		convey.So(len(cmDeviceFuncs), convey.ShouldEqual, 0)
		convey.So(len(cmNodeFuncs), convey.ShouldEqual, 0)
	})
}

func TestAddCmDeviceFunc(t *testing.T) {
	convey.Convey("TestAddCmDeviceFunc", t, func() {
		convey.Convey("add one device func", func() {
			AddCmDeviceFunc(constant.Resource, func(info *constant.DeviceInfo, info2 *constant.DeviceInfo, s string) {})
			convey.So(len(cmDeviceFuncs[constant.Resource]), convey.ShouldEqual, 1)
		})
		convey.Convey("add two device func", func() {
			AddCmDeviceFunc(constant.Resource, func(info *constant.DeviceInfo, info2 *constant.DeviceInfo, s string) {})
			convey.So(len(cmDeviceFuncs[constant.Resource]), convey.ShouldEqual, testTwoDeviceFunc)
		})
		convey.Convey("add two different business func", func() {
			AddCmDeviceFunc(constant.Statistics, func(info *constant.DeviceInfo, info2 *constant.DeviceInfo, s string) {})
			convey.So(len(cmDeviceFuncs), convey.ShouldEqual, testTwoDeviceFunc)
		})
	})
}

func TestAddCmNodeFunc(t *testing.T) {
	convey.Convey("TestAddCmNodeFunc", t, func() {
		convey.Convey("add one node func", func() {
			AddCmNodeFunc(constant.Resource, func(info *constant.NodeInfo, info2 *constant.NodeInfo, s string) {})
			convey.So(len(cmNodeFuncs[constant.Resource]), convey.ShouldEqual, 1)
		})
		convey.Convey("add two node func", func() {
			AddCmNodeFunc(constant.Resource, func(info *constant.NodeInfo, info2 *constant.NodeInfo, s string) {})
			convey.So(len(cmNodeFuncs[constant.Resource]), convey.ShouldEqual, testTwoNodeFunc)
		})
		convey.Convey("add two different business func", func() {
			AddCmNodeFunc(constant.Statistics, func(info *constant.NodeInfo, info2 *constant.NodeInfo, s string) {})
			convey.So(len(cmNodeFuncs), convey.ShouldEqual, testTwoNodeFunc)
		})
	})
}

func TestCheckConfigMapIsDeviceInfo(t *testing.T) {
	convey.Convey("test checkConfigMapIsNodeInfo", t, func() {
		var obj interface{}
		mockMatchedFalse := gomonkey.ApplyFunc(util.IsNSAndNameMatched, func(obj interface{},
			namespace string, namePrefix string) bool {
			return false
		})
		defer mockMatchedFalse.Reset()
		cmCheck := checkConfigMapIsDeviceInfo(obj)
		convey.So(cmCheck, convey.ShouldBeFalse)
	})
}

func TestCheckConfigMapIsNodeInfo(t *testing.T) {
	convey.Convey("test checkConfigMapIsNodeInfo", t, func() {
		var obj interface{}
		mockMatchedTrue := gomonkey.ApplyFunc(util.IsNSAndNameMatched, func(obj interface{},
			namespace string, namePrefix string) bool {
			return true
		})
		defer mockMatchedTrue.Reset()
		nodeCheck := checkConfigMapIsNodeInfo(obj)
		convey.So(nodeCheck, convey.ShouldBeTrue)
	})
}

func TestInitClusterDevice(t *testing.T) {
	convey.Convey("Test initClusterDevice", t, func() {
		fakeNodes := []*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node-1",
				},
			},
		}
		patchGetNodes := gomonkey.ApplyFunc(getNodesFromInformer, func() []*v1.Node {
			return fakeNodes
		})
		defer patchGetNodes.Reset()

		patchGetNodeDevice := gomonkey.ApplyFunc(superpod.GetNodeDeviceAndSuperPodID,
			func(node *v1.Node) (*api.NodeDevice, string) {
				return &api.NodeDevice{
					NodeName: node.Name,
				}, "test-superpod-id"
			})
		defer patchGetNodeDevice.Reset()

		var calledSaveNode bool
		patchSaveNode := gomonkey.ApplyFunc(superpod.SaveNode,
			func(superPodID string, node *api.NodeDevice) {
				calledSaveNode = true
				convey.So(superPodID, convey.ShouldEqual, "test-superpod-id")
				convey.So(node.NodeName, convey.ShouldEqual, "test-node-1")
			})
		defer patchSaveNode.Reset()

		initClusterDevice()
		convey.So(calledSaveNode, convey.ShouldBeTrue)
	})
}
