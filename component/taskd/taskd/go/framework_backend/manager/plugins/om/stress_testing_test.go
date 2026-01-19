// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package om a series of service function
package om

import (
	"strings"
	"testing"
	"time"

	"clusterd/pkg/interface/grpc/recover"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure"
	"taskd/framework_backend/manager/infrastructure/storage"
	"taskd/toolkit_backend/net/common"
)

func TestStressTestPluginName(t *testing.T) {
	t.Run("get plugin name after get init plugin", func(t *testing.T) {
		plugin := NewOmStressTestPlugin()
		assert.NotNil(t, plugin)
		assert.Equal(t, constant.OMStressTestPluginName, plugin.Name())
	})
}

func TestStressTestRelease(t *testing.T) {
	t.Run("test release plugin", func(t *testing.T) {
		plugin := NewOmStressTestPlugin()
		err := plugin.Release()
		assert.Nil(t, err)
	})
}

func TestResetPluginStatus(t *testing.T) {
	t.Run("TestResetPluginStatus, ok", func(t *testing.T) {
		plugin := &StressTestPlugin{
			workerStatus: map[string]*pb.StressTestRankResult{
				"worker": {},
			},
		}
		plugin.resetPluginStatus()
		assert.Equal(t, 0, len(plugin.workerStatus))
	})
}

func TestGetWorkerName(t *testing.T) {
	t.Run("Get worker name ok, after get init plugin", func(t *testing.T) {
		plugin := &StressTestPlugin{
			pullMsg:      make([]infrastructure.Msg, 0),
			uuid:         "",
			jobID:        "",
			workerStatus: make(map[string]*pb.StressTestRankResult),
			heartbeat:    make(map[string]heartbeatInfo),
		}
		plugin.workerStatus = map[string]*pb.StressTestRankResult{
			"worker1": {},
		}
		names := plugin.getWorkerName()
		assert.Equal(t, 1, len(names))
	})
}

func TestUpdateStressTestWorkerStatusOK(t *testing.T) {
	t.Run("update WorkerStatus result, after recv shot, ok ", func(t *testing.T) {
		plugin := &StressTestPlugin{
			pullMsg:      make([]infrastructure.Msg, 0),
			uuid:         "",
			jobID:        "",
			workerStatus: make(map[string]*pb.StressTestRankResult),
			heartbeat:    make(map[string]heartbeatInfo),
		}
		workerName := "worker1"
		uuid := "uuid"
		res := "ok"
		plugin.uuid = uuid
		result := &pb.StressTestRankResult{
			RankResult: map[string]*pb.StressTestOpResult{
				"1": {Code: constant.StressTestOK, Result: res},
			},
		}
		plugin.workerStatus = map[string]*pb.StressTestRankResult{
			strings.TrimPrefix(workerName, common.WorkerRole): {},
		}
		patches := gomonkey.NewPatches()
		patches.ApplyPrivateMethod(plugin, "handleWorkerHeartbeat", func(name string, info *storage.WorkerInfo) bool {
			return true
		})
		defer patches.Reset()
		shot := storage.SnapShot{
			WorkerInfos: &storage.WorkerInfos{
				Workers: map[string]*storage.WorkerInfo{
					workerName: {
						Status: map[string]string{
							constant.StressTestUUID: uuid,
							constant.StressTest:     utils.ObjToString(result),
						},
					},
				},
			},
		}
		plugin.updateWorkerStatus(shot)
		assert.Equal(t, 1, len(plugin.workerStatus[strings.TrimPrefix(workerName, common.WorkerRole)].RankResult))
	})
}

