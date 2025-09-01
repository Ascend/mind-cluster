// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package recover a series of service function
package recover

import (
	"context"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"clusterd/pkg/application/faultmanager"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/interface/grpc/recover"
)

func TestSwitchNicErrorParam(t *testing.T) {
	patches := gomonkey.NewPatches()
	ctx := context.Background()
	jobID := "jobID"
	nodeName := "nodeName"
	deviceID := "device"
	rankID := "1"
	job.SaveJobCache(jobID, constant.JobInfo{
		PreServerList: []constant.ServerHccl{
			{ServerName: nodeName, DeviceList: []constant.Device{{DeviceID: deviceID, RankID: rankID}}},
		},
		Status: job.StatusJobRunning,
	})
	defer job.DeleteJobCache(jobID)
	defer patches.Reset()
	t.Run("switch nic, error param", func(t *testing.T) {
		s := fakeService()
		res, _ := s.SwitchNicTrack(ctx, nil)
		assert.Equal(t, int32(common.OMParamInvalid), res.Code)
	})
}

func TestSwitchNicCanNotDoSwitch(t *testing.T) {
	patches := gomonkey.NewPatches()
	ctx := context.Background()
	jobID := "jobID"
	nodeName := "nodeName"
	deviceID := "device"
	rankID := "1"
	job.SaveJobCache(jobID, constant.JobInfo{
		PreServerList: []constant.ServerHccl{
			{ServerName: nodeName, DeviceList: []constant.Device{{DeviceID: deviceID, RankID: rankID}}},
		},
		Status: job.StatusJobRunning,
	})
	defer job.DeleteJobCache(jobID)
	defer patches.Reset()
	t.Run("can not do switch nic", func(t *testing.T) {
		s := fakeService()
		patches.ApplyPrivateMethod(s, "checkNicsParam", func(_ *pb.SwitchNics) (bool, string) {
			return true, ""
		})
		patches.ApplyPrivateMethod(&EventController{}, "canDoSwitchingNic", func(*FaultRecoverService) bool {
			return false
		})
		s.eventCtl[jobID] = &EventController{state: common.NewStateMachine(common.InitState, nil)}
		res, _ := s.SwitchNicTrack(ctx, &pb.SwitchNics{
			JobID: jobID,
			NicOps: map[string]*pb.DeviceList{
				nodeName: {Dev: []string{deviceID}, Op: []bool{true}}},
		})
		assert.Equal(t, int32(common.OMIsRunning), res.Code)
	})
}

func TestSwitchNicOperationSuccess(t *testing.T) {
	patches := gomonkey.NewPatches()
	ctx := context.Background()
	jobID := "jobID"
	nodeName := "nodeName"
	deviceID := "device"
	rankID := "1"
	job.SaveJobCache(jobID, constant.JobInfo{
		PreServerList: []constant.ServerHccl{
			{ServerName: nodeName, DeviceList: []constant.Device{{DeviceID: deviceID, RankID: rankID}}},
		},
		Status: job.StatusJobRunning,
	})
	defer job.DeleteJobCache(jobID)
	defer patches.Reset()
	t.Run("switch nic operation success", func(t *testing.T) {
		s := fakeService()
		s.eventCtl[jobID] = &EventController{state: common.NewStateMachine(common.InitState, nil)}
		patches.ApplyPrivateMethod(s, "checkNicsParam", func(_ *pb.SwitchNics) (bool, string) {
			return true, ""
		})
		res, _ := s.SwitchNicTrack(ctx, &pb.SwitchNics{
			JobID:  jobID,
			NicOps: map[string]*pb.DeviceList{nodeName: {Dev: []string{deviceID}, Op: []bool{true}}},
		})
		assert.Equal(t, int32(common.OK), res.Code)
	})
}

func TestSubscribeSwitchNicSignal(t *testing.T) {
	info := &pb.SwitchNicRequest{
		JobID: "jobID",
	}
	t.Run("case job not registered", func(t *testing.T) {
		s := fakeService()
		err := s.SubscribeSwitchNicSignal(info, nil)
		assert.Error(t, err)
	})
	t.Run("case job registered", func(t *testing.T) {
		s := fakeService()
		patch := gomonkey.ApplyPrivateMethod(&EventController{}, "listenSwitchNicChannel",
			func(stream pb.Recover_SubscribeSwitchNicSignalServer) {
				return
			})
		defer patch.Reset()
		s.eventCtl[info.JobID] = &EventController{}
		s.eventCtl[info.JobID].globalSwitchRankIDs = []string{"1"}
		err := s.SubscribeSwitchNicSignal(info, nil)
		assert.Nil(t, err)
	})
}

