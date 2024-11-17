/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

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
Package rescheduling is using for HuaWei Ascend pin fault rescheduling.
*/
package rescheduling

import (
	"reflect"
	"testing"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/plugin"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

type FaultNodeGetNodeHeartbeatFromDeviceInfoArgs struct {
	node *plugin.NPUNode
}

const (
	eight = 8
	ten   = 10
)

type FaultNodeGetNodeHeartbeatFromDeviceInfoTests struct {
	fields  *FaultNode
	name    string
	args    FaultNodeGetNodeHeartbeatFromDeviceInfoArgs
	want    int64
	wantErr bool
}

func buildFaultGetNodeHeartbeatFromDeviceInfoTests() []FaultNodeGetNodeHeartbeatFromDeviceInfoTests {
	test1 := FaultNodeGetNodeHeartbeatFromDeviceInfoTests{
		name:   "01-FaultNodeUpdateFaultNodesFromDeviceInfoTests() nil device info",
		fields: fakeTestFaultNodeNodeHealthy("node0"),
		args: FaultNodeGetNodeHeartbeatFromDeviceInfoArgs{
			node: fakeNPUNodeNilDeviceInfo("node0"),
		},
		want:    zero,
		wantErr: true,
	}
	test2 := FaultNodeGetNodeHeartbeatFromDeviceInfoTests{
		name:   "01-FaultNodeUpdateFaultNodesFromDeviceInfoTests() succeed",
		fields: fakeTestFaultNodeNodeHealthy("node0"),
		args: FaultNodeGetNodeHeartbeatFromDeviceInfoArgs{
			node: fakeNPUNodeWithDeviceInfo("node0"),
		},
		want:    eight,
		wantErr: false,
	}
	tests := []FaultNodeGetNodeHeartbeatFromDeviceInfoTests{
		test1,
		test2,
	}
	return tests
}

// TestFaultNodeGetNodeHeartbeatFromDeviceInfo get node heartbeat
func TestFaultNodeGetNodeHeartbeatFromDeviceInfo(t *testing.T) {
	tests := buildFaultGetNodeHeartbeatFromDeviceInfoTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fNode := tt.fields
			got, err := fNode.getNodeHeartbeatFromNodeDInfo(tt.args.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("getNodeHeartbeatFromDeviceInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getNodeHeartbeatFromDeviceInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type FaultNodeGetNodeHeartbeatIntervalFromDeviceInfoArgs struct {
	node *plugin.NPUNode
}

type FaultNodeGetNodeHeartbeatIntervalFromDeviceInfoTests struct {
	fields  *FaultNode
	name    string
	args    FaultNodeGetNodeHeartbeatIntervalFromDeviceInfoArgs
	want    int
	wantErr bool
}

func buildFaultNodeGetNodeHeartbeatIntFromDeviceInfoTests() []FaultNodeGetNodeHeartbeatIntervalFromDeviceInfoTests {
	test1 := FaultNodeGetNodeHeartbeatIntervalFromDeviceInfoTests{
		name:   "01-FaultNodeGetNodeHeartbeatIntervalFromDeviceInfoTests() nil device info",
		fields: fakeTestFaultNodeNodeHealthy("node0"),
		args: FaultNodeGetNodeHeartbeatIntervalFromDeviceInfoArgs{
			node: fakeNPUNodeNilDeviceInfo("node0"),
		},
		want:    nodeUpdateTime,
		wantErr: true,
	}
	test2 := FaultNodeGetNodeHeartbeatIntervalFromDeviceInfoTests{
		name:   "02-FaultNodeGetNodeHeartbeatIntervalFromDeviceInfoTests() succeed",
		fields: fakeTestFaultNodeNodeHealthy("node0"),
		args: FaultNodeGetNodeHeartbeatIntervalFromDeviceInfoArgs{
			node: fakeNPUNodeWithDeviceInfo("node0"),
		},
		want:    ten,
		wantErr: false,
	}
	tests := []FaultNodeGetNodeHeartbeatIntervalFromDeviceInfoTests{
		test1,
		test2,
	}
	return tests
}

// TestFaultNodeGetNodeHeartbeatIntervalFromDeviceInfo test for get node heartbeat
func TestFaultNodeGetNodeHeartbeatIntervalFromDeviceInfo(t *testing.T) {
	tests := buildFaultNodeGetNodeHeartbeatIntFromDeviceInfoTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fNode := tt.fields
			got, err := fNode.getNodeHeartbeatIntervalFromNodeDInfo(tt.args.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("getNodeHeartbeatIntervalFromDeviceInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getNodeHeartbeatIntervalFromDeviceInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type FaultNodeGetAllNPUCardsFromDeviceInfoArgs struct {
	node     *plugin.NPUNode
	cardName string
}

type FaultNodeGetAllNPUCardsFromDeviceInfoTests struct {
	fields  *FaultNode
	name    string
	args    FaultNodeGetAllNPUCardsFromDeviceInfoArgs
	want    []string
	wantErr bool
}

func buildFaultNodeGetAllNPUCardsFromDeviceInfoTests() []FaultNodeGetAllNPUCardsFromDeviceInfoTests {
	node2 := fakeNPUNodeWithDeviceInfo("node0")
	test1 := FaultNodeGetAllNPUCardsFromDeviceInfoTests{
		name:   "01-FaultNodeGetNodeHeartbeatIntervalFromDeviceInfoTests() nil device info",
		fields: fakeTestFaultNodeNodeHealthy("node0"),
		args: FaultNodeGetAllNPUCardsFromDeviceInfoArgs{
			node:     fakeNPUNodeNilDeviceInfo("node0"),
			cardName: util.NPU910CardName,
		},
		want:    []string{},
		wantErr: true,
	}
	test2 := FaultNodeGetAllNPUCardsFromDeviceInfoTests{
		name:   "02-FaultNodeGetNodeHeartbeatIntervalFromDeviceInfoTests() succeed",
		fields: fakeTestFaultNodeNodeHealthy("node0"),
		args: FaultNodeGetAllNPUCardsFromDeviceInfoArgs{
			node:     node2,
			cardName: util.NPU910CardName,
		},
		want:    []string{"Ascend910-0", "Ascend910-1", "Ascend910-2"},
		wantErr: false,
	}
	tests := []FaultNodeGetAllNPUCardsFromDeviceInfoTests{
		test1,
		test2,
	}
	return tests
}

// TestFaultNodeGetAllNPUCardsFromDeviceInfo test for get npu card
func TestFaultNodeGetAllNPUCardsFromDeviceInfo(t *testing.T) {
	tests := buildFaultNodeGetAllNPUCardsFromDeviceInfoTests()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fNode := tt.fields
			got, err := fNode.getAllNPUCardsFromDeviceInfo(tt.args.node, tt.args.cardName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAllNPUCardsFromDeviceInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getAllNPUCardsFromDeviceInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}
