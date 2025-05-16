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

// Package reporter for the report manager main test
package reporter

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"ascend-common/common-utils/hwlog"
	"nodeD/pkg/common"
	"nodeD/pkg/kubeclient"
)

const (
	testNodeName              = "test-node-name"
	testNodeInfoName          = common.NodeInfoCMNamePrefix + testNodeName
	testDeviceType            = "CPU"
	faultCode1                = "00000001"
	faultCode2                = "00000002"
	waitGoroutineFinishedTime = 100 * time.Millisecond
)

var (
	testErr          = errors.New("test error")
	testK8sClient    *kubeclient.ClientK8s
	testFaultDevList = []*common.FaultDev{
		{
			DeviceType: testDeviceType,
			DeviceId:   0,
			FaultCode:  []string{faultCode1, faultCode2},
			FaultLevel: common.PreSeparateFault,
		},
		{
			DeviceType: testDeviceType,
			DeviceId:   1,
			FaultCode:  []string{faultCode1, faultCode2},
			FaultLevel: common.PreSeparateFault,
		},
	}
	testFaultDevInfo = &common.FaultDevInfo{
		FaultDevList: testFaultDevList,
		NodeStatus:   common.PreSeparate,
	}
)

func TestMain(m *testing.M) {
	var patches = gomonkey.ApplyFuncReturn(
		kubeclient.NewClientK8s, &kubeclient.ClientK8s{
			ClientSet:    fake.NewSimpleClientset(),
			NodeName:     testNodeName,
			NodeInfoName: testNodeInfoName,
		}, nil)
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

func deleteCM(cmName, cmNS string) error {
	_, err := testK8sClient.GetConfigMap(cmName, cmNS)
	if err != nil {
		return nil
	}
	return testK8sClient.ClientSet.CoreV1().ConfigMaps(cmNS).Delete(context.TODO(), cmName, v1.DeleteOptions{})
}
