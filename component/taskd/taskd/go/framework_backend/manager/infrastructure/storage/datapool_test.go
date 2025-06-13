/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package storage for taskd manager backend data type
package storage

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"taskd/common/constant"
	"taskd/toolkit_backend/net/common"
)

// All mock agent,cluster,worker
var (
	agentName = "agent1"
	agentInfo = &AgentInfo{
		Status:    map[string]string{"status1": "value1"},
		NodeRank:  "1",
		HeartBeat: time.Now(),
		Pos:       &common.Position{Role: "agent"},
		RWMutex:   sync.RWMutex{},
	}
	workerName = "worker1"
	workerInfo = &WorkerInfo{
		Status:     map[string]string{"status1": "value1"},
		GlobalRank: "1",
		HeartBeat:  time.Now(),
		Pos:        &common.Position{Role: "worker"},
		RWMutex:    sync.RWMutex{},
	}
	clusterName = "cluster1"
	clusterInfo = &ClusterInfo{
		Command:   map[string]string{"cmd1": "value1"},
		HeartBeat: time.Now(),
		Business:  []int32{0, 0, 0},
		RWMutex:   sync.RWMutex{},
	}
	testPos = &common.Position{
		Role:        "testRole",
		ServerRank:  "1",
		ProcessRank: "0",
	}
)

// TestMain test main
func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		return
	}
	code := m.Run()
	fmt.Printf("exit_code = %v\n", code)
}

func setup() error {
	return initLog()
}

func initLog() error {
	logConfig := &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := hwlog.InitRunLogger(logConfig, context.Background()); err != nil {
		fmt.Printf("init hwlog failed, %v\n", err)
		return err
	}
	return nil
}

func newMsgQueue(length int32) *MsgQueue {
	if length >= 0 {
		return &MsgQueue{Queue: make([]BaseMessage, length), Mutex: sync.Mutex{}}
	}
	return nil
}

func newDataPool() *DataPool {
	return &DataPool{
		Snapshot: &SnapShot{
			AgentInfos: &AgentInfos{
				Agents:    make(map[string]*AgentInfo),
				AllStatus: make(map[string]string),
				RWMutex:   sync.RWMutex{},
			},
			WorkerInfos: &WorkerInfos{
				Workers:   make(map[string]*WorkerInfo),
				AllStatus: make(map[string]string),
				RWMutex:   sync.RWMutex{},
			},
			ClusterInfos: &ClusterInfos{
				Clusters:  make(map[string]*ClusterInfo),
				AllStatus: make(map[string]string),
				RWMutex:   sync.RWMutex{},
			},
		},
		RWMutex: sync.RWMutex{},
	}
}

// TestNewMsg test new a message
func TestNewMsg(t *testing.T) {
	mq := newMsgQueue(0)
	convey.Convey("NewMsg should create valid message", t, func() {
		msg := mq.NewMsg("test123", "TEST_BIZ", &common.Position{Role: "test"}, MsgBody{MsgType: "TEST"})
		convey.So(msg.Header.Uuid, convey.ShouldEqual, "test123")
		convey.So(msg.Header.BizType, convey.ShouldEqual, "TEST_BIZ")
		convey.So(msg.Body.MsgType, convey.ShouldEqual, "TEST")
		convey.So(msg.Header.Timestamp.IsZero(), convey.ShouldBeFalse)
	})
}

