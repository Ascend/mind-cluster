// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package pingmesh a series of function handle ping mesh configmap create/update/delete
package pingmesh

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"

	"ascend-common/api"
	"ascend-common/api/slownet"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"clusterd/pkg/application/fdapi"
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
	defaultPerm           = 0644
	jsonIndentCnt         = 4
)

var pingMeshLabel = map[string]string{"app": "pingmesh"}
var jsonIndent = strings.Repeat(" ", jsonIndentCnt)

type publishLog struct {
	publishType  string
	publishKey   string
	preCheckCode string
}

type publishManager struct {
	inited            atomic.Bool
	cmPublishLogMap   map[string]*publishLog
	filePublishLogMap map[string]*publishLog
	eventMap          map[string]string
	rwLock            sync.RWMutex
}

var publishMgr *publishManager

// rasNetDetectInst the switch for net fault detect feature in ras
var rasNetDetectInst = RasNetFaultCmManager{}
var rasConfig = constant.CathelperConf{}

func init() {
	publishMgr = &publishManager{
		cmPublishLogMap:   make(map[string]*publishLog),
		filePublishLogMap: make(map[string]*publishLog),
		eventMap:          make(map[string]string, initSuperPodNum),
		rwLock:            sync.RWMutex{},
	}
	publishMgr.inited.Store(false)
	rasNetDetectInst = RasNetFaultCmManager{
		RWMutex: sync.RWMutex{},
		NetInfo: constant.NetFaultInfo{NetFault: constant.RasNetDetectOff},
	}
	rasConfig = NewCathelperConf()
}

// RasNetFaultCmManager ras feature net fault detect configmap manager info
type RasNetFaultCmManager struct {
	sync.RWMutex
	NetInfo constant.NetFaultInfo
}

// UpdateNetInfo update netinfo value
func (r *RasNetFaultCmManager) UpdateNetInfo(info *constant.NetFaultInfo) {
	if info == nil {
		return
	}
	r.Lock()
	r.NetInfo.NetFault = info.NetFault
	r.Unlock()
}

// Update update net fault info
func (r *RasNetFaultCmManager) Update(info *constant.NetFaultInfo) {
	if info == nil {
		return
	}
	r.UpdateNetInfo(info)
}

// CheckIsOn check net fault detect feature is active status
func (r *RasNetFaultCmManager) CheckIsOn() bool {
	r.RLock()
	defer r.RUnlock()
	return r.NetInfo.NetFault == constant.RasNetDetectOn
}

func updateSuperPodDeviceFile(device *api.SuperPodDevice, checkCode string, init bool) error {
	if device == nil {
		hwlog.RunLog.Warnf("nil device")
		return nil
	}
	b, err := json.MarshalIndent(device, "", jsonIndent)
	if err != nil || len(b) == 0 {
		hwlog.RunLog.Warnf("marshal bytes illegal, SuperPodID=%s, init=%v, err=%v",
			device.SuperPodID, init, err)
		return nil
	}
	if errWrite := writeJsonDataByteToFile(device.SuperPodID, b); errWrite != nil {
		return errWrite
	}
	return nil
}

func writeJsonDataByteToFile(superPodID string, data []byte) error {
	filePath, err := slownet.GetSuperPodInfoFilePath(superPodID, publishCmNamePrefix)
	if err != nil {
		hwlog.RunLog.Errorf("get super pod info file path failed, err: %v", err)
		return err
	}

	fileParentDir := filepath.Dir(filePath)
	if !utils.IsLexist(fileParentDir) {
		if mkErr := utils.MakeSureDir(filePath); mkErr != nil {
			hwlog.RunLog.Infof("create file path %s failed, err: %v", fileParentDir, mkErr)
			return mkErr
		}
		hwlog.RunLog.Infof("create the file path %s success", fileParentDir)
	}

	if !utils.IsLexist(filePath) {
		hwlog.RunLog.Infof("file %s is not exist, will create it", filePath)
	}

	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, defaultPerm)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			hwlog.RunLog.Warnf("close file %s failed, err: %v", filePath, err)
			return
		}
	}()

	dataJsonStr := string(data)
	if _, err = f.WriteString(dataJsonStr); err != nil {
		return err
	}

	if err = os.Chmod(filePath, defaultPerm); err != nil {
		return err
	}
	hwlog.RunLog.Infof("write file %s success", filePath)
	return nil
}

