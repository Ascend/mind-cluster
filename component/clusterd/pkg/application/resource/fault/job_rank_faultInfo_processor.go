package fault

import (
	"clusterd/pkg/application/job"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/device"
	"clusterd/pkg/interface/kube"
)

type JobRankFaultInfoProcessor struct {
	deviceCenter  *DeviceFaultProcessCenter
	jobFaultInfos map[string]FaultInfo
}

func (processor *JobRankFaultInfoProcessor) GetJobFaultRankInfos() map[string]FaultInfo {
	return processor.jobFaultInfos
}

func (processor *JobRankFaultInfoProcessor) Process() {
	deviceInfos := processor.deviceCenter.GetDeviceInfos()
	nodesName := getNodesNameFromDeviceInfo(deviceInfos)
	kube.JobMgr.RwMutex.RLock()
	defer kube.JobMgr.RwMutex.RUnlock()
	if processor.jobFaultInfos == nil {
		processor.jobFaultInfos = make(map[string]FaultInfo)
	}
	for jobId, worker := range kube.JobMgr.BsWorker {
		jobFaultInfo := FaultInfo{
			jobId:     jobId,
			faultList: make([]FaultRank, 0),
		}

		workerInfo := worker.GetWorkerInfo()
		serverList := workerInfo.CMData.GetServerList()
		for _, nodeName := range nodesName {
			faultRankList := findFaultOnNodeForJob(deviceInfos, nodeName, serverList, jobId)
			jobFaultInfo.faultList = append(jobFaultInfo.faultList, faultRankList...)
		}
		processor.jobFaultInfos[jobId] = jobFaultInfo
	}
}

func findFaultOnNodeForJob(
	deviceInfos map[string]*constant.DeviceInfo, nodeName string, serverList []*job.ServerHccl, jobId string) []FaultRank {
	faultMap := device.GetFaultMap(deviceInfos[nodeName])
	devicesOfJobOnNode := getDevicesNameOfJobOnNode(nodeName, serverList, jobId)
	faultRankList := make([]FaultRank, 0)
	if len(devicesOfJobOnNode) == 0 {
		for _, deviceInfo := range devicesOfJobOnNode {
			deviceName := deviceID2DeviceKey(deviceInfo.DeviceID)
			if faultList, ok := faultMap[deviceName]; ok {
				for _, fault := range faultList {
					faultRankList = append(faultRankList, FaultRank{
						rankId:    deviceInfo.RankID,
						faultCode: fault.FaultCode,
					})
				}
			}
		}
	}
	return faultRankList
}

type FaultRank struct {
	rankId    string
	faultCode string
}

type FaultInfo struct {
	jobId     string
	faultList []FaultRank
}
