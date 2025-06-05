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

// Package service is to provide other service tools, i.e. clusterd
package service

import (
	"fmt"
	"sync"
	"time"

	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure/storage"
)

func (mpc *MsgProcessor) workerHandler(dataPool *storage.DataPool, data storage.BaseMessage) error {
	if data.Body.MsgType == constant.REGISTER {
		return mpc.workerRegister(dataPool, data)
	}
	workerName := data.Header.Src.Role + data.Header.Src.ServerRank
	workerInfo, err := dataPool.GetWorker(workerName)
	if err != nil {
		return err
	}
	switch data.Body.MsgType {
	case constant.STATUS:
		err := mpc.workerStatus(data, workerInfo)
		if err != nil {
			return err
		}
	case constant.KeepAlive:
		workerInfo.HeartBeat = time.Now()
	default:
		return fmt.Errorf("unknow message type: %v", data.Body.MsgType)
	}
	err = dataPool.UpdateWorker(workerName, workerInfo)
	return err

}

func (mpc *MsgProcessor) workerRegister(dataPool *storage.DataPool, data storage.BaseMessage) error {
	workerInfo := &storage.WorkerInfo{
		Status:     map[string]string{constant.REGISTER: constant.REGISTER},
		GlobalRank: data.Header.Src.ProcessRank,
		Pos:        data.Header.Src,
		HeartBeat:  time.Now(),
		RWMutex:    sync.RWMutex{},
	}
	workerName := data.Header.Src.Role + data.Header.Src.ProcessRank
	err := dataPool.RegisterWorker(workerName, workerInfo)
	return err
}

func (mpc *MsgProcessor) workerStatus(data storage.BaseMessage, workerInfo *storage.WorkerInfo) error {
	statusType := utils.GetThousandsAndHundreds(data.Body.Code)
	switch statusType {
	case constant.ProfilingAllCloseCode:
		return profilingStatus(data, workerInfo)
	default:
		return fmt.Errorf("unknow message status code: %v", data.Body.Code)
	}
}

func profilingStatus(data storage.BaseMessage, workerInfo *storage.WorkerInfo) error {
	commDomainStatus := utils.GetOnesDigit(data.Body.Code)
	defaultDomainStatus := utils.GetTensDigit(data.Body.Code)
	statusMap := map[int32]string{
		constant.OffCode: constant.Off,
		constant.OnCode:  constant.On,
		constant.ExpCode: constant.Exp,
	}
	if statusText, exists := statusMap[commDomainStatus]; exists {
		workerInfo.Status[constant.CommDomainStatus] = statusText
	} else {
		return fmt.Errorf("unknown comm domain status: %v", commDomainStatus)
	}
	if statusText, exists := statusMap[defaultDomainStatus]; exists {
		workerInfo.Status[constant.DefaultDomainStatus] = statusText
	} else {
		return fmt.Errorf("unknown default domain status: %v", defaultDomainStatus)
	}
	return nil
}