func TestSubscribeNotifySwitch(t *testing.T) {
	info := &pb.ClientInfo{
		JobId: "jobID",
	}
	t.Run("case job not registered", func(t *testing.T) {
		s := fakeService()
		err := s.SubscribeNotifySwitch(info, nil)
		assert.Error(t, err)
	})
	t.Run("case job registered", func(t *testing.T) {
		s := fakeService()
		patch := gomonkey.ApplyPrivateMethod(&EventController{}, "listenSwitchNicNotifyChannel",
			func(pb.Recover_SubscribeNotifySwitchServer) {
				return
			})
		defer patch.Reset()
		s.eventCtl[info.JobId] = &EventController{}
		err := s.SubscribeNotifySwitch(info, nil)
		assert.Nil(t, err)
	})
}

func TestCheckNicsParam(t *testing.T) {
	jobID := "jobID"
	nodeName := "nodeName"
	deviceID := "device"
	rankID := "1"
	job.SaveJobCache(jobID, constant.JobInfo{
		JobRankTable: constant.RankTable{
			ServerList: []constant.ServerHccl{
				{ServerName: nodeName, DeviceList: []constant.Device{{DeviceID: deviceID, RankID: rankID}}}},
		},
		Status: job.StatusJobRunning,
	})
	defer job.DeleteJobCache(jobID)
	t.Run("nics is nil", func(t *testing.T) {
		s := fakeService()
		ok, _ := s.checkNicsParam(nil)
		assert.False(t, ok)
	})
	t.Run("job is not exist", func(t *testing.T) {
		s := fakeService()
		ok, _ := s.checkNicsParam(&pb.SwitchNics{JobID: jobID + "1"})
		assert.False(t, ok)
	})
	t.Run("job is not registered", func(t *testing.T) {
		s := fakeService()
		ok, _ := s.checkNicsParam(&pb.SwitchNics{JobID: jobID})
		assert.False(t, ok)
	})
	t.Run("job is not running", func(t *testing.T) {
		s := fakeService()
		s.eventCtl[jobID] = &EventController{}
		jobInfo, _ := job.GetJobCache(jobID)
		jobInfo.Status = job.StatusJobPending
		job.SaveJobCache(jobID, jobInfo)
		ok, _ := s.checkNicsParam(&pb.SwitchNics{JobID: jobID})
		assert.False(t, ok)
		defer func() {
			jobInfo.Status = job.StatusJobRunning
			job.SaveJobCache(jobID, jobInfo)
		}()
	})
}

func TestCheckNicsParamOK(t *testing.T) {
	jobID := "jobID"
	nodeName := "nodeName"
	deviceID := "device"
	rankID := "1"
	patches := gomonkey.ApplyFunc(job.GetJobCache, func(jobKey string) (constant.JobInfo, bool) {
		return constant.JobInfo{
			Status: job.StatusJobRunning,
			JobRankTable: constant.RankTable{
				ServerList: []constant.ServerHccl{
					{
						ServerName: nodeName,
						DeviceList: []constant.Device{
							{
								DeviceID: deviceID,
								RankID:   rankID,
							},
						},
					},
				},
			},
		}, true
	})
	defer patches.Reset()
	t.Run("check param ok", func(t *testing.T) {
		patch := gomonkey.ApplyFunc(faultmanager.QueryDeviceInfoToReport,
			func() map[string]*constant.AdvanceDeviceFaultCm {
				res := make(map[string]*constant.AdvanceDeviceFaultCm)
				res[nodeName] = &constant.AdvanceDeviceFaultCm{
					SuperPodID: 1,
				}
				return res
			})
		s := fakeService()
		s.eventCtl[jobID] = &EventController{}
		ok, _ := s.checkNicsParam(&pb.SwitchNics{
			JobID: jobID,
			NicOps: map[string]*pb.DeviceList{
				nodeName: {Dev: []string{deviceID}, Op: []bool{true}},
			},
		})
		assert.True(t, ok)
		patch.Reset()
	})
}