// TestEnqueue test message enter the queue
func TestEnqueue(t *testing.T) {
	convey.Convey("TestEnqueue test message enter the queue success", t, func() {
		mq := newMsgQueue(0)
		msg := BaseMessage{}
		err := mq.Enqueue(msg)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("TestEnqueue test message fail", t, func() {
		mq := newMsgQueue(constant.MaxMsgQueueLength)
		msg := BaseMessage{}
		err := mq.Enqueue(msg)
		convey.So(err.Error(), convey.ShouldContainSubstring, "full")
	})
}

// TestDequeue test message departure
func TestDequeue(t *testing.T) {
	convey.Convey("TestDequeue test message departure success", t, func() {
		msg := BaseMessage{Header: MsgHeader{}, Body: MsgBody{}}
		mq := &MsgQueue{
			Queue: []BaseMessage{msg},
			Mutex: sync.Mutex{},
		}
		dqMsg, err := mq.Dequeue()
		convey.So(dqMsg, convey.ShouldEqual, msg)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestRegisterAgent test register agent in the data pool
func TestRegisterAgent(t *testing.T) {
	dp := newDataPool()
	convey.Convey("TestRegisterAgent should add new agent", t, func() {
		err := dp.RegisterAgent(agentName, agentInfo)
		convey.So(err, convey.ShouldBeNil)
		agent, err := dp.GetAgent(agentName)
		convey.So(err, convey.ShouldBeNil)
		convey.So(agent.NodeRank, convey.ShouldEqual, "1")
		convey.So(agent.Status["status1"], convey.ShouldEqual, "value1")
	})
}

// TestRegisterWorker test register worker in the data pool
func TestRegisterWorker(t *testing.T) {
	dp := newDataPool()
	convey.Convey("TestRegisterWorker should add new worker", t, func() {
		err := dp.RegisterWorker(workerName, workerInfo)
		convey.So(err, convey.ShouldBeNil)
		worker, err := dp.GetWorker(workerName)
		convey.So(err, convey.ShouldBeNil)
		convey.So(worker.GlobalRank, convey.ShouldEqual, "1")
		convey.So(worker.Status["status1"], convey.ShouldEqual, "value1")
	})
}

// TestRegisterCluster test register cluster in the data pool
func TestRegisterCluster(t *testing.T) {
	dp := newDataPool()
	convey.Convey("RegisterCluster should add new cluster", t, func() {
		_ = dp.RegisterCluster(clusterName)

		cluster, err := dp.GetCluster(clusterName)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(cluster.Command), convey.ShouldEqual, 0)
		convey.So(cluster.Business, convey.ShouldResemble, []int32{})
	})
}

// TestUpdateAgent test update agent info in the data pool
func TestUpdateAgent(t *testing.T) {
	dp := newDataPool()
	convey.Convey("UpdateAgent should modify existing agent", t, func() {
		_ = dp.RegisterAgent(agentName, agentInfo)
		updatedAgent := &AgentInfo{
			Status:    map[string]string{"status1": "updated"},
			NodeRank:  "2",
			HeartBeat: time.Now(),
			Pos:       &common.Position{Role: "agent"},
			RWMutex:   sync.RWMutex{},
		}
		err := dp.UpdateAgent(agentName, updatedAgent)
		convey.So(err, convey.ShouldBeNil)
		agent, _ := dp.GetAgent(agentName)
		convey.So(agent.NodeRank, convey.ShouldEqual, "2")
		convey.So(agent.Status["status1"], convey.ShouldEqual, "updated")
	})
	convey.Convey("UpdateAgent should fail with nil data pool", t, func() {
		nilDp := &DataPool{}
		err := nilDp.UpdateAgent(agentName, agentInfo)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "not initialized")
	})
}

// TestUpdateWorker test update worker info in the data pool
func TestUpdateWorker(t *testing.T) {
	dp := newDataPool()
	convey.Convey("UpdateWorker should modify existing worker", t, func() {
		_ = dp.RegisterWorker(workerName, workerInfo)
		updatedWorker := &WorkerInfo{
			Status:     map[string]string{"status1": "updated"},
			GlobalRank: "2",
			HeartBeat:  time.Now(),
			Pos:        &common.Position{Role: "worker"},
			RWMutex:    sync.RWMutex{},
		}
		err := dp.UpdateWorker(workerName, updatedWorker)
		convey.So(err, convey.ShouldBeNil)
		worker, _ := dp.GetWorker(workerName)
		convey.So(worker.GlobalRank, convey.ShouldEqual, "2")
		convey.So(worker.Status["status1"], convey.ShouldEqual, "updated")
	})
	convey.Convey("UpdateWorker should fail with nil data pool", t, func() {
		var nilDp *DataPool
		err := nilDp.UpdateWorker(workerName, workerInfo)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "not initialized")
	})
}

// TestUpdateCluster test update cluster info in the data pool
func TestUpdateCluster(t *testing.T) {
	dp := newDataPool()
	convey.Convey("UpdateCluster should modify existing cluster", t, func() {
		_ = dp.RegisterCluster(clusterName)
		business := []int32{0, 0}
		updatedCluster := &ClusterInfo{
			Command:   map[string]string{"cmd1": "updated"},
			HeartBeat: time.Now(),
			Business:  business,
			RWMutex:   sync.RWMutex{},
		}

		err := dp.UpdateCluster(clusterName, updatedCluster)
		convey.So(err, convey.ShouldBeNil)

		cluster, _ := dp.GetCluster(clusterName)
		convey.So(cluster.Command["cmd1"], convey.ShouldEqual, "updated")
		convey.So(cluster.Business, convey.ShouldResemble, business)
	})
}

// TestGetAgent test get agent info from data pool
func TestGetAgent(t *testing.T) {
	dp := newDataPool()
	convey.Convey("GetAgent should return agent info", t, func() {
		_ = dp.RegisterAgent(agentName, agentInfo)
		agent, err := dp.GetAgent(agentName)
		convey.So(err, convey.ShouldBeNil)
		convey.So(agent, convey.ShouldNotBeNil)
		convey.So(agent.NodeRank, convey.ShouldEqual, "1")
	})

	convey.Convey("GetAgent should fail for non-existent agent", t, func() {
		_, err := dp.GetAgent("nonexistent")
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "nonexistent")
	})
}

// TestGetWorker test get worker info from data pool
func TestGetWorker(t *testing.T) {
	dp := newDataPool()
	convey.Convey("GetWorker should return worker info", t, func() {
		_ = dp.RegisterWorker(workerName, workerInfo)
		worker, err := dp.GetWorker(workerName)
		convey.So(err, convey.ShouldBeNil)
		convey.So(worker, convey.ShouldNotBeNil)
		convey.So(worker.GlobalRank, convey.ShouldEqual, "1")
	})

	convey.Convey("GetWorker should fail for non-existent worker", t, func() {
		_, err := dp.GetWorker("nonexistent")
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "nonexistent")
	})
}

