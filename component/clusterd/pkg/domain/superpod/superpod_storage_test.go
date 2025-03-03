// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package superpod a series of cluster device info storage function
package superpod

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

const (
	node1        = "node1"
	node2        = "node2"
	device1      = "device1"
	device2      = "device2"
	superDevice1 = "superDevice1"
	superDevice2 = "superDevice2"
	superNode1   = "superNode1"
	superNode2   = "superNode2"
	numInt2      = 2
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

func clearHistory() {
	superPodManager.snMap = make(map[string]*api.SuperPodDevice, initSuperPodNum)
}

func TestDeepCopyNodeDevice(t *testing.T) {
	clearHistory()
	convey.Convey("Testing deepCopyNodeDevice", t, func() {
		original := &api.NodeDevice{
			NodeName: node1,
			DeviceMap: map[string]string{
				device1: superDevice1,
				device2: superDevice2,
			},
		}
		device := deepCopyNodeDevice(original)
		convey.So(device, convey.ShouldNotBeNil)
		convey.So(device.NodeName, convey.ShouldEqual, original.NodeName)
		convey.So(device.DeviceMap, convey.ShouldResemble, original.DeviceMap)

		device = deepCopyNodeDevice(nil)
		convey.So(device, convey.ShouldBeNil)
	})
}

func TestDeepCopySupperNodeDevice(t *testing.T) {
	clearHistory()
	convey.Convey("Testing deepCopySupperNodeDevice", t, func() {
		original := &api.SuperPodDevice{
			SuperPodID: superNode1,
			NodeDeviceMap: map[string]*api.NodeDevice{
				node1: {
					NodeName: node1,
					DeviceMap: map[string]string{
						device1: superDevice1,
					},
				},
			},
		}
		device := deepCopySuperPodDevice(original)
		convey.So(device, convey.ShouldNotBeNil)
		convey.So(device.SuperPodID, convey.ShouldEqual, original.SuperPodID)
		convey.So(len(device.NodeDeviceMap), convey.ShouldEqual, len(original.NodeDeviceMap))
		convey.So(device.NodeDeviceMap, convey.ShouldResemble, original.NodeDeviceMap)

		device = deepCopySuperPodDevice(nil)
		convey.So(device, convey.ShouldBeNil)
	})
}

func TestGetSupperNode(t *testing.T) {
	clearHistory()
	convey.Convey("Testing GetSupperNode", t, func() {
		supperNodeID := superNode1
		nodeDevice := &api.NodeDevice{
			NodeName: node1,
			DeviceMap: map[string]string{
				device1: superDevice1,
			},
		}
		SaveNode(supperNodeID, nodeDevice)

		retrieved := GetSuperPodDevice(supperNodeID)
		convey.So(retrieved, convey.ShouldNotBeNil)
		convey.So(retrieved.SuperPodID, convey.ShouldEqual, supperNodeID)
		convey.So(retrieved.NodeDeviceMap[node1].NodeName, convey.ShouldEqual, nodeDevice.NodeName)
		convey.So(retrieved.NodeDeviceMap[node1].DeviceMap, convey.ShouldResemble, nodeDevice.DeviceMap)
	})
}

func TestSaveNode(t *testing.T) {
	clearHistory()
	convey.Convey("Testing SaveNode", t, func() {
		supperNodeID := superNode2
		nodeDevice := &api.NodeDevice{
			NodeName: node2,
			DeviceMap: map[string]string{
				device1: superDevice1,
			},
		}
		SaveNode(supperNodeID, nodeDevice)

		retrieved := GetSuperPodDevice(supperNodeID)
		convey.So(retrieved, convey.ShouldNotBeNil)
		convey.So(retrieved.SuperPodID, convey.ShouldEqual, supperNodeID)
		convey.So(retrieved.NodeDeviceMap[node2].NodeName, convey.ShouldEqual, nodeDevice.NodeName)
		convey.So(retrieved.NodeDeviceMap[node2].DeviceMap, convey.ShouldResemble, nodeDevice.DeviceMap)
	})
}

func TestSaveNodeNil(t *testing.T) {
	clearHistory()
	convey.Convey("Testing SaveNode with nil node device", t, func() {
		SaveNode(superNode1, nil)
		retrieved := GetSuperPodDevice(superNode1)
		convey.So(retrieved, convey.ShouldBeNil)
	})
}

func TestSaveNodeEmptySupperNodeID(t *testing.T) {
	clearHistory()
	convey.Convey("Testing SaveNode with empty supperNodeID", t, func() {
		nodeDevice := &api.NodeDevice{
			NodeName: node1,
			DeviceMap: map[string]string{
				device1: superDevice1,
			},
		}
		SaveNode("", nodeDevice)
		retrieved := GetSuperPodDevice("")
		convey.So(retrieved, convey.ShouldBeNil)
	})
}

func TestDeleteNode(t *testing.T) {
	clearHistory()
	convey.Convey("Testing DeleteNode", t, func() {
		supperNodeID := superNode1
		nodeDevice5 := &api.NodeDevice{
			NodeName: node1,
			DeviceMap: map[string]string{
				device1: superDevice1,
			},
		}
		nodeDevice6 := &api.NodeDevice{
			NodeName: node2,
			DeviceMap: map[string]string{
				device1: superDevice1,
			},
		}
		SaveNode(supperNodeID, nodeDevice5)
		SaveNode(supperNodeID, nodeDevice6)

		DeleteNode(supperNodeID, nodeDevice5.NodeName)
		DeleteNode("", "")

		retrieved := GetSuperPodDevice(supperNodeID)
		convey.So(retrieved, convey.ShouldNotBeNil)
		DeleteNode(supperNodeID, nodeDevice6.NodeName)
		retrieved = GetSuperPodDevice(supperNodeID)
		convey.So(retrieved, convey.ShouldBeNil)
	})
}

func TestGetNodeDeviceAndSuperNodeID(t *testing.T) {
	clearHistory()
	convey.Convey("Testing GetNodeDeviceAndSuperNodeID", t, func() {
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: node1,
				Annotations: map[string]string{
					superPodIDKey: superNode1,
					deviceKey:     `{"Ascend910-1":{"IP":"192.168.1.1","SuperDeviceID":1}}`,
				},
			},
		}
		nodeDevice, supperNodeID := GetNodeDeviceAndSuperPodID(node)
		convey.So(nodeDevice, convey.ShouldNotBeNil)
		convey.So(supperNodeID, convey.ShouldEqual, superNode1)
		convey.So(nodeDevice.NodeName, convey.ShouldEqual, node1)
		convey.So(nodeDevice.DeviceMap["1"], convey.ShouldEqual, "1")
	})
}