func TestUpdateStressTestWorkerStatusFail(t *testing.T) {
	t.Run("update WorkerStatus result, after recv shot, fail ", func(t *testing.T) {
		plugin := &StressTestPlugin{
			pullMsg:      make([]infrastructure.Msg, 0),
			uuid:         "",
			jobID:        "",
			workerStatus: make(map[string]*pb.StressTestRankResult),
			heartbeat:    make(map[string]heartbeatInfo),
		}
		workerName := "worker1"
		uuid := "uuid"
		res := "ok"
		plugin.uuid = uuid
		plugin.workerStatus = map[string]*pb.StressTestRankResult{
			strings.TrimPrefix(workerName, common.WorkerRole): {},
		}
		patches := gomonkey.NewPatches()
		patches.ApplyPrivateMethod(plugin, "handleWorkerHeartbeat", func(name string, info *storage.WorkerInfo) bool {
			return true
		})
		defer patches.Reset()
		shot := storage.SnapShot{
			WorkerInfos: &storage.WorkerInfos{
				Workers: map[string]*storage.WorkerInfo{
					workerName: {
						Status: map[string]string{
							constant.StressTestUUID: uuid,
							constant.StressTest:     res,
						},
					},
				},
			},
		}
		plugin.updateWorkerStatus(shot)
		assert.Equal(t, 0, len(plugin.workerStatus[strings.TrimPrefix(workerName, common.WorkerRole)].RankResult))
	})
}

func TestInitStressTestPluginStatus(t *testing.T) {
	t.Run("initStressTestPluginStatus, parse data fail", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		called := false
		patches.ApplyPrivateMethod(&StressTestPlugin{}, "replyToClusterDMsg", func(retryTime time.Duration, result bool) {
			called = true
			return
		})
		plugin := &StressTestPlugin{
			workerStatus: make(map[string]*pb.StressTestRankResult),
			heartbeat:    make(map[string]heartbeatInfo),
		}
		shot := storage.SnapShot{
			ClusterInfos: &storage.ClusterInfos{
				Clusters: map[string]*storage.ClusterInfo{
					constant.ClusterDRank: {
						Command: map[string]string{constant.StressTestRankOPStr: "ok"}}}},
		}
		plugin.initPluginStatus(shot)
		assert.True(t, called)
	})
	t.Run("initStressTestPluginStatus ok", func(t *testing.T) {
		plugin := &StressTestPlugin{
			pullMsg: make([]infrastructure.Msg, 0), workerStatus: make(map[string]*pb.StressTestRankResult),
			heartbeat: make(map[string]heartbeatInfo),
		}
		uuid := "uuid"
		res := "ok"
		jobID := "jobID"
		plugin.uuid = uuid
		result := &pb.StressTestRankResult{
			RankResult: map[string]*pb.StressTestOpResult{"1": {Code: constant.StressTestOK, Result: res}},
		}
		shot := storage.SnapShot{
			ClusterInfos: &storage.ClusterInfos{
				Clusters: map[string]*storage.ClusterInfo{
					constant.ClusterDRank: {
						Command: map[string]string{
							constant.StressTestJobID: jobID, constant.StressTestUUID: uuid,
							constant.StressTestRankOPStr: utils.ObjToString(result)}}},
			},
		}
		plugin.initPluginStatus(shot)
		assert.Equal(t, uuid, plugin.uuid)
		assert.Equal(t, jobID, plugin.jobID)
	})
}

func TestHandleWorkerHeartbeatOK(t *testing.T) {
	t.Run("TestHandleWorkerHeartbeat, heartbeat ok", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		plugin := &StressTestPlugin{
			pullMsg:      make([]infrastructure.Msg, 0),
			uuid:         "",
			jobID:        "",
			workerStatus: make(map[string]*pb.StressTestRankResult),
			heartbeat:    make(map[string]heartbeatInfo),
		}
		uuid := "uuid"
		plugin.uuid = uuid
		workerName := "1"
		nowTime := time.Now()
		plugin.heartbeat[workerName] = heartbeatInfo{
			heartbeat: nowTime.Unix() - 1,
		}
		workInfo := &storage.WorkerInfo{
			Status: map[string]string{
				constant.StressTest: "ok",
			},
			HeartBeat: nowTime,
		}
		result := plugin.handleWorkerHeartbeat(workerName, workInfo)
		assert.Equal(t, true, result)
	})
}

