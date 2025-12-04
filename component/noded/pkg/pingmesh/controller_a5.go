/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
Package pingmesh is using for checking pingmesh network
*/
package pingmesh

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ascend-common/api/slownet"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager/common"
	"nodeD/pkg/pingmesh/consts"
	"nodeD/pkg/pingmesh/roceping"
	"nodeD/pkg/pingmesh/types"
)

const taskInterval = 10 * 60

// InitFileWatcher for init ping list file watcher in A3/A5
func (c *Manager) InitFileWatcher(ctx context.Context) {
	pingListFile, err := slownet.GetPingListFilePath(c.superPodId, c.serverIndex)
	if err != nil {
		hwlog.RunLog.Errorf("get ping list file path failed, err: %v", err)
		return
	}

	c.pingFileWatcher = roceping.NewFileWatcherLoop(ctx, taskInterval, consts.ResultRootDir)
	err = c.pingFileWatcher.AddListenPath(pingListFile)
	if err != nil {
		hwlog.RunLog.Errorf("add listen file path failed, err: %v", err)
		return
	}
}

// StartFileWatcher for start ping list file watcher in A3/A5
func (c *Manager) StartFileWatcher(stopCh <-chan struct{}) {
	if c.pingFileWatcher == nil {
		hwlog.RunLog.Info("ping file watcher is empty, can not watch ping list file")
		return
	}

	go c.pingFileWatcher.ListenEvents()

	for {
		select {
		case <-stopCh:
			hwlog.RunLog.Info("receive signal of stopCh, will stop ping file watcher event loop")
			return
		case event := <-c.pingFileWatcher.GetEventChan():
			hwlog.RunLog.Infof("receive file change event: %v", event)
			c.updateConfig()
		}
	}
}

// InitRoCEFileWatcher for init RoCE ping list file watcher in A5
func (c *Manager) InitRoCEFileWatcher(ctx context.Context) {
	if c.executor == nil || c.executor.GetDeviceType() != common.Ascend910A5 {
		hwlog.RunLog.Info("no need watch roce ping list file")
		return
	}
	if c.pingManager == nil {
		hwlog.RunLog.Info("roce ping manager is empty, no need watch roce ping list file")
		return
	}

	pingListFile, err := slownet.GetRoCEPingListFilePath(c.superPodId, c.serverIndex)
	if err != nil {
		hwlog.RunLog.Errorf("get ping list file path failed, err: %v", err)
		return
	}

	c.rocePingFileWatcher = roceping.NewFileWatcherLoop(ctx, taskInterval, consts.ResultRootDir)
	err = c.rocePingFileWatcher.AddListenPath(pingListFile)
	if err != nil {
		hwlog.RunLog.Errorf("add listen file path failed, err: %v", err)
		return
	}
}

// StartRoCEFileWatcher for start RoCE ping list file watching in A5
func (c *Manager) StartRoCEFileWatcher(stopCh <-chan struct{}) {
	if c.executor == nil || c.executor.GetDeviceType() != common.Ascend910A5 {
		return
	}

	if c.rocePingFileWatcher == nil {
		hwlog.RunLog.Info("roce ping file watcher is empty, can not watch roce ping list file")
		return
	}

	go c.rocePingFileWatcher.ListenEvents()
	for {
		select {
		case <-stopCh:
			hwlog.RunLog.Info("receive signal of stopCh, will stop roce ping file watcher event loop")
			return
		case event := <-c.rocePingFileWatcher.GetEventChan():
			hwlog.RunLog.Infof("receive file change event: %v", event)
			c.updateRoCEConfig()
		}
	}
}

