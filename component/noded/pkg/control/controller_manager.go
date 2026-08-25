/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package control for fault handling
package control

import (
	"time"

	"k8s.io/apimachinery/pkg/util/rand"

	"ascend-common/api/annotation"
	"ascend-common/api/label"
	"ascend-common/common-utils/hwlog"
	"nodeD/pkg/common"
	"nodeD/pkg/common/manager"
	"nodeD/pkg/control/nodesn"
	"nodeD/pkg/kubeclient"
	"nodeD/pkg/processmanager"
)

const randSecond = 20

// ControllerManager controller manager
type ControllerManager struct {
	kubeClient         *kubeclient.ClientK8s
	configManager      manager.ConfigManager
	nextFaultProcessor common.FaultProcessor
}

// NewControlManager create a control manager
func NewControlManager(client *kubeclient.ClientK8s) *ControllerManager {
	return &ControllerManager{
		kubeClient:    client,
		configManager: manager.NewConfigManager(),
	}
}

// Init initialize node controller
func (nc *ControllerManager) Init() error {
	return nil
}

// Execute process fault device info and send message to next fault processor
func (nc *ControllerManager) Execute(fcInfo *common.FaultAndConfigInfo, processType string) {
	controls := processmanager.GetControlPlugins(processType)
	for _, plugin := range controls {
		fcInfo = plugin.Control(fcInfo)
	}
	nc.nextFaultProcessor.Execute(fcInfo, processType)
}

// SetNextFaultProcessor set the next fault processor
func (nc *ControllerManager) SetNextFaultProcessor(faultProcessor common.FaultProcessor) {
	nc.nextFaultProcessor = faultProcessor
}

// InitNodeSNInfo init node sn
func (nc *ControllerManager) InitNodeSNInfo() error {
	rand.Seed(time.Now().UnixNano())
	randomSecond := time.Duration(rand.Intn(randSecond)) * time.Second
	time.Sleep(randomSecond)
	nodeSN, err := nodesn.GetNodeSN()
	if err != nil {
		hwlog.RunLog.Errorf("get node SN failed, err is %v", err)
		return err
	}
	hwlog.RunLog.Infof("get node SN success, add SN(%s) to node annotation and label", nodeSN)
	err = nc.kubeClient.AddLabel(label.NPUServerSerialNumberLabel, nodeSN)
	if err != nil {
		hwlog.RunLog.Warnf("add server.serial-number label failed, err is %v", err)
	}
	// Also write product-serial-number annotation
	err = nc.kubeClient.AddAnnotation(annotation.NodeSNAnnotationDeprecated, nodeSN)
	if err != nil {
		hwlog.RunLog.Errorf("add node annotation failed, err is %v", err)
		return err
	}
	return nil
}