func TestHandleWorkerHeartbeatStressTestOK(t *testing.T) {
	t.Run("TestHandleWorkerHeartbeat, StressTest ok", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		plugin := &StressTestPlugin{
			pullMsg:      make([]infrastructure.Msg, 0),
			uuid:         "",
			jobID:        "",
			workerStatus: make(map[string]*pb.StressTestRankResult),
			heartbeat:    make(map[string]heartbeatInfo),
		}
		uuid := "uuid"
		res := "ok"
		plugin.uuid = uuid
		workerName := "1"
		nowTime := time.Now()
		plugin.heartbeat[workerName] = heartbeatInfo{
			heartbeat: nowTime.Unix(),
		}
		workInfo := &storage.WorkerInfo{
			Status: map[string]string{
				constant.StressTest: res,
			},
			HeartBeat: nowTime,
		}
		result := plugin.handleWorkerHeartbeat(workerName, workInfo)
		assert.Equal(t, true, result)
	})
}

func TestHandleWorkerHeartbeatLose(t *testing.T) {
	t.Run("TestHandleWorkerHeartbeat, lose heartbeat", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		plugin := &StressTestPlugin{
			pullMsg:      make([]infrastructure.Msg, 0),
			uuid:         "",
			jobID:        "",
			workerStatus: make(map[string]*pb.StressTestRankResult),
			heartbeat:    make(map[string]heartbeatInfo),
			rankOpMap:    make(map[string]*pb.StressOpList),
		}
		uuid := "uuid"
		plugin.uuid = uuid
		workerName := "1"
		nowTime := time.Now()
		plugin.heartbeat[workerName] = heartbeatInfo{
			heartbeat: nowTime.Unix(),
			dropTime:  maxHeartbeatInterval + 1,
		}
		plugin.rankOpMap[workerName] = &pb.StressOpList{
			Ops: []int64{0},
		}
		workInfo := &storage.WorkerInfo{
			Status: map[string]string{
				constant.StressTest: "",
			},
			HeartBeat: nowTime,
		}
		result := plugin.handleWorkerHeartbeat(workerName, workInfo)
		assert.Equal(t, false, result)
	})
}

func TestPullStressTestMsg(t *testing.T) {
	t.Run("Pull plugin msg success", func(t *testing.T) {
		plugin := &StressTestPlugin{
			pullMsg:      make([]infrastructure.Msg, 0),
			uuid:         "",
			jobID:        "",
			workerStatus: make(map[string]*pb.StressTestRankResult),
		}
		res, err := plugin.PullMsg()
		assert.Nil(t, err)
		assert.Equal(t, 0, len(res))
	})
}

func TestReplyToCluster(t *testing.T) {
	patches := gomonkey.NewPatches()
	ws := make(map[string]*pb.StressTestRankResult)
	plugin := &StressTestPlugin{workerStatus: ws}
	t.Run("reply to clusterD max times", func(t *testing.T) {
		defer func() {
			plugin.pullMsg = make([]infrastructure.Msg, 0)
		}()
		defer patches.Reset()
		plugin.replyToClusterDMsg(ws)
		assert.Equal(t, 1, len(plugin.pullMsg))
	})
}

