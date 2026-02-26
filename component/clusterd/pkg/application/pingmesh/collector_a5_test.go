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
Package pingmesh a series of function handle ping mesh configmap create/update/delete.
*/
package pingmesh

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"ascend-common/api"
)

func TestNodeCollectorInvalidNode(t *testing.T) {
	convey.Convey("Testing NodeCollector with Invalid node", t, func() {
		nodeDevice := &api.NodeDevice{
			NodeName:        testNodeName,
			ServerType:      api.VersionNPU,
			AcceleratorType: api.Ascend800ia5Stacking,
			DeviceMap: map[string]string{
				"device1": "superDevice1",
			},
		}
		result := isValidDeviceType(nodeDevice)
		convey.So(result, convey.ShouldBeFalse)
	})
}

func TestNodeCollectorValidNode(t *testing.T) {
	convey.Convey("Testing NodeCollector with Valid node", t, func() {
		nodeDevice := &api.NodeDevice{
			NodeName:        testNodeName,
			ServerType:      api.VersionNPU,
			AcceleratorType: api.Ascend800ia5SuperPod,
			DeviceMap: map[string]string{
				"device1": "superDevice1",
			},
		}
		result := isValidDeviceType(nodeDevice)
		convey.So(result, convey.ShouldBeTrue)
	})
}

func TestNodeCollectorValidNode2(t *testing.T) {
	convey.Convey("Testing NodeCollector with Valid node 2", t, func() {
		nodeDevice := &api.NodeDevice{
			NodeName:   testNodeName,
			ServerType: api.VersionA3,
			DeviceMap: map[string]string{
				"device1": "superDevice1",
			},
		}
		result := isValidDeviceType(nodeDevice)
		convey.So(result, convey.ShouldBeTrue)
	})
}
