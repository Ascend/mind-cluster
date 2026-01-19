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

// Package nodesn for read SN
package nodesn

import (
	"fmt"

	"github.com/u-root/u-root/pkg/ipmi"

	"ascend-common/common-utils/hwlog"
)

var (
	serialNumberReq     = []byte{0x30, 0x90, 0x05, 0x00, 0x03, 0x04, 0x00, 0xff}
	manufacturerNameReq = []byte{0x30, 0x90, 0x05, 0x00, 0x03, 0x00, 0x00, 0xff}
)

const snFieldStartIndex = 2

// GetNodeSN get node sn
func GetNodeSN() (string, error) {
	ipmiTool, err := ipmi.Open(0)
	if err != nil {
		hwlog.RunLog.Errorf("open ipmi device failed, err is %v", err)
		return "", err
	}
	defer func(ipmiTool *ipmi.IPMI) {
		err := ipmiTool.Close()
		if err != nil {
			hwlog.RunLog.Errorf("close ipmi failed, err is %v", err)
		}
	}(ipmiTool)
	response, err := ipmiTool.RawCmd(serialNumberReq)
	if err != nil {
		hwlog.RunLog.Errorf("get serial number failed, err is %v", err)
		return "", err
	}
	if len(response) <= snFieldStartIndex {
		hwlog.RunLog.Info("serial number not found, use manufacturer name")
		response, err = ipmiTool.RawCmd(manufacturerNameReq)
		if err != nil {
			hwlog.RunLog.Errorf("get manufacturer name failed, err is %v", err)
			return "", err
		} else if len(response) <= snFieldStartIndex {
			hwlog.RunLog.Errorf("get manufacturer name failed, (%v)length is too short", response)
			return "", fmt.Errorf("get manufacturer name failed, (%v)length is too short", response)
		}
	}
	snMsgByte := response[snFieldStartIndex:]
	return string(snMsgByte), nil
}
