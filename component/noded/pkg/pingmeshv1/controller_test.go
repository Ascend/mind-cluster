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
Package pingmeshv1 is using for checking hccs network
*/

package pingmeshv1

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"ascend-common/api"
	"nodeD/pkg/kubeclient"
	"nodeD/pkg/pingmeshv1/consts"
	"nodeD/pkg/pingmeshv1/executor"
	"nodeD/pkg/pingmeshv1/resulthandler"
	"nodeD/pkg/pingmeshv1/types"
	_ "nodeD/pkg/testtool"
)

const (
	fakeNode = "node"
)

func TestNewManager(t *testing.T) {
	convey.Convey("Testing New", t, func() {
		convey.Convey("01-new executor failed, should return nil", func() {
			config := &Config{
				KubeClient: &kubeclient.ClientK8s{},
			}
			patch := gomonkey.ApplyFunc(executor.New, func() (*executor.DevManager, error) {
				return nil, fmt.Errorf("executor new failed")
			})
			defer patch.Reset()
			m := NewManager(config)
			convey.So(m, convey.ShouldBeNil)
		})
		convey.Convey("02-new executor success, should not return nil", func() {
			config := &Config{
				KubeClient: &kubeclient.ClientK8s{},
			}
			patch := gomonkey.ApplyFunc(executor.New, func() (*executor.DevManager, error) {
				return &executor.DevManager{SuperPodId: 1}, nil
			})
			defer patch.Reset()
			m := NewManager(config)
			convey.So(m, convey.ShouldNotBeNil)
		})
	})
}

func TestRun(t *testing.T) {
	convey.Convey("Testing Run", t, func() {
		expected := 0
		handler := resulthandler.NewAggregatedHandler(
			func(result *types.HccspingMeshResult) error {
				expected++
				return nil
			})
		fakeClient := fake.NewSimpleClientset()
		patch := gomonkey.ApplyFunc(executor.New, func() (*executor.DevManager, error) {
			return &executor.DevManager{SuperPodId: 1}, nil
		})
		defer patch.Reset()
		patch2 := gomonkey.ApplyMethod(&executor.DevManager{}, "Start", func(_ *executor.DevManager, _ <-chan struct{}) {
			handler.Receive(&types.HccspingMeshResult{})
		})
		defer patch2.Reset()
		m := NewManager(&Config{
			ResultMaxAge: DefaultResultMaxAge,
			KubeClient: &kubeclient.ClientK8s{
				ClientSet: fakeClient,
				NodeName:  fakeNode,
			},
		})
		m.handler = handler

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		m.Run(ctx)
		err := createFakeConfigCM(fakeClient)
		convey.So(err, convey.ShouldBeNil)

		err = createFakeAddrCM(fakeClient, m.ipCmName)
		convey.So(err, convey.ShouldBeNil)

		time.Sleep(time.Second)
	})
}

func createFakeConfigCM(client *fake.Clientset) error {
	globalConfig := types.HccspingMeshConfig{
		Activate:     "on",
		TaskInterval: 1,
	}
	cfg, err := json.Marshal(globalConfig)
	if err != nil {
		return err
	}

	cm1 := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: api.ClusterNS,
			Name:      consts.PingMeshConfigCm,
		},
		Data: map[string]string{
			globalConfigKey: string(cfg),
		},
	}
	_, err = client.CoreV1().ConfigMaps(api.ClusterNS).Create(context.TODO(), cm1,
		metav1.CreateOptions{})
	return err
}

func createFakeAddrCM(client *fake.Clientset, cmName string) error {
	spDevice := &api.SuperPodDevice{
		SuperPodID: "1",
		NodeDeviceMap: map[string]*api.NodeDevice{
			fakeNode: {
				NodeName: fakeNode,
				DeviceMap: map[string]string{
					"0": "0",
					"1": "1",
				},
			},
		},
	}
	spd, err := json.Marshal(spDevice)
	convey.So(err, convey.ShouldBeNil)
	cm2 := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: api.ClusterNS,
			Name:      cmName,
		},
		Data: map[string]string{
			superPodCMKey: string(spd),
		},
	}

	_, err = client.CoreV1().ConfigMaps(api.ClusterNS).Create(context.TODO(), cm2,
		metav1.CreateOptions{})
	return err
}
