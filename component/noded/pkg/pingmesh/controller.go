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
Package pingmesh is using for checking hccs network
*/
package pingmesh

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	v1 "k8s.io/api/core/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager/common"
	"nodeD/pkg/pingmesh/consts"
	"nodeD/pkg/pingmesh/executor"
	"nodeD/pkg/pingmesh/policygenerator"
	"nodeD/pkg/pingmesh/policygenerator/fullmesh"
	"nodeD/pkg/pingmesh/resulthandler"
	"nodeD/pkg/pingmesh/resulthandler/filewriter"
	"nodeD/pkg/pingmesh/types"
	"nodeD/pkg/pingmesh/watcher"
	"nodeD/pkg/pingmesh/watcher/configmap"
)

const (
	maxRetryTimes        = 24
	waitTimesForGenerate = 5
)

// Manager is the controller for pingmesh
type Manager struct {
	watcher       watcher.Interface
	executor      *executor.DevManager
	handler       resulthandler.Interface
	policyFactory policygenerator.Factory
	superPodId    string
	serverIndex   string
	nodeName      string
	ipCmName      string
	current       *types.HccspingMeshPolicy
}

// NewManager create a new Manager
func NewManager(config *Config) *Manager {
	if config == nil || config.KubeClient == nil {
		hwlog.RunLog.Error("pingmesh config or kubeclient is nil")
		return nil
	}
	devExecutor, err := executor.New()
	if err != nil {
		hwlog.RunLog.Errorf("new device manager failed, err: %v", err)
		return nil
	}
	c := &Manager{
		executor:    devExecutor,
		ipCmName:    consts.IpConfigmapNamePrefix + strconv.Itoa(int(devExecutor.SuperPodId)),
		nodeName:    config.KubeClient.NodeName,
		current:     &types.HccspingMeshPolicy{},
		superPodId:  strconv.Itoa(int(devExecutor.SuperPodId)),
		serverIndex: strconv.Itoa(int(devExecutor.ServerIndex)),
	}

	gen := fullmesh.New(c.nodeName, c.superPodId, c.serverIndex)
	c.policyFactory = policygenerator.NewFactory().Register(fullmesh.Rule, gen)
	c.initWatcher(config)
	c.initHandler(config)
	c.executor.SetResultHandler(c.handler.Receive)
	return c
}

func (c *Manager) initWatcher(config *Config) {
	var opts []configmap.Option
	opts = append(opts, configmap.WithNamespace(api.ClusterNS))
	opts = append(opts, configmap.WithLabelSector(fmt.Sprintf("%s=%s", consts.PingMeshConfigLabelKey,
		consts.PingMeshConfigLabelValue)))
	opts = append(opts, configmap.WithNamedHandlers(
		configmap.NamedHandler{Name: c.ipCmName, Handle: c.handleClusterAddress},
		configmap.NamedHandler{Name: consts.PingMeshConfigCm, Handle: c.handleUserConfig},
	))
	w := configmap.NewWatcher(config.KubeClient, opts...)
	w.Init()
	c.watcher = w
}

func (c *Manager) initHandler(config *Config) {
	var handleFuncs []resulthandler.HandleFunc
	fw := filewriter.New(&filewriter.Config{
		Path:        consts.ResultRootDir + "/" + c.nodeName + consts.SuffixOfPingMeshLogFile,
		MaxAge:      config.ResultMaxAge,
		SuperPodId:  c.superPodId,
		ServerIndex: c.serverIndex,
	})
	if fw != nil {
		handleFuncs = append(handleFuncs, fw.HandlePingMeshInfo)
	}
	c.handler = resulthandler.NewAggregatedHandler(handleFuncs...)
}

// Run start the pingmesh controller
func (c *Manager) Run(ctx context.Context) {
	if c == nil || ctx == nil {
		hwlog.RunLog.Error("pingmesh manager or context is nil")
		return
	}
	go c.watcher.Watch(ctx.Done())
	go c.handler.Handle(ctx.Done())
	go c.executor.Start(ctx.Done())
}

func (c *Manager) handleUserConfig(cm *v1.ConfigMap) {
	pmcfg, err := c.parsePingMeshConfig(cm.Data)
	if err != nil {
		hwlog.RunLog.Errorf("parse pingmesh config failed, err: %v", err)
		return
	}
	if pmcfg.Activate != types.ActivateOn && pmcfg.Activate != types.ActivateOff {
		hwlog.RunLog.Errorf("invalid activate value: %s", pmcfg.Activate)
		return
	}
	hwlog.RunLog.Infof("activate: %s", pmcfg.Activate)
	c.current.Config = pmcfg
	c.updateConfig()
}

