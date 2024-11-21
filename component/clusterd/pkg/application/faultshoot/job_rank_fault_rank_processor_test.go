package faultshoot

import (
	"strings"
	"testing"

	"clusterd/pkg/application/job"
	"clusterd/pkg/common/util"
	"clusterd/pkg/interface/kube"
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
		cmDeviceInfos, jobsPodWorkers, expectFaultRanks, err := readObjectFromJobFaultRankTestYaml()
		if err != nil {
			t.Errorf("%v", err)
		}
		deviceFaultProcessCenter.setProcessingCm(cmDeviceInfos)
		kube.JobMgr = &job.Agent{BsWorker: jobsPodWorkers}
		processor.deviceCenter.jobServerInfoMap = kube.JobMgr.GetJobServerInfoMap()
		processor.process()
		if !isFaultRankMapEqual(processor.getJobFaultRankInfos(), expectFaultRanks) {
			t.Errorf("processor.jobFaultInfos = %s, expectFaultRanks = %s",
				util.ObjToString(processor.jobFaultInfoMap), util.ObjToString(expectFaultRanks))
		}
	})

	t.Run("TestJobRankFaultInfoProcessor_getJobFaultRankInfosFilterLevel", func(t *testing.T) {
		cmDeviceInfos, jobsPodWorkers, expectFaultRanks, err := readObjectFromJobFaultRankTestYaml()
		if err != nil {
			t.Errorf("%v", err)
		}
		deviceFaultProcessCenter.setProcessingCm(cmDeviceInfos)
		kube.JobMgr = &job.Agent{BsWorker: jobsPodWorkers}
		processor.deviceCenter.jobServerInfoMap = kube.JobMgr.GetJobServerInfoMap()
		processor.process()
		if !isFaultRankMapEqual(processor.getJobFaultRankInfos(), expectFaultRanks) {
			t.Errorf("processor.jobFaultInfos = %s, expectFaultRanks = %s",
				util.ObjToString(processor.jobFaultInfoMap), util.ObjToString(expectFaultRanks))
		}
		filterJobFaultRank := processor.getJobFaultRankInfosFilterLevel(1)
		if isContainsAny(util.ObjToString(filterJobFaultRank), NotHandleFaultDesc) {
			t.Errorf("processor.getJobFaultRankInfosFilterLevel = %s", util.ObjToString(filterJobFaultRank))

		}

		filterJobFaultRank = processor.getJobFaultRankInfosFilterLevel(2)
		if isContainsAny(util.ObjToString(filterJobFaultRank), NotHandleFaultDesc, RestartRequestDesc) {
			t.Errorf("processor.getJobFaultRankInfosFilterLevel = %s", util.ObjToString(filterJobFaultRank))
		}

		filterJobFaultRank = processor.getJobFaultRankInfosFilterLevel(3)
		if isContainsAny(util.ObjToString(filterJobFaultRank), NotHandleFaultDesc, RestartRequestDesc,
			RestartBusinessDesc) {
			t.Errorf("processor.getJobFaultRankInfosFilterLevel = %s", util.ObjToString(filterJobFaultRank))
		}

		filterJobFaultRank = processor.getJobFaultRankInfosFilterLevel(4)
		if isContainsAny(util.ObjToString(filterJobFaultRank), NotHandleFaultDesc, RestartRequestDesc,
			RestartBusinessDesc, FreeRestartNPUDesc) {
			t.Errorf("processor.getJobFaultRankInfosFilterLevel = %s", util.ObjToString(filterJobFaultRank))
		}

		filterJobFaultRank = processor.getJobFaultRankInfosFilterLevel(5)
		if isContainsAny(util.ObjToString(filterJobFaultRank), NotHandleFaultDesc, RestartRequestDesc,
			RestartBusinessDesc, FreeRestartNPUDesc, RestartNPUDesc) {
			t.Errorf("processor.getJobFaultRankInfosFilterLevel = %s", util.ObjToString(filterJobFaultRank))
		}

		filterJobFaultRank = processor.getJobFaultRankInfosFilterLevel(6)
		if isContainsAny(util.ObjToString(filterJobFaultRank), NotHandleFaultDesc, RestartRequestDesc,
			RestartBusinessDesc, FreeRestartNPUDesc, RestartNPUDesc, SeparateNPUDesc) {
			t.Errorf("processor.getJobFaultRankInfosFilterLevel = %s", util.ObjToString(filterJobFaultRank))
		}
	})
}