func deleteSuperPodFile(superPodID string) error {
	filePath, err := slownet.GetSuperPodInfoFilePath(superPodID, publishCmNamePrefix)
	if err != nil {
		hwlog.RunLog.Errorf("get super pod info file path failed, err: %v", err)
		return err
	}
	if !utils.IsLexist(filePath) {
		return nil
	}
	if rmErr := os.Remove(filePath); rmErr != nil {
		hwlog.RunLog.Errorf("remove file %s failed, err: %v", filePath, rmErr)
		return rmErr
	}
	hwlog.RunLog.Infof("remove file %s success", filePath)
	return nil
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
	publishMgr.inited.Store(true)
	failedTasks := make([]task, 0)
	for _, superPodDevice := range superpod.ListClusterDevice() {
		if superPodDevice == nil || superPodDevice.SuperPodID == "" {
			return
		}
		checkCode := util.MakeDataHash(superPodDevice)
		err := handleCmUpdate(superPodDevice.SuperPodID, superPodDevice, checkCode, true)
		if err != nil {
			failedTasks = append(failedTasks, task{
				superPodID: superPodDevice.SuperPodID,
				operator:   constant.AddOperator,
			})
			hwlog.RunLog.Errorf("init super pod info cm error, %v", err)
		}
		err = handleFileUpdate(superPodDevice.SuperPodID, superPodDevice, checkCode, true)
		if err != nil {
			failedTasks = append(failedTasks, task{
				superPodID: superPodDevice.SuperPodID,
				operator:   constant.AddOperator,
			})
			hwlog.RunLog.Errorf("init super pod file error, %v", err)
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

func handleCmUpdate(superPodID string, device *api.SuperPodDevice, checkCode string, init bool) error {
	log, exist := publishMgr.cmPublishLogMap[superPodID]
	if exist && log.preCheckCode == checkCode {
		hwlog.RunLog.Debugf("super pod device checkCode not change, superPodID=%s", checkCode)
		return nil
	}
	err := updateSuperPodDeviceCM(device, checkCode, init)
	if err != nil {
		hwlog.RunLog.Errorf("update super pod device cm failed, err=%v, superPodID=%s", err, superPodID)
		return err
	}
	hwlog.RunLog.Infof("update super pod device cm success, superPodID=%s", superPodID)
	publishMgr.cmPublishLogMap[superPodID] = &publishLog{
		publishKey:   superPodID,
		preCheckCode: checkCode,
	}
	return nil
}

func handleFileUpdate(superPodID string, device *api.SuperPodDevice, checkCode string, init bool) error {
	log, exist := publishMgr.filePublishLogMap[superPodID]
	if exist && log.preCheckCode == checkCode {
		hwlog.RunLog.Debugf("super pod device checkCode not change, superPodID=%s", checkCode)
		return nil
	}
	err := updateSuperPodDeviceFile(device, checkCode, init)
	if err != nil {
		hwlog.RunLog.Errorf("update super pod device file failed, err=%v, superPodID=%s", err, superPodID)
		return err
	}
	err = saveConfigToFile(superPodID, &rasConfig)
	if err != nil {
		hwlog.RunLog.Errorf("save config to file failed, err=%v, superPodID=%s", err, superPodID)
		return err
	}
	publishMgr.filePublishLogMap[superPodID] = &publishLog{
		publishKey:   superPodID,
		preCheckCode: checkCode,
	}
	hwlog.RunLog.Infof("update super pod device file success, superPodID=%s", superPodID)
	fdapi.ReloadController()
	return nil
}

func handleUpdate(superPodID string, device *api.SuperPodDevice) error {
	if device == nil || superPodID == "" {
		hwlog.RunLog.Warnf("nil super pod device or superPodID, ignore it. superPodID=%s", superPodID)
		return nil
	}
	checkCode := util.MakeDataHash(device)
	err := handleCmUpdate(superPodID, device, checkCode, false)
	if err != nil {
		return err
	}
	return handleFileUpdate(superPodID, device, checkCode, false)
}

func handleCmDelete(superPodID string) error {
	cmName := fmt.Sprintf("%s-%s", publishCmNamePrefix, superPodID)
	err := kube.DeleteConfigMap(cmName, api.ClusterNS)
	if err == nil || errors.IsNotFound(err) {
		hwlog.RunLog.Infof("delete super pod device cm success, superPodID=%s", superPodID)
		if rasNetDetectInst.CheckIsOn() {
			hwlog.RunLog.Infof("super-pod-%s file is deleted and controller will be reloaded", superPodID)
			fdapi.ReloadController()
		}
		delete(publishMgr.cmPublishLogMap, superPodID)
		return nil
	}
	hwlog.RunLog.Errorf("delete super pod device cm failed, err=%v, superPodID=%s", err, superPodID)
	return fmt.Errorf("delete superPod cm failed, cmName=%s, err=%v", cmName, err)
}

func handleFileDelete(superPodID string) error {
	err := deleteSuperPodFile(superPodID)
	if err == nil || os.IsNotExist(err) {
		hwlog.RunLog.Infof("delete super pod device file success, superPodID=%s",
			superPodID)
		delete(publishMgr.filePublishLogMap, superPodID)
		return nil
	}
	hwlog.RunLog.Errorf("delete super pod file failed, err: %v", err)
	return err
}

func handleDelete(superPodID string) error {
	err := handleCmDelete(superPodID)
	if err != nil {
		return err
	}
	return handleFileDelete(superPodID)
}

type task struct {
	superPodID string
	operator   string
}

func getPartTaskAndClean() []task {
	publishMgr.rwLock.Lock()
	defer publishMgr.rwLock.Unlock()
	n := 0
	tasks := make([]task, 0, handleBatch+handleBatch)
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
	if rasNetDetectInst.CheckIsOn() {
		initSuperPodsCM()
	}
	ticker := time.NewTicker(eventCheckPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if !rasNetDetectInst.CheckIsOn() {
				hwlog.RunLog.Debug("ras feature net fault detect is inactive")
				continue
			}
			if !publishMgr.inited.Load() {
				hwlog.RunLog.Info("cluster super pod device is not initialized, init it")
				initSuperPodsCM()
				continue
			}
			tasks := getPartTaskAndClean()
			hwlog.RunLog.Debugf("event-length=%d, handleBatch=%d",
				len(publishMgr.eventMap), len(tasks))
			handleTasks(tasks)
		case <-ctx.Done():
			return
		}
	}
}