func (c *Manager) handleClusterAddress(cm *v1.ConfigMap) {
	sdids, err := c.parseSuperDeviceIDs(cm.Data)
	if err != nil {
		hwlog.RunLog.Errorf("parse superDeviceIDs failed, err: %v", err)
		return
	}
	c.current.Address = sdids
	c.updateConfig()
}

func (c *Manager) updateConfig() {
	hwlog.RunLog.Infof("has dest: %v, has config %v", len(c.current.Address) != 0, c.current.Config != nil)
	if len(c.current.Address) != 0 && c.current.Config != nil {
		gen := c.policyFactory.Rule(fullmesh.Rule)
		retryTime := 0
		for ; retryTime < maxRetryTimes; retryTime++ {
			if c.current.Config.Activate == types.ActivateOff {
				hwlog.RunLog.Infof("current config Activate is %s, no need retry", types.ActivateOff)
				break
			}
			dstAddrs := gen.Generate(c.current.Address)
			if len(dstAddrs) > 0 {
				c.current.DestAddr = dstAddrs
				c.current.DestAddrMap = gen.GetDestAddrMap()
				hwlog.RunLog.Infof("generate dstAddrs from policyFactory success")
				break
			}
			if c.current.Config.Activate == types.ActivateOff {
				hwlog.RunLog.Infof("current config Activate is %s, no need retry", types.ActivateOff)
				break
			}
			hwlog.RunLog.Infof("generate dstAddrs from policyFactory failed, will retry it")
			time.Sleep(waitTimesForGenerate * time.Second)
		}
		if retryTime >= maxRetryTimes {
			hwlog.RunLog.Errorf("generate ping list info failed")
			return
		}
		uid, err := generateJobUID(c.current.Config, c.current.Address)
		if err != nil {
			hwlog.RunLog.Errorf("generate job uid failed, err: %v", err)
			return
		}
		if c.current.UID == uid {
			return
		}
		c.current.UID = uid
		hwlog.RunLog.Infof("update config %+v, uid: %s", *(c.current.Config), c.current.UID)
		c.executor.UpdateConfig(c.current.DeepCopy())
	}
}

func (c *Manager) parseSuperDeviceIDs(data map[string]string) (map[string]types.SuperDeviceIDs, error) {
	raw, ok := data[superPodCMKey]
	if !ok {
		return nil, errors.New("superPodCMKey not found")
	}
	superPodDevice := &api.SuperPodDevice{}
	if err := json.Unmarshal([]byte(raw), superPodDevice); err != nil {
		return nil, fmt.Errorf("unmarshal superPodDevice failed, err: %v", err)
	}
	if strconv.Itoa(int(c.executor.SuperPodId)) != superPodDevice.SuperPodID {
		return nil, fmt.Errorf("superPodId not match, expect: %s, actual: %s",
			strconv.Itoa(int(c.executor.SuperPodId)), superPodDevice.SuperPodID)
	}
	if _, ok = superPodDevice.NodeDeviceMap[c.nodeName]; !ok {
		return nil, fmt.Errorf("node %s not found in superPodDevice", c.nodeName)
	}

	nodes := make(map[string]types.SuperDeviceIDs, len(superPodDevice.NodeDeviceMap))
	for node, devices := range superPodDevice.NodeDeviceMap {
		nodes[node] = devices.DeviceMap
	}
	return nodes, nil
}

func (c *Manager) parsePingMeshConfig(data map[string]string) (*types.HccspingMeshConfig, error) {
	cfg, ok := data[strconv.Itoa(int(c.executor.SuperPodId))]
	if !ok {
		hwlog.RunLog.Infof("no config for superPodId %d, try to get global config", c.executor.SuperPodId)
		cfg, ok = data[globalConfigKey]
		if !ok {
			return nil, errors.New("get activate config failed")
		}
	}
	var config = &types.HccspingMeshConfig{}
	err := json.Unmarshal([]byte(cfg), config)
	if err != nil {
		return nil, err
	}
	if config.TaskInterval < common.MinTaskInterval || config.TaskInterval > common.MaxTaskInterval {
		return nil, fmt.Errorf("task interval %d is invalid, should be between %d and %d", config.TaskInterval,
			common.MinTaskInterval, common.MaxTaskInterval)
	}
	return config, nil
}
