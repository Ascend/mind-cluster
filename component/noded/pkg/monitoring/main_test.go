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

// Package monitoring for the monitor manager main test
package monitoring

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/u-root/u-root/pkg/ipmi"
	"k8s.io/client-go/kubernetes/fake"

	"ascend-common/common-utils/hwlog"
	"nodeD/pkg/common"
	"nodeD/pkg/kubeclient"
	"nodeD/pkg/monitoring/ipmimonitor"
)

const (
	testDeviceType            = "CPU"
	faultCode1                = "00000001"
	faultCode2                = "00000002"
	waitGoroutineFinishedTime = 100 * time.Millisecond
)

var (
	testK8sClient   = &kubeclient.ClientK8s{}
	currentAlarmReq = []byte{0x30, 0x94, 0xDB, 0x07, 0x00, 0x40, 0x00, 0x00, 0x00, 0x0E, 0xFF}
)

func TestMain(m *testing.M) {
	testFaultEvents := []*common.FaultEvent{
		{
			ErrorCode:  faultCode1,
			Severity:   0,
			DeviceType: testDeviceType,
			DeviceId:   0,
		},
		{
			ErrorCode:  faultCode2,
			Severity:   1,
			DeviceType: testDeviceType,
			DeviceId:   1,
		},
	}

	const testNodeName = "test-node-name"
	var patches = gomonkey.ApplyFuncReturn(
		kubeclient.NewClientK8s, &kubeclient.ClientK8s{
			ClientSet:    fake.NewSimpleClientset(),
			NodeName:     testNodeName,
			NodeInfoName: common.NodeInfoCMNamePrefix + testNodeName,
		}, nil).
		ApplyFuncReturn(ipmi.Open, &ipmi.IPMI{}, nil).
		ApplyMethodReturn(&ipmi.IPMI{}, "RawCmd", currentAlarmReq, nil).
		ApplyMethodReturn(&ipmi.IPMI{}, "Close", nil).
		ApplyGlobalVar(&common.ParamOption, common.Option{MonitorPeriod: 10}).
		ApplyMethodReturn(&ipmimonitor.IpmiEventMonitor{}, "GetCurrentAlarmFaultEvents", testFaultEvents, nil)
	defer patches.Reset()

	if err := setup(); err != nil {
		return
	}
	code := m.Run()
	fmt.Printf("exit_code = %v\n", code)
}

func setup() error {
	if err := initLog(); err != nil {
		return err
	}
	if err := initK8sClient(); err != nil {
		return err
	}
	return nil
}

func initLog() error {
	logConfig := &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	if err := hwlog.InitRunLogger(logConfig, context.Background()); err != nil {
		fmt.Printf("init hwlog failed, %v\n", err)
		return errors.New("init hwlog failed")
	}
	return nil
}

func initK8sClient() error {
	var err error
	testK8sClient, err = kubeclient.NewClientK8s()
	if err != nil {
		hwlog.RunLog.Errorf("init k8s client failed when start, err: %v", err)
		return err
	}
	return nil
}
