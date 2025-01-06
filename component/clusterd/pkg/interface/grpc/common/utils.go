// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package common is grpc common types and functions
package common

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/pod"
	"clusterd/pkg/interface/grpc/pb"
	"clusterd/pkg/interface/kube"
)

var faultSplitLength = 2

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

// Faults2Ranks return rank slice of faults
func Faults2Ranks(faults []*pb.FaultRank) []string {
	if len(faults) == 0 {
		return nil
	}
	ranks := make([]string, 0, len(faults))
	for _, item := range faults {
		ranks = append(ranks, item.RankId)
	}
	return ranks
}

// String2Faults return faults split from string
func String2Faults(faultStr string) []*pb.FaultRank {
	faultStr = strings.TrimSpace(faultStr)
	faultStr = strings.Trim(faultStr, ",")
	if faultStr == "" {
		return nil
	}
	faultStrSlice := strings.Split(faultStr, ",")
	var res []*pb.FaultRank
	for _, fault := range faultStrSlice {
		fs := strings.Split(fault, ":")
		n := len(fs)
		if n == faultSplitLength {
			res = append(res, &pb.FaultRank{
				RankId:    fs[0],
				FaultType: fs[n-1],
			})
		} else {
			hwlog.RunLog.Warnf("bad format, fault=%s, faultStr=%s", fault, faultStr)
		}
	}
	return res
}

// StrategySupported check strategy supported
func StrategySupported(strategy string) bool {
	return strategy == constant.ProcessRetryStrategyName || strategy == constant.ProcessRecoverStrategyName ||
		strategy == constant.ProcessDumpStrategyName || strategy == constant.ProcessExitStrategyName
}

// GetRecoverBaseInfo get recover config
func GetRecoverBaseInfo(name, namespace string) (RecoverConfig, RespCode, error) {
	config := RecoverConfig{}
	pg, err := kube.RetryGetPodGroup(name, namespace, constant.GetPodGroupTimes)
	if err != nil {
		return config, OperatePodGroupError, err
	}
	_, config.PlatFormMode = pg.Annotations[constant.ProcessRecoverStrategy]
	mindXConfig, ok := pg.Annotations[constant.RecoverStrategies]
	strategyList := strings.Split(mindXConfig, ",")
	for _, strategy := range strategyList {
		if StrategySupported(strategy) {
			config.MindXConfigStrategies = append(config.MindXConfigStrategies, strategy)
		}
	}
	config.MindXConfigStrategies = append(config.MindXConfigStrategies, constant.ProcessExitStrategyName)
	value, ok := pg.Labels[constant.ProcessRecoverEnableLabel]
	if !ok {
		hwlog.RunLog.Warn("can not find process rescheduling label")
		config.ProcessRecoverEnable = false
	}
	config.ProcessRecoverEnable = value == constant.ProcessRecoverEnable
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
	if randLen > constant.MaxUuidRandomLength || randLen <= 0 {
		randLen = constant.MaxUuidRandomLength
	}
	randomNumber := make([]byte, randLen)
	_, err := io.ReadFull(rand.Reader, randomNumber)
	if err == nil {
		randomNumberHex = hex.EncodeToString(randomNumber)
	}
	return fmt.Sprintf("%X-%s", timestamp, randomNumberHex)
}

// ChangeProcessRecoverEnableMode change process scheduling mode
func ChangeProcessRecoverEnableMode(jobInfo JobBaseInfo, mode string) (*v1beta1.PodGroup, error) {
	label := map[string]string{constant.ProcessRecoverEnableLabel: mode}
	return kube.RetryPatchPodGroupLabel(jobInfo.PgName, jobInfo.Namespace, constant.RetryTime, label)
}

// RetryWriteResetCM retry write the reset info configMap
func RetryWriteResetCM(taskName, nameSpace string, faultRankList []string, operator string) (*v1.ConfigMap, error) {
	var err error
	var configMap *v1.ConfigMap
	for i := 0; i < constant.WriteResetInfoRetryTimes; i++ {
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
	oldCM, err := kube.GetConfigMap(constant.ResetInfoCMNamePrefix+taskName, namespace)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get reset cm of task %s, err is : %v", taskName, err)
		return nil, err
	}

	oldResetInfoData, ok := oldCM.Data[constant.ResetInfoCMDataKey]
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
			constant.ResetInfoCMDataKey:      string(data),
			constant.ResetInfoCMCheckCodeKey: checkCode,
		},
	}
	return kube.UpdateConfigMap(newCm)
}

