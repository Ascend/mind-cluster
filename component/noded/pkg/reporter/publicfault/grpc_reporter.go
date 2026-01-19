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

// Package publicfault for report fault device info by grpc
package publicfault

import (
	"ascend-common/common-utils/hwlog"
	"nodeD/pkg/common"
	"nodeD/pkg/grpcclient"
)

const clusterdSvcName = "clusterd-grpc-svc.mindx-dl.svc.cluster.local:8899"

// GrpcReporter report fault device info by grpc
type GrpcReporter struct {
	client *grpcclient.Client
}

// NewGrpcReporter create a grpc reporter
func NewGrpcReporter() *GrpcReporter {
	client, err := grpcclient.New(clusterdSvcName)
	if err != nil {
		return &GrpcReporter{}
	}
	return &GrpcReporter{
		client: client,
	}
}

// Report send fault device info by grpc
func (c *GrpcReporter) Report(fcInfo *common.FaultAndConfigInfo) {
	if fcInfo == nil || fcInfo.PubFaultInfo == nil {
		return
	}
	if c.client == nil {
		client, err := grpcclient.New(clusterdSvcName)
		if err != nil {
			hwlog.RunLog.Errorf("connect to clusterd failed, err is %v", err)
			return
		}
		c.client = client
	}
	_, err := c.client.SendToPubFaultCenter(fcInfo.PubFaultInfo)
	if err != nil {
		hwlog.RunLog.Errorf("send to pub fault failed, err is %v", err)
		return
	}
}

// Init initialize grpc client
func (c *GrpcReporter) Init() error {
	return nil
}
