// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package resource a series of resource function
package resource

import (
	"context"
	"strconv"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/faultmanager"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/device"
	"clusterd/pkg/domain/node"
	"clusterd/pkg/domain/switchinfo"
	"clusterd/pkg/interface/kube"
)

var (
	processCount              = 0
	maxTimePerSecond          = 5
	atLeastReportCycle        = int64(5)
	currentClusterDeviceCmNum = 0
	currentClusterNodeCmNum   = 0
	currentClusterSwitchCmNum = 0
	initTime                  int64
	updateChan                = make(chan int, 5)
	reportTime                int64
	cycleTicker               *time.Ticker
)

// AddNewMessageTotal when receive new device info or receive new node info or event 5s,add message to chan
func AddNewMessageTotal() {
	select {
	case updateChan <- constant.AllProcessType:
	default:
		hwlog.RunLog.Warnf("AddNewMessageTotal failed")
	}
}

// Report new message report to configmaps, the number of configmap is determined by the number of messages
func Report(ctx context.Context) {
	initTime = time.Now().UnixMilli()
	reportTime = time.Now().UnixMilli()
	timeSleepInitOnce := sync.Once{}
	faultmanager.GlobalFaultProcessCenter.Register(updateChan, constant.AllProcessType)
	go cycleReport()
	for {
		select {
		case whichToReport, ok := <-updateChan:
			if !ok {
				hwlog.RunLog.Errorf("catch invalid update signal")
				return
			}
			timeSleepInitOnce.Do(func() {
				// when informer begin, frequent add messages
				time.Sleep(time.Second)
			})
			switch whichToReport {
			case constant.DeviceProcessType:
				deviceArr := device.GetSafeData(faultmanager.QueryDeviceInfoToReport())
				updateDeviceInfoCm(deviceArr)
			case constant.NodeProcessType:
				nodeArr := node.GetSafeData(faultmanager.QueryNodeInfoToReport())
				updateNodeInfoCm(nodeArr)
			case constant.SwitchProcessType:
				switchArr := switchinfo.GetSafeData(faultmanager.QuerySwitchInfoToReport())
				updateSwitchInfoCm(switchArr)
			case constant.AllProcessType:
				deviceArr := device.GetSafeData(faultmanager.QueryDeviceInfoToReport())
				nodeArr := node.GetSafeData(faultmanager.QueryNodeInfoToReport())
				switchArr := switchinfo.GetSafeData(faultmanager.QuerySwitchInfoToReport())
				updateAllCm(deviceArr, nodeArr, switchArr)
			default:
				hwlog.RunLog.Errorf("unhandled type %d", whichToReport)
			}
			reportTime = time.Now().UnixMilli()
			processCount++
			limitRate()
		case <-ctx.Done():
			hwlog.RunLog.Info("reporter stop work")
			return
		}
	}
}

func limitRate() {
	if processCount < maxTimePerSecond {
		return
	}
	processCount = 0
	if time.Now().UnixMilli()-initTime < time.Second.Milliseconds() {
		time.Sleep(time.Second)
	}
	initTime = time.Now().UnixMilli()
}

func cycleReport() {
	cycleTicker = time.NewTicker(1 * time.Second)
	for {
		select {
		case _, ok := <-cycleTicker.C:
			if !ok {
				hwlog.RunLog.Errorf("catch invalid signal")
				return
			}
			if time.Now().UnixMilli()-reportTime > atLeastReportCycle*time.Second.Milliseconds() {
				reportTime = time.Now().UnixMilli()
				AddNewMessageTotal()
			}
		}
	}
}

// StopReport when leader is lost, close update chan and stop cycle task
func StopReport() {
	close(updateChan)
	if cycleTicker != nil {
		cycleTicker.Stop()
	}
}

func updateAllCm(deviceArr, nodeArr, switchArr []string) {
	updateSwitchInfoCm(switchArr)
	updateNodeInfoCm(nodeArr)
	updateDeviceInfoCm(deviceArr)
}

func updateSwitchInfoCm(switchArr []string) {
	if currentClusterSwitchCmNum < len(switchArr) {
		currentClusterSwitchCmNum = len(switchArr)
	}
	for i := 0; i < currentClusterSwitchCmNum; i++ {
		cmName := constant.ClusterSwitchInfo + strconv.Itoa(i)
		cmContent := ""
		if i < len(switchArr) {
			cmContent = switchArr[i]
		}
		updateConfig(cmName, cmContent)
	}
}

func updateNodeInfoCm(nodeArr []string) {
	if currentClusterNodeCmNum < len(nodeArr) {
		currentClusterNodeCmNum = len(nodeArr)
	}
	for i := 0; i < len(nodeArr) || i < currentClusterNodeCmNum; i++ {
		cmName := constant.ClusterNodeInfo + strconv.Itoa(i)
		cmContent := ""
		if i < len(nodeArr) {
			cmContent = nodeArr[i]
		}
		updateConfig(cmName, cmContent)
	}
}

func updateDeviceInfoCm(deviceArr []string) {
	if currentClusterDeviceCmNum < len(deviceArr) {
		currentClusterDeviceCmNum = len(deviceArr)
	}
	for i := 0; i < len(deviceArr) || i < currentClusterDeviceCmNum; i++ {
		cmName := constant.ClusterDeviceInfo + strconv.Itoa(i)
		cmContent := ""
		if i < len(deviceArr) {
			cmContent = deviceArr[i]
		}
		updateConfig(cmName, cmContent)
	}
}

func updateConfig(cmName, data string) {
	newClusterCM := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: api.DLNamespace,
			Labels:    map[string]string{constant.CmConsumer: constant.CmConsumerValue},
		},
		Data: map[string]string{cmName: data},
	}
	if _, err := kube.UpdateConfigMap(newClusterCM); err != nil {
		if !errors.IsNotFound(err) {
			hwlog.RunLog.Errorf("update cm failed, err is %v", err)
			return
		}
		if _, err = kube.CreateConfigMap(newClusterCM); err != nil {
			hwlog.RunLog.Errorf("cm is not fount, add cm failed, err is %v", err)
		}
	}
}