func setNewTaskInfo(oldTaskResetInfo TaskResetInfo,
	faultRankList []string, operation string) (TaskResetInfo, error) {
	for _, rank := range oldTaskResetInfo.RankList {
		if rank.Policy == constant.HotResetPolicy {
			return TaskResetInfo{}, errors.New("hotReset=1 is not compatible with process-recover")
		}
	}
	var newTaskInfo TaskResetInfo
	newTaskInfo.RankList = []*TaskDevInfo{}
	newTaskInfo.UpdateTime = time.Now().Unix()
	newTaskInfo.RetryTime = oldTaskResetInfo.RetryTime
	if operation != constant.NotifyFaultFlushingOperation {
		newTaskInfo.FaultFlushing = false
	} else {
		newTaskInfo.FaultFlushing = true
	}
	if operation == constant.RestartAllProcessOperation {
		newTaskInfo.RetryTime += 1
		return newTaskInfo, nil
	}
	if operation != constant.NotifyFaultListOperation {
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
				Status: constant.FaultRankStatus,
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
	_, ok := pg.Labels[constant.ProcessRecoverEnableLabel]
	if !ok {
		hwlog.RunLog.Warn("can not find process rescheduling label")
		return false
	}
	return pg.Labels[constant.ProcessRecoverEnableLabel] == constant.ProcessRecoverEnable
}

// RemoveSliceDuplicateFaults remote duplicate fault
func RemoveSliceDuplicateFaults(faults []*pb.FaultRank) []*pb.FaultRank {
	var res = make([]*pb.FaultRank, 0)
	exitMap := make(map[string]string)
	for _, fault := range faults {
		if typ, ok := exitMap[fault.RankId]; !ok {
			exitMap[fault.RankId] = fault.FaultType
		} else {
			if typ == constant.UceFaultType {
				exitMap[fault.RankId] = fault.FaultType
			}
		}
	}
	for id, typ := range exitMap {
		res = append(res, &pb.FaultRank{
			RankId:    id,
			FaultType: typ,
		})
	}
	return res
}

// LabelFaultPod label fault for software fault
func LabelFaultPod(jobId string, rankList []string, labeledMap map[string]string) (map[string]string, error) {
	var faultPodRankList []string
	devicePerNode := pod.GetPodDeviceNumByJobId(jobId)
	for _, rank := range rankList {
		faultRank, err := strconv.Atoi(rank)
		if err != nil {
			hwlog.RunLog.Errorf("parse pod rank failed, err is %v", err)
			return nil, err
		}
		faultPodRank := faultRank / devicePerNode
		faultPodRankList = append(faultPodRankList, strconv.Itoa(faultPodRank))
	}
	faultPodRankList = util.RemoveSliceDuplicateElement(faultPodRankList)
	podMap, err := labelPodFault(jobId, faultPodRankList, labeledMap)
	if err != nil {
		hwlog.RunLog.Errorf("label fault pod failed, err is %v", err)
		return podMap, fmt.Errorf("label fault pod failed, err is %v", err)
	}
	return podMap, nil
}

// GetPodMap return a dict, key is fault pod rank, value is pod id
func GetPodMap(jobId string, rankList []string) map[string]string {
	podMap := make(map[string]string)
	devicePerNode := pod.GetPodDeviceNumByJobId(jobId)
	for _, rank := range rankList {
		faultRank, err := strconv.Atoi(rank)
		if err != nil {
			hwlog.RunLog.Warnf("parse pod rank failed, err is %v", err)
			continue
		}
		faultPodRank := faultRank / devicePerNode
		podRank := strconv.Itoa(faultPodRank)
		_, ok := podMap[podRank]
		if ok {
			continue
		}
		pod := pod.GetPodByRankIndex(jobId, podRank)
		if pod.Name == "" {
			hwlog.RunLog.Warnf("discard nil pod, jobId=%s", jobId)
			continue
		}
		podMap[podRank] = string(pod.UID)
	}
	return podMap
}

func labelPodFault(jobId string, faultPodRankList []string, labeledMap map[string]string) (map[string]string, error) {
	if labeledMap == nil {
		labeledMap = make(map[string]string)
	}
	faultLabel := map[string]string{"fault-type": "software"}
	var err error = nil
	for _, podRank := range faultPodRankList {
		_, labeled := labeledMap[podRank]
		if labeled {
			continue
		}
		pod := pod.GetPodByRankIndex(jobId, podRank)
		if pod.Name == "" {
			hwlog.RunLog.Infof("discard nil pod, jobId=%s", jobId)
			continue
		}
		if patchErr := kube.RetryPatchPodLabels(pod.Name, pod.Namespace,
			constant.UpdatePodGroupTimes, faultLabel); patchErr != nil {
			hwlog.RunLog.Infof("patch pod label error, jobId=%s, err=%v", jobId, patchErr)
			err = patchErr
		}
		labeledMap[podRank] = string(pod.UID)
	}
	return labeledMap, err
}

// FaultPodAllRescheduled check if all fault pod rescheduled
func FaultPodAllRescheduled(jobId string, oldPodMap map[string]string) bool {
	for podRank, oldPodId := range oldPodMap {
		pod := pod.GetPodByRankIndex(jobId, podRank)
		if pod.Name == "" {
			return false
		}
		if oldPodId == string(pod.UID) {
			return false
		}
	}
	return true
}

// IsUceFault check whether fault type is uce fault
func IsUceFault(faults []*pb.FaultRank) bool {
	for _, fault := range faults {
		if fault.FaultType == constant.NormalFaultType {
			return false
		}
	}
	return true
}