func TestGetNodeDeviceAndSuperNodeIDWithEmptyNode(t *testing.T) {
	clearHistory()
	convey.Convey("Testing GetNodeDeviceAndSuperNodeID with empty node", t, func() {
		nodeDevice, supperNodeID := GetNodeDeviceAndSuperPodID(nil)
		convey.So(nodeDevice, convey.ShouldBeNil)
		convey.So(supperNodeID, convey.ShouldBeEmpty)
	})
}

func TestGetNodeDeviceAndSuperNodeIDWithEmptyName(t *testing.T) {
	clearHistory()
	convey.Convey("Testing GetNodeDeviceAndSuperNodeID with empty node name", t, func() {
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					superPodIDKey: superNode1,
					deviceKey:     `{"device1":{"IP":"192.168.1.1","SuperDeviceID":1}}`,
				},
			},
		}
		nodeDevice, supperNodeID := GetNodeDeviceAndSuperPodID(node)
		convey.So(nodeDevice, convey.ShouldBeNil)
		convey.So(supperNodeID, convey.ShouldBeEmpty)
	})
}

func TestGetNodeDeviceAndSuperNodeIDWithInvalidJSON(t *testing.T) {
	clearHistory()
	convey.Convey("Testing GetNodeDeviceAndSuperNodeID with invalid JSON", t, func() {
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
				Annotations: map[string]string{
					superPodIDKey: superNode1,
					deviceKey:     "invalid json",
				},
			},
		}
		nodeDevice, supperNodeID := GetNodeDeviceAndSuperPodID(node)
		convey.So(nodeDevice, convey.ShouldBeNil)
		convey.So(supperNodeID, convey.ShouldEqual, superNode1)
	})
}

func TestListClusterDevice(t *testing.T) {
	clearHistory()
	convey.Convey("Testing ListClusterDevice", t, func() {
		supperNodeID1 := superNode1
		nodeDevice1 := &api.NodeDevice{
			NodeName: node1,
			DeviceMap: map[string]string{
				device1: superDevice1,
			},
		}
		SaveNode(supperNodeID1, nodeDevice1)

		supperNodeID2 := superNode2
		nodeDevice2 := &api.NodeDevice{
			NodeName: node2,
			DeviceMap: map[string]string{
				device2: superDevice2,
			},
		}
		SaveNode(supperNodeID2, nodeDevice2)

		supperNodes := ListClusterDevice()
		convey.So(supperNodes, convey.ShouldHaveLength, numInt2)
	})
}
