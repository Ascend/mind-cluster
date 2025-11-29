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

/*
Package node funcs about node.
*/
package node

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/api"
)

const (
	resultLen = 2
)

func TestGetNodeAcceleratorType(t *testing.T) {
	convey.Convey("Test getNodeAcceleratorType function", t, func() {
		convey.Convey("When node has accelerator type label", func() {
			node := &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Labels: map[string]string{
						api.AcceleratorTypeKey: "ascend800ia5x8",
					},
				},
			}

			convey.Convey("Should return the correct accelerator type", func() {
				result := getNodeAcceleratorType(node)
				convey.So(result, convey.ShouldEqual, "ascend800ia5x8")
			})
		})

		convey.Convey("When node doesn't have accelerator type label", func() {
			node := &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test-node",
					Labels: map[string]string{}, // empty labels
				},
			}

			convey.Convey("Should return empty string", func() {
				result := getNodeAcceleratorType(node)
				convey.So(result, convey.ShouldBeEmpty)
			})
		})

		convey.Convey("When node has labels but no accelerator type", func() {
			node := &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Labels: map[string]string{
						"otherLabel": "value",
					},
				},
			}

			convey.Convey("Should return empty string", func() {
				result := getNodeAcceleratorType(node)
				convey.So(result, convey.ShouldBeEmpty)
			})
		})
	})
}

func TestGetRackID(t *testing.T) {
	convey.Convey("Test getRackID", t, func() {
		testNode := &v1.Node{}
		convey.Convey("When node is Ascend800ia5SuperPod type", func() {
			patches := gomonkey.ApplyFunc(getNodeAcceleratorType, func(*v1.Node) string {
				return api.Ascend800ia5SuperPod
			})
			defer patches.Reset()
			result := getRackID(testNode)
			convey.So(result, convey.ShouldEqual, "0")
		})
		convey.Convey("When node is normal A5 device with valid rackID", func() {
			patches := gomonkey.ApplyFunc(getNodeAcceleratorType, func(*v1.Node) string {
				return api.A5PodType
			})
			defer patches.Reset()
			patches.ApplyFunc(getRackIdFromNode, func(*v1.Node) (string, error) {
				return "1", nil
			})
			patches.ApplyFunc(getDeviceType, func(*v1.Node) string {
				return "A5"
			})
			result := getRackID(testNode)
			convey.So(result, convey.ShouldEqual, "1")
		})
	})
}

func TestGetRackIdFromNode(t *testing.T) {
	convey.Convey("Test getRackIdFromNode", t, func() {
		convey.Convey("When node has valid rackID annotation", func() {
			node := &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Annotations: map[string]string{
						api.RackIDKey: " 1 ",
					},
				},
			}
			rackID, err := getRackIdFromNode(node)
			convey.So(err, convey.ShouldBeNil)
			convey.So(rackID, convey.ShouldEqual, "1")
		})
		convey.Convey("When node has empty rackID annotation", func() {
			node := &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node",
					Annotations: map[string]string{
						api.RackIDKey: "",
					},
				},
			}
			rackID, err := getRackIdFromNode(node)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(rackID, convey.ShouldBeEmpty)
		})
	})
}

func TestGetNodeDeviceA5(t *testing.T) {
	resetCache()
	convey.Convey("test func getNodeDevice failed, baseDevInfos is nil", t, func() {
		nodeDev := getNodeDeviceA5(nil, nodeName1, "", "", "")
		convey.So(nodeDev, convey.ShouldBeNil)
	})

	convey.Convey("test func getNodeDevice failed, illegal device name", t, func() {
		baseDevInfos := map[string]*api.NpuBaseInfo{
			invalidDevName: {
				IP:            ip0,
				SuperDeviceID: superPodID,
			},
		}
		nodeDev := getNodeDeviceA5(baseDevInfos, nodeName1, "", "", "")
		convey.So(nodeDev, convey.ShouldBeNil)
	})
}
