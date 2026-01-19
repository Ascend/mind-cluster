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
	"encoding/json"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes/fake"
	clientgotest "k8s.io/client-go/testing"

	"ascend-common/api"
	"ascend-common/devmanager/common"
	"nodeD/pkg/kubeclient"
	"nodeD/pkg/pingmesh/types"
	_ "nodeD/pkg/testtool"
)

func TestFaultReporter_HandlePingMeshInfo(t *testing.T) {
	convey.Convey("TestFaultReporter_HandlePingMeshInfo", t, func() {
		cm := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "fake-namespace",
				Name:      "fake-name",
			},
			Data: map[string]string{},
		}
		r := mockReporter()
		fakeClient := fake.NewSimpleClientset()
		convey.Convey("01-get last fault Failed will return error", func() {
			r.client = &kubeclient.ClientK8s{
				ClientSet: fakeClient,
			}
			err := r.HandlePingMeshInfo(&types.HccspingMeshResult{})
			convey.So(err, convey.ShouldBeNil)
			convey.So(r.lastFault, convey.ShouldNotBeNil)
			convey.So(len(r.lastFault.Faults), convey.ShouldEqual, 0)
		})
		fakeClient.AddReactor("get", "configmaps", func(action clientgotest.Action) (handled bool,
			ret runtime.Object, err error) {
			return true, cm, nil
		})
		convey.Convey("02-configmap without public fault will return error", func() {
			r.client = &kubeclient.ClientK8s{
				ClientSet: fakeClient,
			}
			err := r.HandlePingMeshInfo(&types.HccspingMeshResult{})
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestFaultReporter_HandlePingMeshInfo01(t *testing.T) {
	convey.Convey("TestFaultReporter_HandlePingMeshInfo", t, func() {
		const lenOfFaults = 2
		r := mockReporter()
		fakeClient := fake.NewSimpleClientset()
		cm := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "fake-namespace",
				Name:      "fake-name",
			},
			Data: map[string]string{
				api.PubFaultCMDataKey: mockFault([]string{"1", "2"}),
			},
		}
		fakeClient.PrependReactor("get", "configmaps", func(action clientgotest.Action) (handled bool, ret runtime.Object, err error) {
			return true, cm, nil
		})
		convey.Convey("03-configmap with public fault will return nil", func() {
			r.client = &kubeclient.ClientK8s{
				ClientSet: fakeClient,
			}
			err := r.HandlePingMeshInfo(&types.HccspingMeshResult{})
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(r.lastFault.Faults), convey.ShouldEqual, lenOfFaults)
		})
		convey.Convey("04-configmap with public fault will return nil", func() {
			r.client = &kubeclient.ClientK8s{
				ClientSet: fakeClient,
			}
			err := r.HandlePingMeshInfo(&types.HccspingMeshResult{
				Results: map[string]map[uint]*common.HccspingMeshInfo{
					"1": {0: {
						SucPktNum:    []uint{1},
						ReplyStatNum: []int{1},
						DestNum:      1,
					}},
					"2": {0: {
						SucPktNum:    []uint{1},
						ReplyStatNum: []int{1},
						DestNum:      1,
					}},
				},
			})
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(r.lastFault.Faults), convey.ShouldEqual, lenOfFaults)
		})
	})

}

func mockReporter() *faultReporter {
	config := &Config{
		Namespace: "fake-namespace",
		Name:      "fake-name",
		Labels:    map[string]string{"fakekey": "fakevalue"},
		NodeName:  "node",
	}
	return New(config)
}

func mockFault(cards []string) string {
	now := time.Now().Unix()
	newFault := &api.PubFaultInfo{
		Version:   publicFaultVersion,
		Id:        string(uuid.NewUUID()),
		TimeStamp: now,
		Resource:  faultResource,
		Faults:    make([]api.Fault, 0, len(cards)),
	}

	for _, card := range cards {
		newFault.Faults = append(newFault.Faults, constructFaultInfo(card, now))
	}
	fault, err := json.Marshal(newFault)
	if err != nil {
		return ""
	}
	return string(fault)
}
