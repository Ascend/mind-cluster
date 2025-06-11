// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package om a series of service function
package om

import (
	"fmt"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure"
	"taskd/framework_backend/manager/infrastructure/storage"
)

func TestOmPluginName(t *testing.T) {
	t.Run("get plugin name after get init plugin", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		plugin := NewOmPlugin()
		assert.NotNil(t, plugin)
		assert.Equal(t, constant.OMPluginName, plugin.Name())
	})
}

func TestGetAllWorkerName(t *testing.T) {
	t.Run("Get worker name ok, after get init plugin", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		plugin := &OmPlugin{
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
		plugin := &OmPlugin{
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
		plugin := &OmPlugin{
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
		plugin := &OmPlugin{
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

func TestReplyToClusterDGetaddrFailed(t *testing.T) {
	t.Run("get addr failed, GetClusterdAddr will be execute", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		plugin := &OmPlugin{
			pullMsg:      make([]infrastructure.Msg, 0),
			uuid:         "",
			jobID:        "",
			workerStatus: make(map[string]string),
		}
		called := false
		patches.ApplyFunc(utils.GetClusterdAddr, func() (string, error) {
			called = true
			return "127.0.0.1", fmt.Errorf("get addr failed")
		})
		plugin.replyToClusterD(time.Duration(0), true)
		assert.True(t, called)
	})
}

func TestHandleOK(t *testing.T) {
	t.Run("Handle show in switch ok senior, update worker status", func(t *testing.T) {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		patches.ApplyPrivateMethod(&OmPlugin{}, "replyToClusterD", func(retryTime time.Duration, result bool) {
			return
		})
		plugin := &OmPlugin{
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
		patches.ApplyPrivateMethod(&OmPlugin{}, "replyToClusterD", func(retryTime time.Duration, result bool) {
			return
		})
		plugin := &OmPlugin{
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
		plugin := &OmPlugin{
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
		patches.ApplyPrivateMethod(&OmPlugin{}, "updateWorkerStatus", func(shot storage.SnapShot) {
			return
		})
		plugin := &OmPlugin{
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
		patches.ApplyPrivateMethod(&OmPlugin{}, "initPluginStatus", func(shot storage.SnapShot) {
			return
		})
		plugin := &OmPlugin{
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

		patches.ApplyPrivateMethod(&OmPlugin{}, "updateWorkerStatus", func(shot storage.SnapShot) {
			return
		})
		plugin := &OmPlugin{
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
