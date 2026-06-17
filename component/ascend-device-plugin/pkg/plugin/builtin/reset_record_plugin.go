/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

package builtin

import (
	"context"
	"fmt"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/kubeclient"
	"Ascend-device-plugin/pkg/plugin"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

const (
	resetRecordPluginName  = "resetRecord"
	hotResetStartReason    = "HotResetStart"
	hotResetCompleteReason = "HotResetComplete"
	hotResetFailedReason   = "HotResetFailed"
	invalidFaultDevID      = -1
	invalidTokensLeft      = -1
)

type ResetRecordPlugin struct {
	plugin.HotResetPluginAdapter
	client   *kubeclient.ClientK8s
	nodeName string
}

func NewResetRecordPlugin(client *kubeclient.ClientK8s) *ResetRecordPlugin {
	nodeName, err := kubeclient.GetNodeNameFromEnv()
	if err != nil {
		hwlog.RunLog.Warnf("get node name failed: %v", err)
	}
	return &ResetRecordPlugin{client: client, nodeName: nodeName}
}

func (p *ResetRecordPlugin) Name() string {
	return resetRecordPluginName
}

func (p *ResetRecordPlugin) PreReset(ctx context.Context, deviceList []plugin.ResetDevice) {
	devIDs := formatDeviceList(deviceList)
	now := time.Now()
	event := &v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: api.KubeNS,
			Name:      fmt.Sprintf("%s.%d.reset.start", p.nodeName, now.UnixMilli()),
		},
		Type: v1.EventTypeWarning,
		Message: fmt.Sprintf("hot reset start, nodeName:%s, ringDevs:%s, faultDev:%d, tokensLeft:%d, time:%s",
			p.nodeName, devIDs, getFaultDevID(deviceList), getFaultTokensLeft(deviceList), now.Format(common.TimeFormat)),
		EventTime: metav1.MicroTime{Time: now},
		Reason:    hotResetStartReason,
		Action:    hotResetStartReason,
		Source:    v1.EventSource{Component: common.Component, Host: p.nodeName},
		InvolvedObject: v1.ObjectReference{
			Kind: "Node", Name: p.nodeName,
		},
		ReportingController: common.Component,
		ReportingInstance:   p.nodeName,
	}
	hwlog.RunLog.Infof("create hot reset start event to node: %s", p.nodeName)
	if _, err := p.client.CreateEvent(event); err != nil {
		hwlog.RunLog.Warnf("create hot reset start event failed: %v", err)
	}
}

func (p *ResetRecordPlugin) AfterReset(ctx context.Context, deviceList []plugin.ResetDevice,
	resetErr error) {
	devIDs := formatDeviceList(deviceList)
	now := time.Now()
	eventType := v1.EventTypeNormal
	reason := hotResetCompleteReason
	msg := fmt.Sprintf("hot reset complete, nodeName:%s, devices:%s, time:%s", p.nodeName, devIDs, now.Format(common.TimeFormat))
	if resetErr != nil {
		eventType = v1.EventTypeWarning
		reason = hotResetFailedReason
		msg = fmt.Sprintf("hot reset failed, nodeName:%s, devices:%s, time:%s, err:%v", p.nodeName, devIDs, now.Format(common.TimeFormat), resetErr)
	}
	event := &v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: api.KubeNS,
			Name:      fmt.Sprintf("%s.%d.reset.end", p.nodeName, now.UnixMilli()),
		},
		Type:      eventType,
		Message:   msg,
		EventTime: metav1.MicroTime{Time: now},
		Reason:    reason,
		Action:    reason,
		Source:    v1.EventSource{Component: common.Component, Host: p.nodeName},
		InvolvedObject: v1.ObjectReference{
			Kind: "Node", Name: p.nodeName,
		},
		ReportingController: common.Component,
		ReportingInstance:   p.nodeName,
	}
	hwlog.RunLog.Infof("create hot reset end event to node: %s", p.nodeName)
	if _, err := p.client.CreateEvent(event); err != nil {
		hwlog.RunLog.Warnf("create hot reset end event failed: %v", err)
	}
}

func formatDeviceList(deviceList []plugin.ResetDevice) string {
	ids := make([]string, 0, len(deviceList))
	for _, dev := range deviceList {
		ids = append(ids, fmt.Sprintf("%d", dev.LogicID))
	}
	return strings.Join(ids, ",")
}

func getFaultDevID(deviceList []plugin.ResetDevice) int32 {
	for _, dev := range deviceList {
		if dev.IsFaultDev {
			return dev.LogicID
		}
	}
	return invalidFaultDevID
}

func getFaultTokensLeft(deviceList []plugin.ResetDevice) int32 {
	for _, dev := range deviceList {
		if dev.IsFaultDev {
			return dev.TokensLeft
		}
	}
	return invalidTokensLeft
}
