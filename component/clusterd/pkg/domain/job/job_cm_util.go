// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package job a series of job configmap operator function
package job

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/interface/kube"
)

const (
	configmapPrefix = "job-summary"
	configmapLabel  = "outside-job-info"
	// HcclJson hccl info key of configmap
	HcclJson          = "hccl.json"
	configmapOperator = "operator"
	operatorAdd       = "add"
	operatorDelete    = "delete"
	jobId             = "job_id"
	frameWorkKey      = "framework"
	jobName           = "job_name"
	deleteTime        = "deleteTime"
	cmIndex           = "cm_index"
	cmCutNumKey       = "total"
	jobStatus         = "job_status"
	addTime           = "time"
	val910            = api.Ascend910MinuxCase
	ptFramework       = "pytorch"
	torIpTag          = "sharedTorIp"
	masterAddrKey     = "masterAddr"
)

func initCM(jobInfo constant.JobInfo) bool {
	data := make(map[string]string, cmDataInitLength)
	data[jobName] = jobInfo.Name
	data[frameWorkKey] = jobInfo.Framework
	data[jobId] = jobInfo.Key
	data[jobStatus] = jobInfo.Status
	data[cmIndex] = "0"
	data[cmCutNumKey] = strconv.Itoa(jobInfo.TotalCmNum)
	data[HcclJson] = constant.DefaultHcclJson
	data[configmapOperator] = operatorAdd
	data[addTime] = strconv.Itoa(int(jobInfo.AddTime))
	cmName := fmt.Sprintf("%s-%s", configmapPrefix, jobInfo.Name)
	if err := kube.CreateOrUpdateConfigMap(cmName, jobInfo.NameSpace, data, getDefaultLabel()); err != nil {
		hwlog.RunLog.Errorf("initCM CreateOrUpdateConfigMap err: %s", err)
		return false
	}
	return true
}

func updateCM(jobInfo constant.JobInfo, index int, hccl string) bool {
	data := make(map[string]string, cmDataInitLength)
	if jobInfo.Framework == ptFramework {
		data[torIpTag] = jobInfo.SharedTorIp
		data[masterAddrKey] = jobInfo.MasterAddr
	}
	data[jobName] = jobInfo.Name
	data[frameWorkKey] = jobInfo.Framework
	data[jobId] = jobInfo.Key
	data[jobStatus] = jobInfo.Status
	data[addTime] = strconv.Itoa(int(jobInfo.AddTime))
	data[configmapOperator] = operatorAdd
	data[cmIndex] = strconv.Itoa(index)
	data[cmCutNumKey] = strconv.Itoa(jobInfo.TotalCmNum)
	data[HcclJson] = hccl
	// deleteTime should be changed to updateTime next version
	if jobInfo.Status == StatusJobFail || jobInfo.Status == StatusJobCompleted {
		data[deleteTime] = strconv.Itoa(int(time.Now().Unix()))
	}
	var cmName string
	if index == 0 {
		cmName = fmt.Sprintf("%s-%s", configmapPrefix, jobInfo.Name)
	} else {
		cmName = fmt.Sprintf("%s-%s-%d", configmapPrefix, jobInfo.Name, index)
	}
	err := kube.UpdateOrCreateConfigMap(cmName, jobInfo.NameSpace, data, getDefaultLabel())
	if err != nil {
		hwlog.RunLog.Errorf("update configmap %s failed, err: %v", cmName, err)
		return false
	}
	return true
}

func preDeleteCM(jobInfo constant.JobInfo, hccls []string) bool {
	data := make(map[string]string, cmDataInitLength)
	if jobInfo.Framework == ptFramework {
		data[torIpTag] = jobInfo.SharedTorIp
		data[masterAddrKey] = jobInfo.MasterAddr
	}
	data[jobName] = jobInfo.Name
	data[frameWorkKey] = jobInfo.Framework
	data[jobStatus] = jobInfo.Status
	data[jobId] = jobInfo.Key
	data[HcclJson] = constant.DefaultHcclJson
	data[configmapOperator] = operatorDelete
	data[deleteTime] = strconv.Itoa(int(jobInfo.DeleteTime))
	data[cmCutNumKey] = strconv.Itoa(jobInfo.TotalCmNum)
	data[addTime] = strconv.Itoa(int(jobInfo.AddTime))
	result := true
	for i := 0; i < jobInfo.TotalCmNum; i++ {
		var cmName string
		if i == 0 {
			cmName = fmt.Sprintf("%s-%s", configmapPrefix, jobInfo.Name)
		} else {
			cmName = fmt.Sprintf("%s-%s-%d", configmapPrefix, jobInfo.Name, i)
		}
		data[cmIndex] = fmt.Sprintf("%d", i)
		if i < len(hccls) {
			data[HcclJson] = hccls[i]
		}
		err := kube.CreateOrUpdateConfigMap(cmName, jobInfo.NameSpace, data, getDefaultLabel())
		if err != nil {
			hwlog.RunLog.Errorf("create or update configmap failed, err: %v", err)
			result = false
			continue
		}
	}
	return result
}

func deleteCm(jobInfo constant.JobInfo) bool {
	for i := 0; i < jobInfo.TotalCmNum; i++ {
		var cmName string
		if i == 0 {
			cmName = fmt.Sprintf("%s-%s", configmapPrefix, jobInfo.Name)
		} else {
			cmName = fmt.Sprintf("%s-%s-%d", configmapPrefix, jobInfo.Name, i)
		}
		err := kube.DeleteConfigMap(cmName, jobInfo.NameSpace)
		if errors.IsNotFound(err) {
			continue
		} else if err != nil {
			hwlog.RunLog.Errorf("delete configmap failed, name is %s", cmName)
			return false
		}
	}
	return true
}

func getDefaultLabel() map[string]string {
	label := make(map[string]string)
	label[api.AtlasTaskLabel] = val910
	label[configmapLabel] = "true"
	return label
}

func refreshFaultJobInfo() {
	for i := 0; i < constant.RetryTime; i++ {
		cm, getErr := kube.GetConfigMap(api.FaultJobCmName, api.ClusterNS)
		if getErr != nil {
			if errors.IsNotFound(getErr) {
				hwlog.RunLog.Warnf("get configmap fault-job-info err:%v", getErr)
				return
			}
			hwlog.RunLog.Errorf("get configmap fault-job-info err:%v", getErr)
			continue
		}
		cm.Data = refreshFaultJobInfoCmData(cm.Data)
		if _, updateErr := kube.UpdateConfigMap(cm); updateErr != nil {
			hwlog.RunLog.Errorf("update configmap fault-job-info err:%v", updateErr)
			time.Sleep(time.Second)
			continue
		}
		return
	}
}

// RefreshFaultJobInfo refresh configmap fault-job-info
func RefreshFaultJobInfo(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Info("job RefreshFaultJobInfo stop work")
			return
		default:
			refreshFaultJobInfo()
			time.Sleep(time.Hour)
		}
	}
}

// refreshFaultJobInfoCmData  refresh configmap fault-job-info data
func refreshFaultJobInfoCmData(datas map[string]string) map[string]string {
	newData := make(map[string]string)
	for jobUid, data := range datas {
		if _, ok := jobSummaryMap.Load(jobUid); !ok {
			continue
		}
		newData[jobUid] = data
	}
	return newData
}