func (c *Manager) updateRoCEConfig() {
	if c.pingManager == nil {
		hwlog.RunLog.Info("ping manager is empty")
		return
	}
	if c.currentRoCE.Config == nil {
		hwlog.RunLog.Info("current roce ping config is empty")
		return
	}
	if len(c.currentRoCE.Address) == 0 {
		hwlog.RunLog.Info("current roce devices is empty")
		return
	}
	if c.pingManager.GetDevType() != common.Ascend910A5 {
		hwlog.RunLog.Info("current is not A5, no need start ping manager")
		return
	}
	if c.currentRoCE.Config.Activate == types.ActivateOn {
		if !c.pingManager.CheckNodeLabelSupported() {
			return
		}
		if !c.pingManager.IsInPingListRange() {
			hwlog.RunLog.Infof("current node is not in roce ping range file, which superPodId=%s, serverIdx=%s",
				c.superPodId, c.serverIndex)
			return
		}
		hwlog.RunLog.Infof("current node is in roce ping range file, which superPodId=%s, serverIdx=%s",
			c.superPodId, c.serverIndex)
		err := c.generatePingPolicy()
		if err != nil {
			hwlog.RunLog.Errorf("generate roce policy failed, err: %v", err)
			return
		}
		hwlog.RunLog.Info("generate roce policy success")
		if err = c.updateRoCEPingListCheckFile(); err != nil {
			// only record the warning log
			hwlog.RunLog.Warnf("update ping list check info failed, err: %v", err)
		}
	}
	uid, err := generateJobUIDA5(c.currentRoCE.Config, c.currentRoCE.Address, c.getRoCEPingFileDataHash())
	if err != nil {
		hwlog.RunLog.Errorf("generate job uid failed, err: %v", err)
		return
	}
	if c.currentRoCE != nil && c.currentRoCE.UID == uid {
		hwlog.RunLog.Infof("roce ping policy no change, uid=%s", uid)
		return
	}
	c.currentRoCE.UID = uid
	hwlog.RunLog.Infof("update roce ping config %+v, uid: %s", *(c.currentRoCE.Config), c.currentRoCE.UID)
	c.pingManager.UpdateConfig(c.currentRoCE.DeepCopy())
}

func (c *Manager) getPingFileDataHash() string {
	if c.pingFileWatcher == nil {
		return ""
	}
	return c.pingFileWatcher.GetCurFileHash()
}

func (c *Manager) getRoCEPingFileDataHash() string {
	if c.rocePingFileWatcher == nil {
		return ""
	}
	return c.rocePingFileWatcher.GetCurFileHash()
}

func (c *Manager) updatePingListCheckFile() error {
	if c.pingFileWatcher == nil {
		return nil
	}
	return c.pingFileWatcher.UpdateCheckFile()
}

func (c *Manager) updateRoCEPingListCheckFile() error {
	if c.rocePingFileWatcher == nil {
		return nil
	}
	return c.rocePingFileWatcher.UpdateCheckFile()
}

func (c *Manager) generatePingPolicy() error {
	gen := c.policyFactory.Rule(roceping.Rule)
	retryTime := 0
	for ; retryTime < maxRetryTimes; retryTime++ {
		if c.currentRoCE.Config.Activate == types.ActivateOff {
			hwlog.RunLog.Infof("current roce config Activate is %s, no need retry", types.ActivateOff)
			break
		}
		dstAddrs := gen.Generate(c.currentRoCE.Address)
		if dstAddrs != nil {
			c.currentRoCE.DestAddr = dstAddrs
			c.currentRoCE.DestAddrMap = gen.GetDestAddrMap()
			hwlog.RunLog.Infof("generate roce ping policy success: %v", *c.currentRoCE)
			break
		}

		hwlog.RunLog.Infof("generate roce dstAddrs from policyFactory failed, will retry it, retry: %d", retryTime)
		time.Sleep(waitTimesForGenerate * time.Second)
	}

	if retryTime >= maxRetryTimes {
		hwlog.RunLog.Error("generate roce ping list info timeout")
		return errors.New("generate roce ping list info timeout")
	}
	return nil
}

func (c *Manager) parseRoCEPingConfig(data map[string]string) (*types.HccspingMeshConfig, error) {
	const rocePingCfgName = "roce"
	cfg, ok := data[rocePingCfgName]
	if !ok {
		hwlog.RunLog.Info("no config for roce ping, would not enable the roce ping task")
		return nil, errors.New("get activate config for roce failed")
	}
	var config = &types.HccspingMeshConfig{}
	err := json.Unmarshal([]byte(cfg), config)
	if err != nil {
		return nil, err
	}
	if config.Activate != types.ActivateOn && config.Activate != types.ActivateOff {
		return nil, errors.New("invalid activate value for roce")
	}

	if config.TaskInterval < common.MinTaskInterval || config.TaskInterval > common.MaxTaskInterval {
		return nil, fmt.Errorf("roce task interval %d is invalid, should be betwedeen %d and %d", config.TaskInterval,
			common.MinTaskInterval, common.MaxTaskInterval)
	}
	return config, nil
}

func generateJobUIDA5(config *types.HccspingMeshConfig, destAddrs map[string]types.SuperDeviceIDs,
	pingFileHash string) (string, error) {
	cfg, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	address, err := json.Marshal(destAddrs)
	if err != nil {
		return "", err
	}
	address = append(address, cfg...)
	address = append(address, []byte(pingFileHash)...)
	hasher := sha256.New()
	_, err = hasher.Write(address)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
