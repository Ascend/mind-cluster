package fault

import (
	"clusterd/pkg/application/job"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/device"
	"clusterd/pkg/interface/kube"
	"sync"
)

type JobRankFaultInfoProcessor struct {
	deviceCenter  *DeviceFaultProcessCenter
	jobFaultInfos map[string]FaultInfo
	mutex         sync.RWMutex
}

func (processor *JobRankFaultInfoProcessor) GetJobFaultRankInfos() map[string]FaultInfo {
	processor.mutex.RLock()
	defer processor.mutex.RUnlock()
	return processor.jobFaultInfos
}

func (processor *JobRankFaultInfoProcessor) SetJobFaultRankInfos(faultInfos map[string]FaultInfo) {
	processor.mutex.RLock()
	defer processor.mutex.RUnlock()
	processor.jobFaultInfos = faultInfos
}

func (processor *JobRankFaultInfoProcessor) Process() {
	deviceInfos := processor.deviceCenter.GetDeviceInfos()
	nodesName := getNodesNameFromDeviceInfo(deviceInfos)
	kube.JobMgr.RwMutex.RLock()
	defer kube.JobMgr.RwMutex.RUnlock()
	jobFaultInfos := make(map[string]FaultInfo)
	for jobId, worker := range kube.JobMgr.BsWorker {
		jobFaultInfo := FaultInfo{
			JobId:     jobId,
			FaultList: make([]FaultRank, 0),
		}

		workerInfo := worker.GetWorkerInfo()
		serverList := workerInfo.CMData.GetServerList()
		for _, nodeName := range nodesName {
			faultRankList := findFaultRankForJob(deviceInfos, nodeName, serverList, jobId)
			jobFaultInfo.FaultList = append(jobFaultInfo.FaultList, faultRankList...)
		}
		jobFaultInfos[jobId] = jobFaultInfo
	}
	processor.SetJobFaultRankInfos(jobFaultInfos)
}

func findFaultRankForJob(
	deviceInfos map[string]*constant.DeviceInfo, nodeName string, serverList []*job.ServerHccl, jobId string) []FaultRank {
	faultMap := device.GetFaultMap(deviceInfos[nodeNameToCmName(nodeName)])
	devicesOfJobOnNode := getDevicesNameOfJobOnNode(nodeName, serverList, jobId)
	faultRankList := make([]FaultRank, 0)
	if len(devicesOfJobOnNode) == 0 {
		return faultRankList
	}
	for _, deviceInfo := range devicesOfJobOnNode {
		deviceName := deviceID2DeviceKey(deviceInfo.DeviceID)
		faultList, ok := faultMap[deviceName]
		if !ok {
			continue
		}
		for _, fault := range faultList {
			faultRankList = append(faultRankList, FaultRank{
				RankId:    deviceInfo.RankID,
				FaultCode: fault.FaultCode,
			})
		}
	}
	return faultRankList
}

type FaultRank struct {
	RankId    string
	FaultCode string
}

type FaultInfo struct {
	JobId     string
	FaultList []FaultRank
}