// TestGetCluster test get cluster info from data pool
func TestGetCluster(t *testing.T) {
	dp := newDataPool()
	convey.Convey("GetCluster should return cluster info", t, func() {
		_ = dp.RegisterCluster(clusterName)

		cluster, err := dp.GetCluster(clusterName)
		convey.So(err, convey.ShouldBeNil)
		convey.So(cluster, convey.ShouldNotBeNil)
	})

	convey.Convey("GetCluster should fail for non-existent cluster", t, func() {
		_, err := dp.GetCluster("nonexistent")
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "nonexistent")
	})
}

func TestGetPos(t *testing.T) {
	dp := newDataPool()
	_ = dp.RegisterAgent(agentName, &AgentInfo{Pos: testPos})
	_ = dp.RegisterWorker(workerName, &WorkerInfo{Pos: testPos})
	convey.Convey("Test get agent pos success return agent pos", t, func() {
		pos, err := dp.GetPos(common.AgentRole, agentName)
		convey.So(err, convey.ShouldBeNil)
		convey.So(pos, convey.ShouldNotBeNil)
		convey.So(pos.Role, convey.ShouldEqual, testPos.Role)
		convey.So(pos.ServerRank, convey.ShouldEqual, testPos.ServerRank)
	})
	convey.Convey("Test get worker pos success return worker pos", t, func() {
		pos, err := dp.GetPos(common.WorkerRole, workerName)
		convey.So(err, convey.ShouldBeNil)
		convey.So(pos, convey.ShouldNotBeNil)
		convey.So(pos.Role, convey.ShouldEqual, testPos.Role)
	})
	convey.Convey("Test get agent pos, agent is unregistered return error", t, func() {
		_, err := dp.GetPos(common.AgentRole, "unregistered_agent")
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "unregistered_agent")
	})
	convey.Convey("Test get worker pos, worker is unregistered return error", t, func() {
		_, err := dp.GetPos(common.WorkerRole, "unregistered_worker")
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "unregistered_worker")
	})
	convey.Convey("Test get pos, type is invalid return error", t, func() {
		_, err := dp.GetPos("invalid_type", agentName)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "invalid info type")
	})
	convey.Convey("Test get agent pos, pos is nil return error", t, func() {
		_ = dp.RegisterAgent("no_pos_agent", &AgentInfo{Pos: nil})
		_, err := dp.GetPos(common.AgentRole, "no_pos_agent")
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "no_pos_agent")
	})
	convey.Convey("Test get worker pos, pos is nil return error", t, func() {
		_ = dp.RegisterWorker("no_pos_worker", &WorkerInfo{Pos: nil})
		_, err := dp.GetPos(common.WorkerRole, "no_pos_worker")
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "no_pos_worker")
	})
}

// TestGetSnapShot test get data pool snapshot
func TestGetSnapShot(t *testing.T) {
	convey.Convey("TestGetSnapShot test get data pool snapshot success", t, func() {
		snapshot := &SnapShot{
			AgentInfos: &AgentInfos{
				Agents:    map[string]*AgentInfo{"nilAgent": nil, agentName: agentInfo},
				AllStatus: map[string]string{"status": "value1"},
			},
			WorkerInfos: &WorkerInfos{
				Workers:   map[string]*WorkerInfo{"nilWorker": nil, workerName: workerInfo},
				AllStatus: map[string]string{"status": "value1"},
			},
			ClusterInfos: &ClusterInfos{
				Clusters:  map[string]*ClusterInfo{"nilCluster": nil, clusterName: clusterInfo},
				AllStatus: map[string]string{"status": "value1"},
			},
		}
		dp := &DataPool{Snapshot: snapshot}
		getSnapshot, err := dp.GetSnapShot()
		convey.So(err, convey.ShouldBeNil)
		convey.So(getSnapshot, convey.ShouldEqual, snapshot)
	})
	convey.Convey("TestGetSnapShot test get data pool snapshot fail", t, func() {
		dp := &DataPool{Snapshot: nil}
		getSnapshot, err := dp.GetSnapShot()
		convey.So(getSnapshot, convey.ShouldBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "null")
	})
}
