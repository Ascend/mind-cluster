// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package job a series of job configmap operator function
package job

import (
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/api/errors"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/interface/kube"
)

const (
	configmapPrefix   = "job-summary"
	configmapLabel    = "outside-job-info"
	hcclJson          = "hccl.json"
	configmapOperator = "operator"
	operatorAdd       = "add"
	operatorDelete    = "delete"
	defaultHcclJson   = `{"status":"initializing"}`
	jobId             = "job_id"
	frameWorkKey      = "framework"
	jobName           = "job_name"
	deleteTime        = "deleteTime"
	cmIndex           = "cm_index"
	cmCutNumKey       = "total"
	jobStatus         = "job_status"
	addTime           = "time"
	atlasRing         = "ring-controller.atlas"
	val910            = "ascend-910"
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
	data[hcclJson] = defaultHcclJson
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
	data[hcclJson] = hccl
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
	data[hcclJson] = defaultHcclJson
	data[configmapOperator] = operatorDelete
	data[deleteTime] = strconv.Itoa(int(jobInfo.DeleteTime))
	data[cmCutNumKey] = strconv.Itoa(jobInfo.TotalCmNum)
	data[addTime] = strconv.Itoa(int(jobInfo.AddTime))
	for i := 0; i < jobInfo.TotalCmNum; i++ {
		var cmName string
		if i == 0 {
			cmName = fmt.Sprintf("%s-%s", configmapPrefix, jobInfo.Name)
		} else {
			cmName = fmt.Sprintf("%s-%s-%d", configmapPrefix, jobInfo.Name, i)
		}
		data[cmIndex] = fmt.Sprintf("%d", i)
		if i < len(hccls) {
			data[hcclJson] = hccls[i]
		}
		err := kube.CreateOrUpdateConfigMap(cmName, jobInfo.NameSpace, data, getDefaultLabel())
		if err != nil {
			hwlog.RunLog.Errorf("delete configmap failed, err: %v", err)
			continue
		}
	}
	return true
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
	label[atlasRing] = val910
	label[configmapLabel] = "true"
	return label
}