func TestCheckDevsValid(t *testing.T) {
	DeviceID := "device"
	t.Run("dev and op length is not equal", func(t *testing.T) {
		s := fakeService()
		switchDev := &pb.DeviceList{
			Dev: []string{DeviceID},
			Op:  []bool{},
		}
		_, ok := s.checkDevsValid(switchDev, []string{}, "")
		assert.False(t, ok)
	})
	t.Run("dev or op is empty", func(t *testing.T) {
		s := fakeService()
		switchDev := &pb.DeviceList{
			Dev: []string{},
			Op:  []bool{},
		}
		_, ok := s.checkDevsValid(switchDev, []string{}, "")
		assert.False(t, ok)
	})
	t.Run("device is not exist in node:", func(t *testing.T) {
		s := fakeService()
		switchDev := &pb.DeviceList{
			Dev: []string{DeviceID},
			Op:  []bool{true},
		}
		_, ok := s.checkDevsValid(switchDev, []string{DeviceID + "1"}, "")
		assert.False(t, ok)
	})
	t.Run("check ok", func(t *testing.T) {
		s := fakeService()
		switchDev := &pb.DeviceList{
			Dev: []string{DeviceID},
			Op:  []bool{true},
		}
		_, ok := s.checkDevsValid(switchDev, []string{DeviceID}, "")
		assert.True(t, ok)
	})
}

func TestGetNodeDeviceMap(t *testing.T) {
	t.Run("get expect serverMap ", func(t *testing.T) {
		s := fakeService()
		serverMap := s.getNodeDeviceMap([]constant.ServerHccl{
			{ServerName: "node1", DeviceList: []constant.Device{{DeviceID: "device1"}}},
		})
		assert.Equal(t, "device1", serverMap["node1"][0])
	})
}

func TestGetGlobalRankIDAndOp(t *testing.T) {
	t.Run("get expect param ", func(t *testing.T) {
		jobID := "jobID"
		nodeName := "nodeName"
		deviceID := "device"
		rankID := "1"
		patches := gomonkey.ApplyFunc(job.GetJobCache, func(jobKey string) (constant.JobInfo, bool) {
			return constant.JobInfo{
				Status: job.StatusJobRunning,
				JobRankTable: constant.RankTable{
					ServerList: []constant.ServerHccl{
						{
							ServerName: nodeName,
							DeviceList: []constant.Device{
								{
									DeviceID: deviceID,
									RankID:   rankID,
								},
							},
						},
					},
				},
			}, true
		})
		defer patches.Reset()
		s := fakeService()
		globalRankIDs, globalOps := s.getGlobalRankIDAndOp(&pb.SwitchNics{
			JobID: jobID,
			NicOps: map[string]*pb.DeviceList{
				nodeName: {Dev: []string{deviceID}, Op: []bool{true}},
			},
		})
		assert.Equal(t, rankID, globalRankIDs[0])
		assert.Equal(t, true, globalOps[0])
	})
}

func TestReplySwitchNicResult(t *testing.T) {
	patches := gomonkey.NewPatches()
	ctx := context.Background()
	jobID := "jobID"

	defer patches.Reset()
	t.Run("req is nil", func(t *testing.T) {
		s := fakeService()
		res, _ := s.ReplySwitchNicResult(ctx, nil)
		assert.Equal(t, int32(common.OMParamInvalid), res.Code)
	})
	t.Run("not register", func(t *testing.T) {
		s := fakeService()
		res, _ := s.ReplySwitchNicResult(ctx, &pb.SwitchResult{JobId: ""})
		assert.Equal(t, int32(common.UnRegistry), res.Code)
	})
	t.Run("reply success", func(t *testing.T) {
		s := fakeService()
		s.eventCtl[jobID] = &EventController{}
		patches.ApplyPrivateMethod(&EventController{}, "setSwitchNicResult", func(result *pb.SwitchResult) {
			return
		})
		res, _ := s.ReplySwitchNicResult(ctx, &pb.SwitchResult{JobId: jobID})
		assert.Equal(t, int32(common.OK), res.Code)
	})
}
