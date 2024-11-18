// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package resource a series of resource function
package resource

import (
	"strconv"
	"sync"
	"time"

	"huawei.com/npu-exporter/v6/common-utils/hwlog"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	updateChan                = make(chan int, 1)
	reportTime                int64
	cycleTicker               *time.Ticker
)

// AddNewMessageTotal when receive new device info or receive new node info or event 5s,add message to chan
func AddNewMessageTotal() {
	if len(updateChan) == 0 {
		updateChan <- 0
	}
}

// Report new message report to configmaps, the number of configmap is determined by the number of messages
func Report() {
	initTime = time.Now().UnixMilli()
	reportTime = time.Now().UnixMilli()
	timeSleepInitOnce := sync.Once{}
	go cycleReport()
	for {
		select {
		case _, ok := <-updateChan:
			if !ok {
				hwlog.RunLog.Errorf("catch invalid update signal")
				return
			}
			timeSleepInitOnce.Do(func() {
				// when informer begin, frequent add messages
				time.Sleep(time.Second)
			})
			cmManager.Lock()
			deviceArr := device.GetSafeData(cmManager.deviceInfoMap)
			nodeArr := node.GetSafeData(cmManager.nodeInfoMap)
			switchArr := switchinfo.GetSafeData(cmManager.switchInfoMap)
			cmManager.Unlock()
			updateCmWithEmpty(deviceArr, nodeArr, switchArr)
			reportTime = time.Now().UnixMilli()
			processCount++
			limitRate()
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

// updateCmWithEmpty if the length of the deviceArr or nodeArr decreases, use "" to flush extra configmap
func updateCmWithEmpty(deviceArr, nodeArr, switchArr []string) {
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

func updateConfig(cmName, data string) {
	newClusterCM := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: constant.DLNamespace,
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
