// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package pingmesh a series of function handle ping mesh configmap create/update/delete
package pingmesh

import (
	"context"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/node"
	"clusterd/pkg/domain/superpod"
)

const (
	superPodIDKey   = "superPodID"
	testNodeName    = "test-node"
	testDeviceKey   = `{"device1":{"IP":"192.168.1.1","SuperDeviceID":1}}`
	invalidOperator = "invalid-operator"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

func setupNode() *v1.Node {
	return &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNodeName,
			Annotations: map[string]string{
				superPodIDKey:       testSuperPodID,
				api.BaseDevInfoAnno: testDeviceKey,
			},
		},
	}
}

func setupGetNodeDeviceAndSuperPodIDPatches() *gomonkey.Patches {
	return gomonkey.ApplyFunc(node.GetNodeDeviceAndSuperPodID,
		func(node *v1.Node) (*api.NodeDevice, string) {
			if node == nil {
				return nil, ""
			}
			return &api.NodeDevice{
				NodeName: node.Name,
				DeviceMap: map[string]string{
					"device1": "superDevice1",
				},
			}, testSuperPodID
		})
}

func TestNodeCollectorAddOperator(t *testing.T) {
	convey.Convey("Testing NodeCollector with AddOperator", t, func() {
		patches := setupGetNodeDeviceAndSuperPodIDPatches()
		defer patches.Reset()

		var savedSuperPodID string
		var savedNode *api.NodeDevice
		saveNodePatches := gomonkey.ApplyFunc(superpod.SaveNode,
			func(superPodID string, node *api.NodeDevice) {
				savedSuperPodID = superPodID
				savedNode = node
			})
		defer saveNodePatches.Reset()

		node := setupNode()
		oldEventMapLen := len(publishMgr.eventMap)
		NodeCollector(nil, node, constant.AddOperator)

		convey.So(savedSuperPodID, convey.ShouldEqual, testSuperPodID)
		convey.So(savedNode, convey.ShouldNotBeNil)
		convey.So(len(publishMgr.eventMap), convey.ShouldEqual, oldEventMapLen+1)
		convey.So(publishMgr.eventMap[testSuperPodID], convey.ShouldEqual, constant.UpdateOperator)
	})
}

func TestNodeCollectorUpdateOperator(t *testing.T) {
	convey.Convey("Testing NodeCollector with UpdateOperator", t, func() {
		patches := setupGetNodeDeviceAndSuperPodIDPatches()
		defer patches.Reset()

		var savedSuperPodID string
		var savedNode *api.NodeDevice
		saveNodePatches := gomonkey.ApplyFunc(superpod.SaveNode,
			func(superPodID string, node *api.NodeDevice) {
				savedSuperPodID = superPodID
				savedNode = node
			})
		defer saveNodePatches.Reset()

		node := setupNode()
		oldEventMapLen := len(publishMgr.eventMap)
		NodeCollector(nil, node, constant.UpdateOperator)

		convey.So(savedSuperPodID, convey.ShouldEqual, testSuperPodID)
		convey.So(savedNode, convey.ShouldNotBeNil)
		convey.So(len(publishMgr.eventMap), convey.ShouldEqual, oldEventMapLen)
		convey.So(publishMgr.eventMap[testSuperPodID], convey.ShouldEqual, constant.UpdateOperator)
	})
}

func TestNodeCollectorDeleteOperator(t *testing.T) {
	convey.Convey("Testing NodeCollector with DeleteOperator", t, func() {
		patches := setupGetNodeDeviceAndSuperPodIDPatches()
		defer patches.Reset()

		var deletedSuperPodID string
		var deletedNodeName string
		deleteNodePatches := gomonkey.ApplyFunc(superpod.DeleteNode,
			func(superPodID string, nodeName string) {
				deletedSuperPodID = superPodID
				deletedNodeName = nodeName
			})
		defer deleteNodePatches.Reset()

		node := setupNode()
		oldEventMapLen := len(publishMgr.eventMap)
		NodeCollector(nil, node, constant.DeleteOperator)

		convey.So(deletedSuperPodID, convey.ShouldEqual, testSuperPodID)
		convey.So(deletedNodeName, convey.ShouldEqual, testNodeName)
		convey.So(len(publishMgr.eventMap), convey.ShouldEqual, oldEventMapLen)
		convey.So(publishMgr.eventMap[testSuperPodID], convey.ShouldEqual, constant.DeleteOperator)
	})
}

func TestNodeCollectorInvalidOperator(t *testing.T) {
	convey.Convey("Testing NodeCollector with Invalid Operator", t, func() {
		patches := setupGetNodeDeviceAndSuperPodIDPatches()
		defer patches.Reset()

		node := setupNode()
		oldEventMapLen := len(publishMgr.eventMap)
		NodeCollector(nil, node, invalidOperator)

		convey.So(len(publishMgr.eventMap), convey.ShouldEqual, oldEventMapLen)
	})
}

func TestNodeCollectorParseInvalidNode(t *testing.T) {
	convey.Convey("Testing NodeCollector with Invalid node", t, func() {
		patches := gomonkey.ApplyFunc(node.GetNodeDeviceAndSuperPodID,
			func(node *v1.Node) (*api.NodeDevice, string) {
				return nil, ""
			})
		defer patches.Reset()

		node := setupNode()
		oldEventMapLen := len(publishMgr.eventMap)
		NodeCollector(nil, node, constant.AddOperator)

		convey.So(len(publishMgr.eventMap), convey.ShouldEqual, oldEventMapLen)
	})
}
