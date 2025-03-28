// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package pingmesh a series of function handle ping mesh configmap create/update/delete
package pingmesh

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/superpod"
	"clusterd/pkg/interface/kube"
)

const (
	publishInterval       = 10 * time.Millisecond
	handleBatch           = 5
	initSuperPodNum       = 128
	publishCmNamePrefix   = "super-pod"
	superPodDeviceInfoKey = "superPodDevice"
	eventCheckPeriod      = 5 * time.Second
)

var pingMeshLabel = map[string]string{"app": "pingmesh"}

type publishLog struct {
	publishKey   string
	preCheckCode string
}

type publishManager struct {
	publishLogMap map[string]*publishLog
	eventMap      map[string]string
	rwLock        sync.RWMutex
}

var publishMgr *publishManager

func init() {
	publishMgr = &publishManager{
		publishLogMap: make(map[string]*publishLog),
		eventMap:      make(map[string]string, initSuperPodNum),
		rwLock:        sync.RWMutex{},
	}
}

func updateSuperPodDeviceCM(device *api.SuperPodDevice, checkCode string, init bool) error {
	if device == nil {
		hwlog.RunLog.Warnf("nil device")
		return nil
	}
	b, err := json.Marshal(device)
	if err != nil || len(b) == 0 {
		hwlog.RunLog.Warnf("marshal bytes illegal, SuperPodID=%s, init=%v, err=%v",
			device.SuperPodID, init, err)
		return nil
	}
	cmName := fmt.Sprintf("%s-%s", publishCmNamePrefix, device.SuperPodID)
	data := map[string]string{superPodDeviceInfoKey: string(b)}
	if init {
		return kube.CreateOrUpdateConfigMap(cmName, api.ClusterNS, data, pingMeshLabel)
	}
	return kube.UpdateOrCreateConfigMap(cmName, api.ClusterNS, data, pingMeshLabel)
}

func addEvent(superPodID, operator string) {
	publishMgr.rwLock.Lock()
	publishMgr.eventMap[superPodID] = operator
	publishMgr.rwLock.Unlock()
}

func initSuperPodsCM() {
	failedTasks := make([]task, 0)
	for _, superPodDevice := range superpod.ListClusterDevice() {
		if superPodDevice == nil || superPodDevice.SuperPodID == "" {
			return
		}
		checkCode := util.MakeDataHash(superPodDevice)
		err := updateSuperPodDeviceCM(superPodDevice, checkCode, true)
		if err != nil {
			hwlog.RunLog.Debugf("init cm error, superPodID=%s, err=%v",
				superPodDevice.SuperPodID, err)
			failedTasks = append(failedTasks, task{
				superPodID: superPodDevice.SuperPodID,
				operator:   constant.AddOperator,
			})
		} else {
			publishMgr.publishLogMap[superPodDevice.SuperPodID] = &publishLog{
				publishKey:   superPodDevice.SuperPodID,
				preCheckCode: util.MakeDataHash(superPodDevice),
			}
		}
		time.Sleep(publishInterval)
	}
	publishMgr.rwLock.Lock()
	defer publishMgr.rwLock.Unlock()
	for _, failedTask := range failedTasks {
		if _, ok := publishMgr.eventMap[failedTask.superPodID]; !ok {
			publishMgr.eventMap[failedTask.superPodID] = failedTask.operator
		}
	}
}

func handleUpdate(superPodID string, device *api.SuperPodDevice) error {
	if device == nil || superPodID == "" {
		hwlog.RunLog.Warnf("nil super pod device or superPodID, ignore it. superPodID=%s", superPodID)
		return nil
	}
	checkCode := util.MakeDataHash(device)
	log, exist := publishMgr.publishLogMap[superPodID]
	if exist && log.preCheckCode == checkCode {
		hwlog.RunLog.Debugf("super pod device checkCode not change, superPodID=%s", checkCode)
		return nil
	}
	err := updateSuperPodDeviceCM(device, checkCode, false)
	if err != nil {
		hwlog.RunLog.Errorf("update super pod device cm failed, err=%v, superPodID=%s", err, superPodID)
		return err
	}
	hwlog.RunLog.Infof("update super pod device cm success, superPodID=%s", superPodID)
	publishMgr.publishLogMap[superPodID] = &publishLog{
		publishKey:   superPodID,
		preCheckCode: checkCode,
	}
	return nil
}

func handleDelete(superPodID string) error {
	cmName := fmt.Sprintf("%s-%s", publishCmNamePrefix, superPodID)
	err := kube.DeleteConfigMap(cmName, api.ClusterNS)
	if err == nil || errors.IsNotFound(err) {
		hwlog.RunLog.Infof("delete super pod device cm success, superPodID=%s", superPodID)
		delete(publishMgr.publishLogMap, superPodID)
		return nil
	}
	hwlog.RunLog.Errorf("delete super pod device cm failed, err=%v, superPodID=%s", err, superPodID)
	return fmt.Errorf("delete superPod cm failed, cmName=%s, err=%v", cmName, err)
}

type task struct {
	superPodID string
	operator   string
}

func getPartTaskAndClean() []task {
	publishMgr.rwLock.Lock()
	defer publishMgr.rwLock.Unlock()
	n := 0
	tasks := make([]task, 0, handleBatch)
	for superPodID, operator := range publishMgr.eventMap {
		n++
		if n > handleBatch {
			break
		}
		tasks = append(tasks, task{
			superPodID: superPodID,
			operator:   operator,
		})
		delete(publishMgr.eventMap, superPodID)
	}
	return tasks
}

func handleTasks(tasks []task) {
	failedTasks := make([]task, 0)
	var err error
	for _, t := range tasks {
		switch t.operator {
		case constant.AddOperator, constant.UpdateOperator:
			superPodDevice := superpod.GetSuperPodDevice(t.superPodID)
			err = handleUpdate(t.superPodID, superPodDevice)
		case constant.DeleteOperator:
			err = handleDelete(t.superPodID)
		default:
			hwlog.RunLog.Errorf("error operator: %s, superPodID=%s",
				t.operator, t.superPodID)
		}
		if err != nil {
			failedTasks = append(failedTasks, t)
		}
		time.Sleep(publishInterval)
	}
	publishMgr.rwLock.Lock()
	defer publishMgr.rwLock.Unlock()
	for _, failedTask := range failedTasks {
		if _, ok := publishMgr.eventMap[failedTask.superPodID]; !ok {
			publishMgr.eventMap[failedTask.superPodID] = failedTask.operator
		}
	}
}

// TickerCheckSuperPodDevice ticker check super pod device modify event
func TickerCheckSuperPodDevice(ctx context.Context) {
	initSuperPodsCM()
	ticker := time.NewTicker(eventCheckPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			tasks := getPartTaskAndClean()
			hwlog.RunLog.Debugf("event length=%d, handleBatch=%d",
				len(publishMgr.eventMap), len(tasks))
			handleTasks(tasks)
		case <-ctx.Done():
			return
		}
	}
}
