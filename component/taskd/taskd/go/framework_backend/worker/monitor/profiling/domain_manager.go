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

// Package profiling contains functions that support dynamically collecting profiling data
package profiling

import "C"
import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"ascend-common/common-utils/hwlog"
	"taskd/common/constant"
)

// SwitchProfiling is the struct for serialization and deserialization of profiling switches
type SwitchProfiling struct {
	CommunicationOperator string
	Step                  string
	SaveCheckpoint        string
	FP                    string
	DataLoader            string
}

// GetProfilingSwitch get profile switch status from file, if any fault happened return all switch off
func GetProfilingSwitch(filePath string) SwitchProfiling {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		// if reading failed close all
		hwlog.RunLog.Errorf("failed to read file %s, err%v", filePath, err)
		return SwitchProfiling{
			CommunicationOperator: constant.SwitchOFF,
			Step:                  constant.SwitchOFF,
			SaveCheckpoint:        constant.SwitchOFF,
			FP:                    constant.SwitchOFF,
			DataLoader:            constant.SwitchOFF,
		}
	}

	var profiling SwitchProfiling

	err = json.Unmarshal(data, &profiling)
	if err != nil {
		hwlog.RunLog.Errorf("failed to parse profiling switch %#v: %v", profiling, err)
		return SwitchProfiling{
			CommunicationOperator: constant.SwitchOFF,
			Step:                  constant.SwitchOFF,
			SaveCheckpoint:        constant.SwitchOFF,
			FP:                    constant.SwitchOFF,
			DataLoader:            constant.SwitchOFF,
		}
	}
	return profiling
}

// ManageDomainEnableStatus dead loop for manage domain status
func ManageDomainEnableStatus(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			hwlog.RunLog.Errorf("manager of changing domain manager has paniced, err: %v", r)
			fmt.Printf("[ERROR] %s manager of changing domain manager has paniced, err: %v\n", time.Now(), r)
		}
	}()
	hwlog.RunLog.Infof("start to watch for domain config changes")
	lastStatus := ""
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Warnf("domain config received exit signal")
			return
		default:
			profilingSwitches := GetProfilingSwitch(constant.ProfilingSwitchFilePath)
			if lastStatus == getProfilingStatusStr(profilingSwitches) {
				hwlog.RunLog.Debug("status not changed will not call mspti")
				time.Sleep(constant.DomainCheckInterval)
				continue
			}
			changeProfileSwitchStatus(profilingSwitches)
			lastStatus = getProfilingStatusStr(profilingSwitches)
			time.Sleep(constant.DomainCheckInterval)
		}
	}
}

func changeProfileSwitchStatus(profilingSwitches SwitchProfiling) {
	// if all kinds of records are off,  disable all marker
	if profilingSwitches.Step == constant.SwitchOFF && profilingSwitches.SaveCheckpoint == constant.SwitchOFF &&
		profilingSwitches.FP == constant.SwitchOFF && profilingSwitches.DataLoader == constant.SwitchOFF &&
		profilingSwitches.CommunicationOperator == constant.SwitchOFF {
		if err := DisableMsptiActivity(); err != nil {
			hwlog.RunLog.Errorf("failed to disable MsptiActivity: %v", err)
		}
	} else {
		// any kind of domain is on, need to enable marker, FP/dataloader/ckpt/step will be enabled
		if err := EnableMsptiMarkerActivity(); err != nil {
			hwlog.RunLog.Error(err)
		}
		// only change status of communication dynamically
		if err := EnableMarkerDomain(constant.CommunicationDomainName,
			profilingSwitches.CommunicationOperator); err != nil {
			hwlog.RunLog.Errorf("failed to change communication marker domain status, err: %v", err)
		}
	}
}

func getProfilingStatusStr(profilingSwiches SwitchProfiling) string {
	return fmt.Sprintf("communication:%s,step:%s,FP:%s,dataloader:%s,ckpt:%s",
		profilingSwiches.CommunicationOperator, profilingSwiches.Step, profilingSwiches.FP,
		profilingSwiches.DataLoader, profilingSwiches.SaveCheckpoint)
}
