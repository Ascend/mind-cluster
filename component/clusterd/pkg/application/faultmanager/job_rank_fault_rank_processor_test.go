// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package faultmanager contain fault process
package faultmanager

import (
	"sort"
	"strings"
	"testing"

	"clusterd/pkg/common/util"
)

func isContainsAny(str string, subStrs ...string) bool {
	for _, subStr := range subStrs {
		if strings.Contains(str, subStr) {
			return true
		}
	}
	return false
}

func TestJobRankFaultInfoProcessor_GetJobFaultRankInfos(t *testing.T) {
	deviceFaultProcessCenter := newDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.getJobFaultRankProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}

	t.Run("TestJobRankFaultInfoProcessor_getJobFaultRankInfos", func(t *testing.T) {
		cmDeviceInfos, jobServerInfoMap, expectFaultRanks, err := readObjectFromJobFaultRankTestYaml()
		if err != nil {
			t.Errorf("%v", err)
		}
		deviceFaultProcessCenter.setProcessingCm(cmDeviceInfos)
		processor.deviceCenter.jobServerInfoMap = jobServerInfoMap
		processor.process()
		if !isFaultRankMapEqual(processor.getJobFaultRankInfos(), expectFaultRanks) {
			t.Errorf("processor.jobFaultInfos = %s, expectFaultRanks = %s",
				util.ObjToString(processor.jobFaultInfoMap), util.ObjToString(expectFaultRanks))
		}
	})

	t.Run("TestJobRankFaultInfoProcessor_getJobFaultRankInfosFilterLevel", func(t *testing.T) {
		cmDeviceInfos, jobServerInfoMap, expectFaultRanks, err := readObjectFromJobFaultRankTestYaml()
		if err != nil {
			t.Errorf("%v", err)
		}
		deviceFaultProcessCenter.setProcessingCm(cmDeviceInfos)
		processor.deviceCenter.jobServerInfoMap = jobServerInfoMap
		processor.process()
		jobFaultRankInfos := processor.getJobFaultRankInfos()
		for _, faultRankInfo := range jobFaultRankInfos {
			sort.Slice(faultRankInfo.FaultList, func(i, j int) bool {
				if faultRankInfo.FaultList[i].RankId < faultRankInfo.FaultList[j].RankId {
					return true
				}
				if faultRankInfo.FaultList[i].RankId > faultRankInfo.FaultList[j].RankId {
					return false
				}
				return faultRankInfo.FaultList[i].FaultCode < faultRankInfo.FaultList[j].FaultCode
			})
		}
		if !isFaultRankMapEqual(jobFaultRankInfos, expectFaultRanks) {
			t.Errorf("processor.jobFaultInfos = %s, expectFaultRanks = %s",
				util.ObjToString(processor.jobFaultInfoMap), util.ObjToString(expectFaultRanks))
		}
		filterJobFaultRank := processor.getJobFaultRankInfosFilterLevel(NotHandleFault)
		if isContainsAny(util.ObjToString(filterJobFaultRank), NotHandleFault) {
			t.Errorf("processor.getJobFaultRankInfosFilterLevel = %s", util.ObjToString(filterJobFaultRank))

		}

		filterJobFaultRank = processor.getJobFaultRankInfosFilterLevel(RestartRequest)
		if isContainsAny(util.ObjToString(filterJobFaultRank), RestartRequest) {
			t.Errorf("processor.getJobFaultRankInfosFilterLevel = %s", util.ObjToString(filterJobFaultRank))
		}

		filterJobFaultRank = processor.getJobFaultRankInfosFilterLevel(RestartBusiness)
		if isContainsAny(util.ObjToString(filterJobFaultRank), RestartBusiness) {
			t.Errorf("processor.getJobFaultRankInfosFilterLevel = %s", util.ObjToString(filterJobFaultRank))
		}

		filterJobFaultRank = processor.getJobFaultRankInfosFilterLevel(FreeRestartNPU)
		if isContainsAny(util.ObjToString(filterJobFaultRank), FreeRestartNPU) {
			t.Errorf("processor.getJobFaultRankInfosFilterLevel = %s", util.ObjToString(filterJobFaultRank))
		}

		filterJobFaultRank = processor.getJobFaultRankInfosFilterLevel(RestartNPU)
		if isContainsAny(util.ObjToString(filterJobFaultRank), "\""+RestartNPU+"\"") {
			t.Errorf("processor.getJobFaultRankInfosFilterLevel = %s", util.ObjToString(filterJobFaultRank))
		}

		filterJobFaultRank = processor.getJobFaultRankInfosFilterLevel(SeparateNPU)
		if isContainsAny(util.ObjToString(filterJobFaultRank), SeparateNPU) {
			t.Errorf("processor.getJobFaultRankInfosFilterLevel = %s", util.ObjToString(filterJobFaultRank))
		}
	})
}
