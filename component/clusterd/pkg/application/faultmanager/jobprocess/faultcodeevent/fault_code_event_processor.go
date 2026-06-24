// Copyright (c) Huawei Technologies Co., Ltd. 22026. All rights reserved.

// Package faultcodeevent contain fault code event log processor
package faultcodeevent

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager/jobprocess/faultrank"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/logs"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/statistics"
)

// Processor collects fault codes from active job faults and logs them to job event log
var Processor = &FaultCodeEventLogProcessor{
	processedFaults: make(map[string]string),
}

// FaultCodeEventLogProcessor collects fault codes from active job faults and logs them to job event log
type FaultCodeEventLogProcessor struct {
	processedFaults map[string]string
	mu              sync.Mutex
}

// Process collects fault codes from active job faults, updates JobStatistic and logs to job event log
func (p *FaultCodeEventLogProcessor) Process(info any) any {
	jobFaultInfos := faultrank.JobFaultRankProcessor.GetJobFaultRankInfos()

	p.mu.Lock()
	defer p.mu.Unlock()

	for jobId := range p.processedFaults {
		if _, ok := jobFaultInfos[jobId]; !ok {
			delete(p.processedFaults, jobId)
		}
	}

	for k8sJobID, jobFaultInfo := range jobFaultInfos {
		if len(jobFaultInfo.FaultList) == 0 {
			continue
		}

		faultDigest := buildFaultDigest(jobFaultInfo.FaultDevice)
		if p.processedFaults[k8sJobID] == faultDigest {
			continue
		}
		p.processedFaults[k8sJobID] = faultDigest

		faultCodes := collectActiveFaultCodes(&jobFaultInfo)
		if len(faultCodes) == 0 {
			continue
		}
		statistics.JobStcMgrInst.UpdateJobStatistic(k8sJobID, func(jobStc *constant.JobStatisticV2) {
			jobStc.FaultCodesAndTimestamp = appendFaultCodes(k8sJobID,
				jobStc.FaultCodesAndTimestamp, faultCodes)
			logs.JobEventLog.Infof("Job Event: %s", util.ObjToString(jobStc))
		})
	}
	return nil
}

func buildFaultDigest(faultList []constant.FaultDevice) string {
	if len(faultList) == 0 {
		return ""
	}
	keys := make([]string, 0, len(faultList))
	for _, fault := range faultList {
		keys = append(keys, fmt.Sprintf("%s|%s|%s|%d|%s",
			fault.ServerId, fault.FaultCode, fault.FaultLevel, fault.FaultTime, fault.DeviceId))
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}

func collectActiveFaultCodes(jobFaultInfo *constant.JobFaultInfo) []constant.FaultCodeAndTimestamp {
	var result []constant.FaultCodeAndTimestamp
	seen := make(map[string]struct{})

	for _, faultDevice := range jobFaultInfo.FaultDevice {
		if faultDevice.FaultCode == "" {
			continue
		}
		faultKey := faultDevice.ServerName + faultDevice.DeviceId + faultDevice.FaultCode +
			strconv.FormatInt(faultDevice.FaultTime, 10)
		if _, dup := seen[faultKey]; dup {
			continue
		}
		seen[faultKey] = struct{}{}
		faultTime := faultDevice.FaultTime
		if faultTime == 0 {
			faultTime = time.Now().UnixMilli()
		}
		result = append(result, constant.FaultCodeAndTimestamp{
			FaultCode:  faultDevice.FaultCode,
			Timestamp:  faultTime,
			NodeName:   faultDevice.ServerName,
			DeviceId:   faultDevice.DeviceId,
			FaultLevel: faultDevice.FaultLevel,
		})
	}
	return result
}

func appendFaultCodes(k8sJobID string, existing []constant.FaultCodeAndTimestamp,
	newFaults []constant.FaultCodeAndTimestamp) []constant.FaultCodeAndTimestamp {
	result := existing
	seen := make(map[string]struct{}, len(existing))
	for _, e := range result {
		seen[buildFaultCodeKey(e)] = struct{}{}
	}
	for _, f := range newFaults {
		key := buildFaultCodeKey(f)
		if _, duplicated := seen[key]; duplicated {
			continue
		}
		if len(result) >= constant.MaxTimestampRecords {
			hwlog.RunLog.Warnf("job %s faultCodes slice length is over %v", k8sJobID, constant.MaxTimestampRecords)
			result = result[1:]
		}
		seen[key] = struct{}{}
		result = append(result, f)
	}
	return result
}

func buildFaultCodeKey(f constant.FaultCodeAndTimestamp) string {
	return fmt.Sprintf("%s|%d|%s", f.FaultCode, f.Timestamp, f.FaultLevel)
}
