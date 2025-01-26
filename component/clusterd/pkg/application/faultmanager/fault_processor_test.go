package faultmanager

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	v1 "k8s.io/api/core/v1"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/interface/kube"
)

const (
	JobId    = "Job"
	NodeName = "Node"
)

func getDemoJobServerMap() constant.JobServerInfoMap {
	return constant.JobServerInfoMap{
		InfoMap: map[string]map[string]constant.ServerHccl{
			JobId: {
				NodeName: constant.ServerHccl{
					DeviceList: []constant.Device{{
						DeviceID: "0",
						RankID:   "0",
					}, {
						DeviceID: "1",
						RankID:   "1",
					}},
				},
			},
		},
	}
}

func TestFaultProcessorImplProcess(t *testing.T) {
	t.Run("test node fail, job fault rank list should correct", func(t *testing.T) {
		processor := GlobalFaultProcessCenter.faultJobProcessor
		jobServerMap := getDemoJobServerMap()
		GlobalFaultProcessCenter.jobServerInfoMap = jobServerMap
		mockKube := gomonkey.ApplyFunc(kube.GetNode, func(name string) *v1.Node {
			return nil
		})
		defer mockKube.Reset()
		processor.Process()
		rankProcessor, _ := GlobalFaultProcessCenter.DeviceCenter.getJobFaultRankProcessor()
		faultRankInfos := rankProcessor.getJobFaultRankInfos()
		if len(faultRankInfos[JobId].FaultList) != len(jobServerMap.InfoMap[JobId][NodeName].DeviceList) {
			t.Error("TestFaultProcessorImplProcess fail")
		}
	})
}
