// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package common is grpc common types and functions
package common

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	"k8s.io/api/core/v1"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"clusterd/pkg/common/util"
	"clusterd/pkg/interface/grpc/pb"
	"clusterd/pkg/interface/kube"
)

// GetJobInfo return job's name, podGroup name and namespace
func GetJobInfo(jobId string) (jobName, pgName, namespace string) {
	if kube.JobMgr == nil {
		hwlog.RunLog.Error("job mgr is nil")
		return "", "", ""
	}
	worker := kube.JobMgr.GetBsWorker(jobId)
	if worker == nil {
		hwlog.RunLog.Errorf("jobId=%s not exist", jobId)
		return "", "", ""
	}
	info := worker.GetBaseInfo()
	return info.JobName, info.PGName, info.Namespace
}

// Faults2String return string of faults
func Faults2String(faults []*pb.FaultRank) string {
	if len(faults) == 0 {
		return ""
	}
	faultInfo := make([]string, 0, len(faults))
	for _, item := range faults {
		faultInfo = append(faultInfo, item.RankId+":"+item.FaultType)
	}
	return strings.Join(faultInfo, ",")
}

// StrategySupported check strategy supported
func StrategySupported(strategy string) bool {
	return strategy == ProcessRetryStrategyName || strategy == ProcessRecoverStrategyName ||
		strategy == ProcessDumpStrategyName || strategy == ProcessExitStrategyName
}

// GetRecoverBaseInfo get recover config
func GetRecoverBaseInfo(name, namespace string) (RecoverConfig, RespCode, error) {
	config := RecoverConfig{}
	pg, err := kube.RetryGetPodGroup(name, namespace, GetPodGroupTimes)
	if err != nil {
		return config, OperatePodGroupError, err
	}
	_, config.PlatFormMode = pg.Annotations[ProcessRecoverStrategy]
	mindXConfig, ok := pg.Annotations[MindXRecoverStrategies]
	strategyList := strings.Split(mindXConfig, ",")
	for _, strategy := range strategyList {
		if StrategySupported(strategy) {
			config.MindXConfigStrategies = append(config.MindXConfigStrategies, strategy)
		}
	}
	config.MindXConfigStrategies = append(config.MindXConfigStrategies, ProcessExitStrategyName)
	value, ok := pg.Labels[ProcessReschedulingLabel]
	if !ok {
		hwlog.RunLog.Warn("can not find process rescheduling label")
		config.ProcessRescheduleOn = false
	}
	config.ProcessRescheduleOn = value == ProcessReschedulingEnable
	return config, OK, nil
}

// SendRetry send signal util send success or retry times upper retryTimes
func SendRetry(sender SignalRetrySender, signal *pb.ProcessManageSignal, retryTimes int) error {
	var err error
	for i := 0; i < retryTimes; i++ {
		err = sender.Send(signal)
		if err == nil {
			return nil
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	return err
}

// NewEventId return uuid according randLen
func NewEventId(randLen int) string {
	timestamp := time.Now().UnixNano()
	randomNumberHex := ""
	if randLen > MaxUuidRandomLength || randLen <= 0 {
		randLen = MaxUuidRandomLength
	}
	randomNumber := make([]byte, randLen)
	_, err := io.ReadFull(rand.Reader, randomNumber)
	if err == nil {
		randomNumberHex = hex.EncodeToString(randomNumber)
	}
	return fmt.Sprintf("%X-%s", timestamp, randomNumberHex)
}

// ChangeProcessSchedulingMode change process scheduling mode
func ChangeProcessSchedulingMode(taskName, namespace, mode string) (*v1beta1.PodGroup, error) {
	pg, err := kube.GetPodGroup(taskName, namespace)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get pg when change process scheduling, err: %v", err)
		return nil, err
	}
	_, ok := pg.Labels[ProcessReschedulingLabel]
	if !ok {
		hwlog.RunLog.Error("can not find process rescheduling label when change")
		return nil, fmt.Errorf("can not find process rescheduling label when change")
	}
	pg.Labels[ProcessReschedulingLabel] = mode
	return kube.UpdatePodGroup(pg)
}

// RetryWriteResetCM retry write the reset info configMap
func RetryWriteResetCM(taskName, nameSpace string, faultRankList []string, operator string) (*v1.ConfigMap, error) {
	var err error
	var configMap *v1.ConfigMap
	for i := 0; i < WriteResetInfoRetryTimes; i++ {
		time.Sleep(time.Duration(i) * time.Second) // first i==0, sleep zero second
		configMap, err = WriteResetInfoToCM(taskName, nameSpace, faultRankList, operator)
		if err == nil {
			return configMap, err
		}
	}
	return configMap, err
}

// WriteResetInfoToCM write the reset info configMap
func WriteResetInfoToCM(taskName, namespace string,
	faultRankList []string, operation string) (*v1.ConfigMap, error) {
	oldCM, err := kube.GetConfigMap(ResetInfoCMNamePrefix+taskName, namespace)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get reset cm of task %s, err is : %v", taskName, err)
		return nil, err
	}

	oldResetInfoData, ok := oldCM.Data[ResetInfoCMDataKey]
	if !ok {
		return nil, fmt.Errorf("invalid old reset info data")
	}
	var oldTaskInfo TaskResetInfo
	err = json.Unmarshal([]byte(oldResetInfoData), &oldTaskInfo)
	if err != nil {
		hwlog.RunLog.Errorf("failed to unmarshal reset info data, err: %v", err)
		return nil, fmt.Errorf("failed to unmarshal reset info data, err: %v", err)
	}
	newTaskInfo, err := setNewTaskInfo(oldTaskInfo, faultRankList, operation)
	if err != nil {
		hwlog.RunLog.Errorf("failed to set new task info, err: %v", err)
	}
	checkCode := util.MakeDataHash(newTaskInfo)
	var data []byte
	if data, err = json.Marshal(newTaskInfo); err != nil || len(data) == 0 {
		return nil, fmt.Errorf("marshal new task reset info data failed")
	}
	newCm := &v1.ConfigMap{
		TypeMeta:   oldCM.TypeMeta,
		ObjectMeta: oldCM.ObjectMeta,
		Data: map[string]string{
			ResetInfoCMDataKey:      string(data),
			ResetInfoCMCheckCodeKey: checkCode,
		},
	}
	return kube.UpdateConfigMap(newCm)
}

