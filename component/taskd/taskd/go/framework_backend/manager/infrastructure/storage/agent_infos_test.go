/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

package storage

import (
	"sync"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"

	"taskd/toolkit_backend/net/common"
)

func TestAgentInfos_SetAllStatusVal(t *testing.T) {
	convey.Convey("TestAgentInfos_SetAllStatusVal", t, func() {
		agentInfos := &AgentInfos{
			Agents:    make(map[string]*AgentInfo),
			AllStatus: make(map[string]string),
			RWMutex:   sync.RWMutex{},
		}
		convey.Convey("should set status for new agent", func() {
			err := agentInfos.SetAllStatusVal("agent1", "running")
			convey.So(err, convey.ShouldBeNil)
			convey.So(agentInfos.AllStatus["agent1"], convey.ShouldEqual, "running")
		})
	})
}

func TestAgentInfos_DeepCopy(t *testing.T) {
	convey.Convey("TestAgentInfos_DeepCopy", t, func() {
		convey.Convey("should deep copy agentInfos with agents", func() {
			agentInfos := &AgentInfos{
				Agents: map[string]*AgentInfo{
					"agent1": {
						Status:   map[string]string{"key": "val"},
						NodeRank: "0",
						Pos:      &common.Position{Role: "agent"},
						RWMutex:  sync.RWMutex{},
					},
					"agent2": nil,
				},
				AllStatus: map[string]string{"agent1": "running"},
				RWMutex:   sync.RWMutex{},
			}
			clone := agentInfos.DeepCopy()
			convey.So(clone.AllStatus["agent1"], convey.ShouldEqual, "running")
			convey.So(clone.Agents["agent1"].NodeRank, convey.ShouldEqual, "0")
			convey.So(clone.Agents["agent2"], convey.ShouldBeNil)
		})
	})
}

func TestAgentInfo_GetStatusVal(t *testing.T) {
	convey.Convey("TestAgentInfo_GetStatusVal", t, func() {
		agentInfo := &AgentInfo{
			Status:  map[string]string{"key1": "value1"},
			RWMutex: sync.RWMutex{},
		}
		convey.Convey("should return value for existing key", func() {
			val, ok := agentInfo.GetStatusVal("key1")
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(val, convey.ShouldEqual, "value1")
		})
	})
}

func TestAgentInfo_SetStatusVal(t *testing.T) {
	convey.Convey("TestAgentInfo_SetStatusVal", t, func() {
		agentInfo := &AgentInfo{
			Status:  map[string]string{},
			RWMutex: sync.RWMutex{},
		}
		agentInfo.SetStatusVal("key1", "value1")
		convey.So(agentInfo.Status["key1"], convey.ShouldEqual, "value1")
	})
}

func TestAgentInfo_DeepCopy(t *testing.T) {
	convey.Convey("TestAgentInfo_DeepCopy", t, func() {
		convey.Convey("should deep copy agentInfo with pos", func() {
			agentInfo := &AgentInfo{
				Config:    map[string]string{"c1": "v1"},
				Actions:   map[string]string{"a1": "v1"},
				Status:    map[string]string{"s1": "v1"},
				FaultInfo: map[string]string{"f1": "v1"},
				NodeRank:  "0",
				HeartBeat: time.Now(),
				Pos:       &common.Position{Role: "agent", ServerRank: "1", ProcessRank: "0"},
				RWMutex:   sync.RWMutex{},
			}
			clone := agentInfo.DeepCopy()
			convey.So(clone.NodeRank, convey.ShouldEqual, "0")
			convey.So(clone.Pos.Role, convey.ShouldEqual, "agent")
			convey.So(clone.Config["c1"], convey.ShouldEqual, "v1")
		})
	})
}