func TestStressTestHandle(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	patches.ApplyPrivateMethod(&StressTestPlugin{}, "replyToClusterDMsg", func(retryTime time.Duration, result bool) {
		return
	})
	t.Run("Handle show in worker status is empty, skip", func(t *testing.T) {
		plugin := &StressTestPlugin{
			workerStatus: make(map[string]*pb.StressTestRankResult),
		}
		res, err := plugin.Handle()
		assert.Nil(t, err)
		assert.Equal(t, constant.HandleStageFinal, res.Stage)
	})
	t.Run("Handle show in stress test ok senior, update worker status", func(t *testing.T) {
		plugin := &StressTestPlugin{
			workerStatus: make(map[string]*pb.StressTestRankResult),
		}
		plugin.workerStatus = map[string]*pb.StressTestRankResult{
			"worker1": {
				RankResult: map[string]*pb.StressTestOpResult{
					"1": {
						Code:   "0",
						Result: "ok",
					},
				},
			},
		}
		res, err := plugin.Handle()
		assert.Nil(t, err)
		assert.Equal(t, constant.HandleStageFinal, res.Stage)
	})
}

func TestStressTestPredicateGetParamFailed(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	plugin := &StressTestPlugin{
		workerStatus: map[string]*pb.StressTestRankResult{
			"workerName": {},
		},
	}
	t.Run("get param failed, ClusterDRank empty", func(t *testing.T) {
		shot := storage.SnapShot{
			ClusterInfos: &storage.ClusterInfos{
				Clusters: map[string]*storage.ClusterInfo{
					constant.TaskDRank: {},
				},
			},
		}
		res, err := plugin.Predicate(shot)
		assert.Nil(t, err)
		assert.Equal(t, constant.UnselectStatus, res.CandidateStatus)
	})
	t.Run("get param failed, stress test param empty", func(t *testing.T) {
		shot := storage.SnapShot{
			ClusterInfos: &storage.ClusterInfos{
				Clusters: map[string]*storage.ClusterInfo{
					constant.ClusterDRank: {
						Command: map[string]string{
							constant.StressTestJobID: "",
						},
					},
				},
			},
		}
		res, err := plugin.Predicate(shot)
		assert.Nil(t, err)
		assert.Equal(t, constant.UnselectStatus, res.CandidateStatus)
	})
}

func TestPredicateStressTest(t *testing.T) {
	t.Run("Predicate: stress test", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		uuid := "uuid"
		ops := "ops"
		jobid := "jobid"
		patches.ApplyPrivateMethod(&StressTestPlugin{}, "updateWorkerStatus", func(shot storage.SnapShot) {
			return
		})
		plugin := &StressTestPlugin{
			uuid:  uuid,
			jobID: jobid,
			workerStatus: map[string]*pb.StressTestRankResult{
				"workerName": {},
			},
		}
		shot := storage.SnapShot{
			ClusterInfos: &storage.ClusterInfos{
				Clusters: map[string]*storage.ClusterInfo{
					constant.ClusterDRank: {
						Command: map[string]string{
							constant.StressTestJobID:     jobid,
							constant.StressTestRankOPStr: ops,
							constant.StressTestUUID:      uuid,
						},
					},
				},
			},
		}
		res, err := plugin.Predicate(shot)
		assert.Nil(t, err)
		assert.Equal(t, constant.CandidateStatus, res.CandidateStatus)
	})
}

func TestPredicateNewStressTest(t *testing.T) {
	t.Run("Predicate: new stress test", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		uuid := "uuid"
		ops := "ops"
		jobid := "jobid"
		patches.ApplyPrivateMethod(&StressTestPlugin{}, "initPluginStatus", func(shot storage.SnapShot) {
			return
		})
		plugin := &StressTestPlugin{
			workerStatus: map[string]*pb.StressTestRankResult{},
		}
		shot := storage.SnapShot{
			ClusterInfos: &storage.ClusterInfos{
				Clusters: map[string]*storage.ClusterInfo{
					constant.ClusterDRank: {
						Command: map[string]string{
							constant.StressTestJobID:     jobid,
							constant.StressTestRankOPStr: ops,
							constant.StressTestUUID:      uuid,
						},
					},
				},
			},
		}
		res, err := plugin.Predicate(shot)
		assert.Nil(t, err)
		assert.Equal(t, constant.CandidateStatus, res.CandidateStatus)
	})
}
