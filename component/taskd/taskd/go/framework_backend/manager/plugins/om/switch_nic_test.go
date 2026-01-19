// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package om a series of service function
package om

import (
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"taskd/common/constant"
	"taskd/framework_backend/manager/infrastructure"
	"taskd/framework_backend/manager/infrastructure/storage"
)

func TestOmPluginName(t *testing.T) {
	t.Run("get plugin name after get init plugin", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		plugin := NewOmSwitchNicPlugin()
		assert.NotNil(t, plugin)
		assert.Equal(t, constant.OMSwitchNicPluginName, plugin.Name())
	})
}

func TestGetAllWorkerName(t *testing.T) {
	t.Run("Get worker name ok, after get init plugin", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		plugin := &SwitchNicPlugin{
			pullMsg:      make([]infrastructure.Msg, 0),
			uuid:         "",
			jobID:        "",
			workerStatus: make(map[string]string),
		}
		plugin.workerStatus = map[string]string{
			"worker1": "",
		}
		names := plugin.getAllWorkerName()
		assert.Equal(t, 1, len(names))
	})
}

func TestUpdateWorkerStatus(t *testing.T) {
	t.Run("update WorkerStatus switch result correct, after recv shot ", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		plugin := &SwitchNicPlugin{
			pullMsg:      make([]infrastructure.Msg, 0),
			uuid:         "",
			jobID:        "",
			workerStatus: make(map[string]string),
		}

		workerName := "worker1"
		uuid := "uuid"
		res := "ok"
		plugin.uuid = uuid
		plugin.workerStatus = map[string]string{
			workerName: "",
		}
		shot := storage.SnapShot{
			WorkerInfos: &storage.WorkerInfos{
				Workers: map[string]*storage.WorkerInfo{
					workerName: {
						Status: map[string]string{
							constant.SwitchNicUUID: uuid,
							constant.SwitchNic:     res,
						},
					},
				},
			},
		}
		plugin.updateWorkerStatus(shot)
		assert.Equal(t, res, plugin.workerStatus[workerName])
	})
}

func TestInitPluginStatus(t *testing.T) {
	t.Run("initPluginStatus ok", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		plugin := &SwitchNicPlugin{
			pullMsg:      make([]infrastructure.Msg, 0),
			uuid:         "",
			jobID:        "",
			workerStatus: make(map[string]string),
		}

		workerName := "worker1"
		uuid := "uuid"
		res := "ok"
		jobID := "jobID"
		plugin.uuid = uuid
		plugin.workerStatus = map[string]string{
			workerName: "",
		}
		shot := storage.SnapShot{
			WorkerInfos: &storage.WorkerInfos{
				Workers: map[string]*storage.WorkerInfo{
					workerName: {
						Status: map[string]string{
							constant.SwitchNicUUID: uuid,
							constant.SwitchNic:     res,
						},
					},
				},
			},
			ClusterInfos: &storage.ClusterInfos{
				Clusters: map[string]*storage.ClusterInfo{
					constant.ClusterDRank: {
						Command: map[string]string{
							constant.SwitchJobID:   jobID,
							constant.SwitchNicUUID: uuid,
						},
					},
				},
			},
		}
		plugin.initPluginStatus(shot)
		assert.Equal(t, uuid, plugin.uuid)
		assert.Equal(t, jobID, plugin.jobID)
	})
}

func TestPullMsg(t *testing.T) {
	t.Run("Pull plugin msg success", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		plugin := &SwitchNicPlugin{
			pullMsg:      make([]infrastructure.Msg, 0),
			uuid:         "",
			jobID:        "",
			workerStatus: make(map[string]string),
		}

		res, err := plugin.PullMsg()
		assert.Nil(t, err)
		assert.Equal(t, 0, len(res))
	})
}

func TestReplyToClusterDMsg(t *testing.T) {
	t.Run("get addr failed, GetClusterdAddr will be execute", func(t *testing.T) {
		plugin := &SwitchNicPlugin{
			pullMsg:      make([]infrastructure.Msg, 0),
			uuid:         "",
			jobID:        "",
			workerStatus: make(map[string]string),
		}
		defer func() {
			plugin.pullMsg = make([]infrastructure.Msg, 0)
		}()
		plugin.replyToClusterDMsg(true)
		assert.Equal(t, 1, len(plugin.pullMsg))
	})
}

func TestHandleWorkerStatusEmpty(t *testing.T) {
	t.Run("Handle show in worker status is empty, skip", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyPrivateMethod(&SwitchNicPlugin{}, "replyToClusterDMsg", func(retryTime time.Duration, result bool) {
			return
		})
		plugin := &SwitchNicPlugin{
			pullMsg:      make([]infrastructure.Msg, 0),
			uuid:         "",
			jobID:        "",
			workerStatus: make(map[string]string),
		}
		res, err := plugin.Handle()
		assert.Nil(t, err)
		assert.Equal(t, constant.HandleStageFinal, res.Stage)
	})
}

