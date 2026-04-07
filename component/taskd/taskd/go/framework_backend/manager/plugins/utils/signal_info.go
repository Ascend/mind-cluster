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

// Package utils for common func
package utils

import (
	"errors"
	"strconv"

	"ascend-common/common-utils/hwlog"
	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/framework_backend/manager/infrastructure/storage"
)

const (
	tenBase   = 10
	bitSize64 = 64
)

// ParseSignalInfo parses signal info
func ParseSignalInfo(clusterInfo *storage.ClusterInfo) (*SignalInfo, error) {
	if clusterInfo == nil || clusterInfo.Command == nil {
		return nil, errors.New("cluster info or command is nil")
	}

	signalInfo := &SignalInfo{
		SignalType:     clusterInfo.Command[constant.SignalType],
		ChangeStrategy: clusterInfo.Command[constant.ChangeStrategy],
		ExtraParams:    clusterInfo.Command[constant.ExtraParams],
		Command:        clusterInfo.Command,
		Uuid:           clusterInfo.Command[constant.Uuid],
	}

	if signalInfo.SignalType == "" {
		return signalInfo, nil
	}

	if err := parseSignalFields(signalInfo, clusterInfo.Command); err != nil {
		return nil, err
	}

	return signalInfo, nil
}

func parseSignalFields(signalInfo *SignalInfo, command map[string]string) error {
	var err error

	signalInfo.Timeout, err = strconv.ParseInt(command[constant.Timeout], tenBase, bitSize64)
	if err != nil {
		hwlog.RunLog.Errorf("ParseInt failed: %v", err)
		return err
	}

	signalInfo.Actions, err = utils.StringToObj[[]string](command[constant.Actions])
	if err != nil {
		hwlog.RunLog.Errorf("unmarshal actions failed: %v", err)
		return err
	}

	signalInfo.FaultRanks, err = utils.StringToObj[map[int]int](command[constant.FaultRanks])
	if err != nil {
		hwlog.RunLog.Errorf("unmarshal FaultRanks failed: %v", err)
		return err
	}

	signalInfo.NodeRankIds, err = utils.StringToObj[[]string](command[constant.NodeRankIds])
	if err != nil {
		hwlog.RunLog.Errorf("unmarshal NodeRankIds failed: %v", err)
		return err
	}

	return nil
}
