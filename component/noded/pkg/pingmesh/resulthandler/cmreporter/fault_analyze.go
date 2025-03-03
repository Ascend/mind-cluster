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
Package cmreporter is using for pingmesh result report to configmap
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
	nodecommon "nodeD/pkg/common"
	"nodeD/pkg/kubeclient"
	"nodeD/pkg/pingmesh/consts"
	"nodeD/pkg/pingmesh/types"
)

const (
	faultAssertionRecover = "recover"
	faultAssertionOccur   = "occur"
	publicFaultVersion    = "1.0"
	faultResource         = "ping-mesh"
	faultConfigmapKey     = "publicFault"
	faultCode             = "010001001"
)

type faultReporter struct {
	client          *kubeclient.ClientK8s
	namespace, name string
	lastFault       *api.PubFaultInfo
	labels          map[string]string
}

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
	curFault := f.calCurrentFault(res)
	fault, change := f.checkFault(lastFault, curFault)
	if !change {
		return nil
	}
	hwlog.RunLog.Infof("fault change, lastFault:%v, curFault:%v", lastFault, curFault)
	f.lastFault = fault
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
			faultConfigmapKey: string(faultByte),
		},
	}

	return f.client.CreateOrUpdateConfigMap(cm)
}

func (f *faultReporter) getLastFault() (*api.PubFaultInfo, error) {
	if f.lastFault != nil {
		return f.lastFault, nil
	}
	cm, err := f.client.GetConfigMap(f.name, consts.ConfigmapNamespace)
	if errors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	lastFault := &api.PubFaultInfo{}
	err = json.Unmarshal([]byte(cm.Data[faultConfigmapKey]), lastFault)
	if err != nil {
		return nil, err
	}
	return lastFault, nil
}

type faultCard struct {
	physicID string
}

func (f *faultReporter) filterFaultCards(res *types.HccspingMeshResult) []*faultCard {
	faultCardIds := make([]*faultCard, 0)
	for physicID, infos := range res.Results {
		if !isFaultCard(infos) {
			continue
		}
		faultCardIds = append(faultCardIds, &faultCard{physicID: physicID})
	}
	return faultCardIds
}

func isFaultCard(infos map[uint]*common.HccspingMeshInfo) bool {
	hasSuc := false
	hasReply := false
	for _, info := range infos {
		for i := 0; i < info.DestNum; i++ {
			if info.ReplyStatNum[i] == 0 {
				continue
			}
			hasReply = true
			if info.SucPktNum[i] != 0 {
				hasSuc = true
				break
			}
		}
		if hasSuc {
			break
		}
	}
	return hasReply && !hasSuc
}

func (f *faultReporter) constructFaultInfo(faultCards []*faultCard) *api.PubFaultInfo {
	pf := &api.PubFaultInfo{
		Version:   publicFaultVersion,
		Id:        string(uuid.NewUUID()),
		TimeStamp: time.Now().Unix(),
		Resource:  faultResource,
		Faults:    make([]api.Fault, 0, len(faultCards)),
	}

	for _, fc := range faultCards {
		id, err := strconv.Atoi(fc.physicID)
		if err != nil {
			hwlog.RunLog.Errorf("faultCardId %s is not a number", fc.physicID)
			continue
		}
		now := time.Now()
		nodeName := os.Getenv(nodecommon.ENVNodeNameKey)
		pf.Faults = append(pf.Faults, api.Fault{
			Assertion:     faultAssertionOccur,
			FaultId:       generateFaultID(nodeName, fc.physicID),
			FaultType:     "NPU",
			FaultCode:     faultCode,
			FaultTime:     now.Unix(),
			FaultLocation: map[string]string{},
			Influence: []api.Influence{
				{
					NodeName:  nodeName,
					DeviceIds: []int32{int32(id)},
				},
			},
			Description: "hccsping-mesh fault",
		})
	}

	return pf
}

func (f *faultReporter) calCurrentFault(res *types.HccspingMeshResult) *api.PubFaultInfo {
	faultCardIds := f.filterFaultCards(res)
	if len(faultCardIds) == 0 {
		return nil
	}
	return f.constructFaultInfo(faultCardIds)
}

func (f *faultReporter) checkFault(last, cur *api.PubFaultInfo) (*api.PubFaultInfo, bool) {
	hwlog.RunLog.Debugf("checkFault, last: %v, cur: %v", last, cur)
	if last == nil {
		return cur, cur != nil
	}
	newFault := &api.PubFaultInfo{
		Version:   last.Version,
		Id:        last.Id,
		TimeStamp: time.Now().Unix(),
		Resource:  last.Resource,
		Faults:    last.Faults,
	}
	if cur == nil {
		for i, fault := range newFault.Faults {
			fault.Assertion = faultAssertionRecover
			newFault.Faults[i] = fault
		}
		return newFault, true
	}
	oldFaultMap := make(map[string]api.Fault, len(last.Faults))
	curFaultMap := make(map[string]api.Fault, len(cur.Faults))

	for _, fault := range last.Faults {
		oldFaultMap[fault.FaultId] = fault
	}
	for _, fault := range cur.Faults {
		curFaultMap[fault.FaultId] = fault
	}

	change := false
	for faultID, fault := range oldFaultMap {
		if _, ok := curFaultMap[faultID]; !ok {
			fault.Assertion = faultAssertionRecover
			oldFaultMap[faultID] = fault
			change = true
		}
	}

	for faultID, fault := range curFaultMap {
		if _, ok := oldFaultMap[faultID]; !ok {
			oldFaultMap[faultID] = fault
			change = true
		}
	}
	newFault.Faults = make([]api.Fault, 0, len(oldFaultMap))
	for _, fault := range oldFaultMap {
		newFault.Faults = append(newFault.Faults, fault)
	}
	return newFault, change
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
