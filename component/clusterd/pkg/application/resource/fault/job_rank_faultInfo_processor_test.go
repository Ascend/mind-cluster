package fault

import (
	"clusterd/pkg/application/job"
	"clusterd/pkg/common/util"
	"clusterd/pkg/interface/kube"
	"testing"
)

func TestJobRankFaultInfoProcessor_GetJobFaultRankInfos(t *testing.T) {
	deviceFaultProcessCenter := NewDeviceFaultProcessCenter()
	processor, err := deviceFaultProcessCenter.GetJobFaultRankProcessor()
	if err != nil {
		t.Errorf("%v", err)
	}

	t.Run("TestJobRankFaultInfoProcessor_GetJobFaultRankInfos", func(t *testing.T) {
		cmDeviceInfos, jobsPodWorkers, expectFaultRanks, err := readObjectFromJobFaultRankTestYaml()
		if err != nil {
			t.Errorf("%v", err)
		}
		deviceFaultProcessCenter.setDeviceInfos(cmDeviceInfos)
		kube.JobMgr = &job.Agent{BsWorker: jobsPodWorkers}
		processor.Process()
		if !isFaultRankMapEqual(processor.jobFaultInfos, expectFaultRanks) {
			t.Errorf("processor.jobFaultInfos = %s, expectFaultRanks = %s",
				util.ObjToString(processor.jobFaultInfos), util.ObjToString(expectFaultRanks))
		}
	})
}
