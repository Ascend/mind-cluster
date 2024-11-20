/* Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package pkg for noded
package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"

	"huawei.com/npu-exporter/v5/common-utils/hwlog"
)

const (
	nodeNameString     = "NODE_NAME"
	nodeInfoNamePrefix = "mindx-dl-nodeinfo-"
	nodeInfoNamespace  = "mindx-dl"
	nodeInfoLabelKey   = "mx-consumer-cim"
	nodeInfoLabelValue = "true"
	nodeInfoDataKey    = "NodeInfo"
)

type nodeInfoCM struct {
	NodeInfo  FaultDevInfo
	CheckCode string
}

// FaultDevInfo fault device info
type FaultDevInfo struct {
	HeartbeatTime     int64
	HeartbeatInterval int
}

// HeartbeatSender send heartbeat
type HeartbeatSender struct {
	k8sClient    *kubernetes.Clientset
	nodeName     string
	sendInterval int
}

// NewHeartbeatSender create HeartbeatSender
func newHeartbeatSender(nodeName string, interval int) (*HeartbeatSender, error) {
	k8sClient, err := newClientK8s()
	if err != nil {
		return nil, fmt.Errorf("failed to create kube client: %v", err)
	}
	return &HeartbeatSender{
		k8sClient:    k8sClient,
		nodeName:     nodeName,
		sendInterval: interval,
	}, nil
}

// SyncNodeHeartbeat sync heartbeat
func (hbs *HeartbeatSender) SyncNodeHeartbeat() {
	now := time.Now().Unix()
	nodeInfo := &FaultDevInfo{
		HeartbeatInterval: hbs.sendInterval,
		HeartbeatTime:     now,
	}
	info := nodeInfoCM{
		NodeInfo:  *nodeInfo,
		CheckCode: makeDataHash(nodeInfo),
	}
	nodeUpdateData, err := json.Marshal(info)
	if err != nil {
		hwlog.RunLog.Errorf("failed to marshal the node status data, error is %v", err)
		return
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nodeInfoNamePrefix + hbs.nodeName,
			Namespace: nodeInfoNamespace,
			Labels:    map[string]string{nodeInfoLabelKey: nodeInfoLabelValue},
		},
		Data: map[string]string{
			nodeInfoDataKey: string(nodeUpdateData),
		},
	}
	if err = hbs.createOrUpdateNodeInfoConfigMap(cm); err == nil {
		hwlog.RunLog.Infof("update node heartbeat success, interval: %d, time: %d", hbs.sendInterval, now)
	}
}

func (hbs *HeartbeatSender) createOrUpdateNodeInfoConfigMap(cm *corev1.ConfigMap) error {
	_, err := hbs.k8sClient.CoreV1().ConfigMaps(cm.Namespace).Update(context.TODO(), cm, metav1.UpdateOptions{})
	if err == nil {
		return nil
	}

	if !errors.IsNotFound(err) {
		hwlog.RunLog.Errorf("failed to update the node status data, error is %v", err)
		return err
	}
	hwlog.RunLog.Infof("node info configmap is not exist, try to create it")
	_, err = hbs.k8sClient.CoreV1().ConfigMaps(cm.Namespace).Create(context.TODO(), cm, metav1.CreateOptions{})
	if err != nil {
		hwlog.RunLog.Errorf("failed to create the node node info configmap, error is %v", err)
		return err
	}
	return nil
}

// SendHeartbeat start send heartbeat
func SendHeartbeat(interval int) error {
	nodeName := os.Getenv(nodeNameString)
	if err := checkNodeName(nodeName); err != nil {
		return fmt.Errorf("check node name failed: %v", err)
	}
	hbs, err := newHeartbeatSender(nodeName, interval)
	if err != nil {
		return fmt.Errorf("init heartbeat sender failed: %v", err)
	}
	wait.Until(hbs.SyncNodeHeartbeat, time.Duration(hbs.sendInterval)*time.Second, wait.NeverStop)
	return nil
}
