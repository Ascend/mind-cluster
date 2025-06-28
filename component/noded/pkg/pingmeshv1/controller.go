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
Package pingmeshv1 is using for checking hccs network
*/
package pingmeshv1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"k8s.io/api/core/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"nodeD/pkg/pingmeshv1/consts"
	"nodeD/pkg/pingmeshv1/executor"
	"nodeD/pkg/pingmeshv1/policygenerator"
	"nodeD/pkg/pingmeshv1/policygenerator/fullmesh"
	"nodeD/pkg/pingmeshv1/resulthandler"
	"nodeD/pkg/pingmeshv1/resulthandler/cmreporter"
	"nodeD/pkg/pingmeshv1/resulthandler/filewriter"
	"nodeD/pkg/pingmeshv1/types"
	"nodeD/pkg/pingmeshv1/watcher"
	"nodeD/pkg/pingmeshv1/watcher/configmap"
)

// Manager is the controller for pingmeshv1
type Manager struct {
	watcher       watcher.Interface
	executor      *executor.DevManager
	handler       resulthandler.Interface
	policyFactory policygenerator.Factory
	nodeName      string
	ipCmName      string
	current       *types.HccspingMeshPolicy
}

// NewManager create a new Manager
func NewManager(config *Config) *Manager {
	devExecutor, err := executor.New()
	if err != nil {
		hwlog.RunLog.Errorf("new device manager failed, err: %v", err)
		return nil
	}
	c := &Manager{
		executor: devExecutor,
		ipCmName: consts.IpConfigmapNamePrefix + strconv.Itoa(int(devExecutor.SuperPodId)),
		nodeName: config.KubeClient.NodeName,
		current:  &types.HccspingMeshPolicy{},
	}
	c.policyFactory = policygenerator.NewFactory().Register(fullmesh.Rule, fullmesh.New(c.nodeName))
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
		Path:   consts.ResultRootDir + "/" + c.nodeName + consts.SuffixOfPingMeshLogFile,
		MaxAge: config.ResultMaxAge,
	})
	if fw != nil {
		handleFuncs = append(handleFuncs, fw.HandlePingMeshInfo)
	}
	reporter := cmreporter.New(&cmreporter.Config{
		Client:    config.KubeClient,
		Namespace: api.ClusterNS,
		Name:      consts.PingMeshFaultCmPrefix + c.nodeName,
		Labels: map[string]string{
			api.PubFaultCMLabelKey: consts.FaultConfigmapLabelValue,
		},
		NodeName: c.nodeName,
	})
	if reporter != nil {
		handleFuncs = append(handleFuncs, reporter.HandlePingMeshInfo)
	}
	c.handler = resulthandler.NewAggregatedHandler(handleFuncs...)
}

// Run start the pingmeshv1 controller
func (c *Manager) Run(ctx context.Context) {
	go c.watcher.Watch(ctx.Done())
	go c.handler.Handle(ctx.Done())
	go c.executor.Start(ctx.Done())
}

func (c *Manager) handleUserConfig(cm *v1.ConfigMap) {
	pmcfg, err := c.parsePingMeshConfig(cm.Data)
	if err != nil {
		hwlog.RunLog.Errorf("parse pingmeshv1 config failed, err: %v", err)
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
		c.current.DestAddr = c.policyFactory.Rule(fullmesh.Rule).Generate(c.current.Address)
		uid, err := generateJobUID(c.current.Config, c.current.Address)
		if err != nil {
			hwlog.RunLog.Errorf("generate job uid failed, err: %v", err)
			return
		}
		if c.current.UID == uid {
			return
		}
		c.current.UID = uid
		hwlog.RunLog.Infof("update config %v, uid: %s", c.current.Config, c.current.UID)
		c.executor.UpdateConfig(c.current.DeepCopy())
	}
}

func (c *Manager) parseSuperDeviceIDs(data map[string]string) (map[string]types.SuperDeviceIDs, error) {
	raw, ok := data[superPodCMKey]
	if !ok {
		return nil, fmt.Errorf("superPodCMKey not found")
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
	return config, nil
}