func TestHandleOK(t *testing.T) {
	t.Run("Handle show in switch ok senior, update worker status", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyPrivateMethod(&SwitchNicPlugin{}, "replyToClusterDMsg", func(retryTime time.Duration, result bool) {
			return
		})
		plugin := &SwitchNicPlugin{
			pullMsg:      make([]infrastructure.Msg, 0),
			uuid:         "",
			jobID:        "",
			workerStatus: make(map[string]string),
		}
		plugin.workerStatus = map[string]string{
			"worker1": constant.SwitchOK,
		}

		res, err := plugin.Handle()
		assert.Nil(t, err)
		assert.Equal(t, constant.HandleStageFinal, res.Stage)
	})
}

func TestHandleFAIL(t *testing.T) {
	t.Run("Handle switch result in switch failed senior", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyPrivateMethod(&SwitchNicPlugin{}, "replyToClusterDMsg", func(retryTime time.Duration, result bool) {
			return
		})
		plugin := &SwitchNicPlugin{
			pullMsg:      make([]infrastructure.Msg, 0),
			uuid:         "",
			jobID:        "",
			workerStatus: make(map[string]string),
		}
		plugin.workerStatus = map[string]string{
			"worker1": constant.SwitchFail,
		}

		res, err := plugin.Handle()
		assert.Nil(t, err)
		assert.Equal(t, constant.HandleStageFinal, res.Stage)
	})
}

func TestPredicateGetParamFailed(t *testing.T) {
	t.Run("get param failed", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		plugin := &SwitchNicPlugin{
			pullMsg: make([]infrastructure.Msg, 0),
			uuid:    "",
			jobID:   "",
			workerStatus: map[string]string{
				"workerName": "",
			},
		}
		plugin.workerStatus = map[string]string{}
		shot := storage.SnapShot{
			ClusterInfos: &storage.ClusterInfos{
				Clusters: map[string]*storage.ClusterInfo{
					constant.ClusterDRank: {
						Command: map[string]string{
							constant.SwitchJobID:   "",
							constant.GlobalRankKey: "",
							constant.GlobalOpKey:   "",
							constant.SwitchNicUUID: "",
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

func TestPredicateSwitchingNic(t *testing.T) {
	t.Run("Predicate: switching nic", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		uuid := "uuid"
		ranks := "ranks"
		ops := "ops"
		jobid := "jobid"
		patches.ApplyPrivateMethod(&SwitchNicPlugin{}, "updateWorkerStatus", func(shot storage.SnapShot) {
			return
		})
		plugin := &SwitchNicPlugin{
			pullMsg: make([]infrastructure.Msg, 0),
			uuid:    uuid,
			jobID:   "",
			workerStatus: map[string]string{
				"workerName": "",
			},
		}
		shot := storage.SnapShot{
			ClusterInfos: &storage.ClusterInfos{
				Clusters: map[string]*storage.ClusterInfo{
					constant.ClusterDRank: {
						Command: map[string]string{
							constant.SwitchJobID:   jobid,
							constant.GlobalRankKey: ranks,
							constant.GlobalOpKey:   ops,
							constant.SwitchNicUUID: uuid,
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

func TestPredicateNewSwitchNic(t *testing.T) {
	t.Run("Predicate: new switch nic", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		uuid := "uuid"
		ranks := "ranks"
		ops := "ops"
		jobid := "jobid"
		patches.ApplyPrivateMethod(&SwitchNicPlugin{}, "initPluginStatus", func(shot storage.SnapShot) {
			return
		})
		plugin := &SwitchNicPlugin{
			pullMsg:      make([]infrastructure.Msg, 0),
			uuid:         "",
			jobID:        "",
			workerStatus: map[string]string{},
		}
		shot := storage.SnapShot{
			ClusterInfos: &storage.ClusterInfos{
				Clusters: map[string]*storage.ClusterInfo{
					constant.ClusterDRank: {
						Command: map[string]string{
							constant.SwitchJobID:   jobid,
							constant.GlobalRankKey: ranks,
							constant.GlobalOpKey:   ops,
							constant.SwitchNicUUID: uuid,
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

func TestPredicateOtherScenarios(t *testing.T) {
	t.Run("Predicate: other scenarios", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		uuid := "uuid"

		patches.ApplyPrivateMethod(&SwitchNicPlugin{}, "updateWorkerStatus", func(shot storage.SnapShot) {
			return
		})
		plugin := &SwitchNicPlugin{
			pullMsg:      make([]infrastructure.Msg, 0),
			uuid:         uuid,
			jobID:        "",
			workerStatus: map[string]string{},
		}
		plugin.workerStatus = map[string]string{}
		shot := storage.SnapShot{
			ClusterInfos: &storage.ClusterInfos{
				Clusters: map[string]*storage.ClusterInfo{
					constant.ClusterDRank: {
						Command: map[string]string{
							constant.SwitchJobID:   "",
							constant.GlobalRankKey: "",
							constant.GlobalOpKey:   "",
							constant.SwitchNicUUID: uuid,
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

func TestSwitchNecRelease(t *testing.T) {
	t.Run("test release plugin", func(t *testing.T) {
		plugin := NewOmSwitchNicPlugin()
		err := plugin.Release()
		assert.Nil(t, err)
	})
}
