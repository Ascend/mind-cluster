/* Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package common for common function
package common

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"

	"k8s.io/apimachinery/pkg/util/sets"

	"ascend-common/common-utils/hwlog"
)

const decimal = 10

// DeepEqualFaultDevInfo compare two FaultDevInfo
func DeepEqualFaultDevInfo(this, other *FaultDevInfo) bool {
	if this == nil && other == nil {
		return true
	}
	if this == nil || other == nil {
		return false
	}
	if this.NodeStatus != other.NodeStatus {
		return false
	}
	return faultDevListEqual(this.FaultDevList, other.FaultDevList)
}

type faultDevWithCodeSet struct {
	*FaultDev
	codeSet sets.String
}

func faultDevListEqual(thisList, otherList []*FaultDev) bool {
	if len(thisList) != len(otherList) {
		return false
	}
	thisMap := faultDevListToMap(thisList)
	otherMap := faultDevListToMap(otherList)
	if len(thisMap) != len(otherMap) {
		return false
	}
	for k, v1 := range thisMap {
		v2, ok := otherMap[k]
		if !ok {
			return false
		}
		if v1.FaultLevel != v2.FaultLevel {
			return false
		}
		if !v1.codeSet.Equal(v2.codeSet) {
			return false
		}
	}
	return true
}

func faultDevListToMap(list []*FaultDev) map[string]*faultDevWithCodeSet {
	m := make(map[string]*faultDevWithCodeSet, len(list))
	for _, dev := range list {
		if dev == nil {
			hwlog.RunLog.Error("dev is nil")
			continue
		}
		m[dev.DeviceType+"/"+strconv.FormatInt(dev.DeviceId, decimal)] = &faultDevWithCodeSet{
			FaultDev: dev,
			codeSet:  sets.NewString(dev.FaultCode...),
		}
	}
	return m
}

// GenerateFaultID get uuid by nodeName and Id
func GenerateFaultID(nodeName, id string) string {
	h := sha256.New()
	_, err := h.Write([]byte(nodeName + "/" + id))
	if err != nil {
		hwlog.RunLog.Warnf("GenerateFaultID failed, err: %v", err)
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}
