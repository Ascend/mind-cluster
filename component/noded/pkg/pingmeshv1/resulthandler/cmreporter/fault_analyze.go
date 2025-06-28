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
Package cmreporter is using for pingmeshv1 result report to configmap
*/
package cmreporter

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager/common"
	"nodeD/pkg/kubeclient"
	"nodeD/pkg/pingmeshv1/types"
)

const (
	faultAssertionRecover = "recover"
	faultAssertionOccur   = "occur"
	publicFaultVersion    = "1.0"
	faultResource         = "pingmesh"
	faultCode             = "220001001"
)

type faultReporter struct {
	client          *kubeclient.ClientK8s
	namespace, name string
	lastFault       *api.PubFaultInfo
	labels          map[string]string
}

type state string

const (
	stateFault   state = "fault"
	stateHealthy state = "healthy"
	stateUnknown state = "unknown"
)

// New creat a new fault-reporter
func New(cfg *Config) *faultReporter {
	return &faultReporter{
		client:    cfg.Client,
		namespace: cfg.Namespace,
		name:      cfg.Name,
		labels:    cfg.Labels,
	}
}

// HandlePingMeshInfo handle ping-mesh result
func (f *faultReporter) HandlePingMeshInfo(res *types.HccspingMeshResult) error {
	hwlog.RunLog.Debugf("start to handle ping-mesh result, res:%v", res)
	lastFault, err := f.getLastFault()
	if err != nil {
		return err
	}
	lastFault = refreshFault(lastFault)

	cardStates := make(map[string]state, len(res.Results))
	for physicID, infos := range res.Results {
		cardStates[physicID] = checkFaultCard(infos)
	}

	fault, change := f.checkFault(lastFault, cardStates)
	f.lastFault = fault
	if !change {
		return nil
	}
	hwlog.RunLog.Infof("fault change, cur Fault:%v", fault)
	return f.reportFault()
}

func refreshFault(lastFault *api.PubFaultInfo) *api.PubFaultInfo {
	if lastFault == nil {
		return nil
	}
	faults := make([]api.Fault, 0)
	for _, fault := range lastFault.Faults {
		if fault.Assertion != faultAssertionRecover {
			faults = append(faults, fault)
		}
	}

	if len(faults) == 0 {
		return nil
	}

	lastFault.Faults = faults
	return lastFault
}

func (f *faultReporter) reportFault() error {
	if f.lastFault == nil {
		return nil
	}
	faultByte, err := json.Marshal(f.lastFault)
	if err != nil {
		return err
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      f.name,
			Namespace: f.namespace,
			Labels:    f.labels,
		},
		Data: map[string]string{
			api.PubFaultCMDataKey: string(faultByte),
		},
	}

	return f.client.CreateOrUpdateConfigMap(cm)
}

func (f *faultReporter) getLastFault() (*api.PubFaultInfo, error) {
	if f.lastFault != nil {
		return f.lastFault, nil
	}
	hwlog.RunLog.Info("try to get last fault from configmap")
	cm, err := f.client.GetConfigMap(f.name, api.ClusterNS)
	if errors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	faultInfo, ok := cm.Data[api.PubFaultCMDataKey]
	if !ok {
		hwlog.RunLog.Warnf("fault configmap not found key %s", api.PubFaultCMDataKey)
		return nil, nil
	}

	lastFault := &api.PubFaultInfo{}
	err = json.Unmarshal([]byte(faultInfo), lastFault)
	if err != nil {
		hwlog.RunLog.Warnf("unmarshal fault info failed, err:%v, will ignore last fault info", err)
		return nil, nil
	}
	return lastFault, nil
}

func checkFaultCard(infos map[uint]*common.HccspingMeshInfo) state {
	hasSuc := false
	hasDest := false
	for _, info := range infos {
		if info.DestNum == 0 {
			continue
		}
		hasDest = true
		for i := 0; i < info.DestNum; i++ {
			if info.ReplyStatNum[i] == 0 {
				return stateUnknown
			}
			if info.SucPktNum[i] != 0 {
				hasSuc = true
			}
		}
	}
	if !hasDest {
		return stateUnknown
	}
	if hasSuc {
		return stateHealthy
	}
	return stateFault
}

func (f *faultReporter) checkFault(last *api.PubFaultInfo, states map[string]state) (*api.PubFaultInfo, bool) {
	hwlog.RunLog.Debugf("checkFault, last: %v, cur states: %v", last, states)
	now := time.Now().UnixMilli()
	newFault := &api.PubFaultInfo{
		Version:   publicFaultVersion,
		Id:        string(uuid.NewUUID()),
		TimeStamp: now,
		Resource:  faultResource,
	}

	oldFaultMap := make(map[string]api.Fault, 0)
	if last != nil {
		for _, fault := range last.Faults {
			oldFaultMap[strconv.Itoa(int(fault.Influence[0].DeviceIds[0]))] = fault
		}
	}

	change := false
	for cardID, fault := range oldFaultMap {
		if st, ok := states[cardID]; ok && st == stateHealthy {
			fault.Assertion = faultAssertionRecover
			oldFaultMap[cardID] = fault
			change = true
		}
	}
	for cardID, st := range states {
		if st != stateFault {
			continue
		}
		if _, ok := oldFaultMap[cardID]; !ok {
			change = true
			oldFaultMap[cardID] = constructFaultInfo(cardID, now)
		}
	}

	newFault.Faults = make([]api.Fault, 0, len(oldFaultMap))
	for _, fault := range oldFaultMap {
		newFault.Faults = append(newFault.Faults, fault)
	}
	return newFault, change
}

func constructFaultInfo(cardID string, timestamp int64) api.Fault {
	nodeName := os.Getenv(api.NodeNameEnv)
	id, err := strconv.Atoi(cardID)
	if err != nil {
		hwlog.RunLog.Errorf("faultCardId %s is not a number", cardID)
		return api.Fault{}
	}
	return api.Fault{
		Assertion:     faultAssertionOccur,
		FaultId:       generateFaultID(nodeName, cardID),
		FaultType:     "NPU",
		FaultCode:     faultCode,
		FaultTime:     timestamp,
		FaultLocation: map[string]string{},
		Influence: []api.Influence{
			{
				NodeName:  nodeName,
				DeviceIds: []int32{int32(id)},
			},
		},
		Description: "hccsping-mesh fault",
	}
}

func generateFaultID(nodeName, cardId string) string {
	h := sha256.New()
	_, err := h.Write([]byte(nodeName + "/" + cardId))
	if err != nil {
		hwlog.RunLog.Warnf("generateFaultID failed, err: %v", err)
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}