func setNewTaskInfo(oldTaskResetInfo TaskResetInfo,
	faultRankList []string, operation string) (TaskResetInfo, error) {
	var newTaskInfo TaskResetInfo
	newTaskInfo.RankList = []*TaskDevInfo{}
	newTaskInfo.UpdateTime = time.Now().Unix()
	newTaskInfo.RetryTime = oldTaskResetInfo.RetryTime
	if operation == RestartAllProcessOperation {
		newTaskInfo.RetryTime += 1
	}
	if operation != FaultRankStatus {
		return newTaskInfo, nil
	}
	for _, rank := range faultRankList {
		rankId, err := strconv.Atoi(rank)
		if err != nil {
			hwlog.RunLog.Errorf("failed to convert rank id %s to int", rank)
			return TaskResetInfo{}, err
		}
		newTaskInfo.RankList = append(newTaskInfo.RankList, &TaskDevInfo{
			RankId: rankId,
			DevFaultInfo: DevFaultInfo{
				Status: FaultRankStatus,
			},
		})
	}
	return newTaskInfo, nil
}

// GetFaultRankIdsInSameNode get all ranks in a node which has fault ranks
func GetFaultRankIdsInSameNode(faultRankIds []string, deviceNumPerNode int) []string {
	if deviceNumPerNode <= 0 || len(faultRankIds) == 0 {
		return faultRankIds
	}
	faultRanks := util.StringSliceToIntSlice(faultRankIds)
	sort.Ints(faultRanks)
	var faultRankIdsResult []string
	rankIdMap := make(map[int]struct{}, 0)
	for _, num := range faultRanks {
		rankIndexStart := num / deviceNumPerNode * deviceNumPerNode
		for i := rankIndexStart; i < rankIndexStart+deviceNumPerNode; i++ {
			if _, ok := rankIdMap[i]; ok {
				break
			}
			rankIdMap[i] = struct{}{}
			faultRankIdsResult = append(faultRankIdsResult, strconv.Itoa(i))
		}
	}
	return faultRankIdsResult
}

// CheckProcessRecoverOpen check whether process recover mode open
func CheckProcessRecoverOpen(name, nameSpace string) bool {
	pg, err := kube.GetPodGroup(name, nameSpace)
	if err != nil {
		hwlog.RunLog.Errorf("get pg err: %v", err)
		return false
	}
	_, ok := pg.Labels[ProcessReschedulingLabel]
	if !ok {
		hwlog.RunLog.Warn("can not find process rescheduling label")
		return false
	}
	return pg.Labels[ProcessReschedulingLabel] == ProcessReschedulingEnable
}
