/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
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

package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	fakecm "k8s.io/client-go/kubernetes/typed/core/v1/fake"

	ranktablev1 "hccl-controller/pkg/ring-controller/ranktable/v1"
	v2 "hccl-controller/pkg/ring-controller/ranktable/v2"
	_ "hccl-controller/pkg/testtool"
)

const (
	NameSpace = "namespace"
	DataKey   = "hccl.json"
	DataValue = `{"status":"initializing"}`
	CMName    = "rings-config-test1"
)

// TestGetWorkName test GetWorkName
func TestGetWorkName(t *testing.T) {
	convey.Convey("agent GetWorkName", t, func() {
		labels := make(map[string]string, 1)

		convey.Convey(" return volcano-job when label contains VolcanoJobNameKey ", func() {
			labels[VolcanoJobNameKey] = VolcanoJobNameKey
			labels[DeploymentNameKey] = DeploymentNameKey
			work := getWorkName(labels)
			convey.So(work, convey.ShouldEqual, VolcanoJobNameKey)
		})
		convey.Convey("  return deployment-name when label contains VolcanoJobNameKey ", func() {
			labels[DeploymentNameKey] = DeploymentNameKey
			work := getWorkName(labels)
			convey.So(work, convey.ShouldEqual, DeploymentNameKey)
		})
	})
}

// TestUpdateConfigMap test UpdateConfigMap
func TestUpdateConfigMap(t *testing.T) {
	convey.Convey("agent updateConfigMap", t, func() {
		kube := fake.NewSimpleClientset()
		work := &WorkerInfo{kubeclientset: kube, configmapName: CMName}
		convey.Convey(" return err != nil when  cm not exist ", func() {
			err := updateConfigMap(work, NameSpace)
			convey.So(err, convey.ShouldNotEqual, nil)
		})
		convey.Convey(" return err != nil when label in  cm not exist Key910 ", func() {
			data := make(map[string]string, 1)
			putCM := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: CMName,
				Namespace: NameSpace}, Data: data}
			kube.CoreV1().ConfigMaps(NameSpace).Create(context.TODO(), putCM, metav1.CreateOptions{})
			err := updateConfigMap(work, NameSpace)
			convey.So(err, convey.ShouldNotEqual, nil)
		})
		convey.Convey(" return err != nil when update cm error ", func() {
			updateWhenUpdateCmErr(kube, work)
		})
		convey.Convey(" return err == nil when label in  cm normal ", func() {
			updateWhenCMNormal(kube, work)
		})
	})

}

// TestNewCachedIndex test for newCachedIndex
func TestNewCachedIndex(t *testing.T) {
	convey.Convey("test newCachedIndex", t, func() {
		convey.Convey("return empty map when input is 0", func() {
			c := newCachedIndex(0)
			convey.ShouldEqual(c, sync.Map{})
		})
		convey.Convey(" map has value when input is 0", func() {
			const jobReplicas = 2
			c := newCachedIndex(jobReplicas)
			v, exist := c.Load("1")
			convey.ShouldEqual(exist, true)
			convey.ShouldEqual(v.(bool), false)
		})
	})
}

func updateWhenCMNormal(kube *fake.Clientset, work *WorkerInfo) {
	data := make(map[string]string, 1)
	label := make(map[string]string, 1)
	data[DataKey] = DataValue
	label[Key910] = Val910
	putCM := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: CMName,
		Namespace: NameSpace, Labels: label}, Data: data}
	kube.CoreV1().ConfigMaps(NameSpace).Create(context.TODO(), putCM, metav1.CreateOptions{})
	work.configmapData = &ranktablev1.RankTable{RankTableStatus: ranktablev1.RankTableStatus{
		Status: "initializing",
	}}
	work.configmapData.SetStatus(ConfigmapCompleted)
	err := updateConfigMap(work, NameSpace)
	convey.So(err, convey.ShouldEqual, nil)
	cm, _ := kube.CoreV1().ConfigMaps(NameSpace).Get(context.TODO(), CMName,
		metav1.GetOptions{})
	convey.So(cm.Data[DataKey], convey.ShouldEqual, `{"status":"completed","group_list":null,"group_count":""}`)
}

func updateWhenUpdateCmErr(kube *fake.Clientset, work *WorkerInfo) {
	label := make(map[string]string, 1)
	label[Key910] = Val910
	data := make(map[string]string, 1)
	data[DataKey] = DataValue
	putCM := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: CMName,
		Namespace: NameSpace, Labels: label}, Data: data}
	kube.CoreV1().ConfigMaps("namespace").Create(context.TODO(), putCM, metav1.CreateOptions{})
	work.configmapData = &ranktablev1.RankTable{RankTableStatus: ranktablev1.RankTableStatus{
		Status: "initializing",
	}}
	work.configmapData.SetStatus(ConfigmapCompleted)
	patch := gomonkey.ApplyMethod(reflect.TypeOf(kube.CoreV1().ConfigMaps(NameSpace)),
		"Update", func(_ *fakecm.FakeConfigMaps, _ context.Context, _ *corev1.ConfigMap,
			_ metav1.UpdateOptions) (*corev1.ConfigMap, error) {
			return nil, fmt.Errorf("update config error")
		})
	defer patch.Reset()
	err := updateConfigMap(work, NameSpace)
	convey.So(err, convey.ShouldNotEqual, nil)
	cm, _ := kube.CoreV1().ConfigMaps(NameSpace).Get(context.TODO(), CMName,
		metav1.GetOptions{})
	convey.So(cm.Data[DataKey], convey.ShouldEqual, DataValue)
}

// TestWorkerInfoCloseStatistic test WorkerInfo_CloseStatistic
func TestWorkerInfoCloseStatistic(t *testing.T) {
	convey.Convey("agent TestWorkerInfo_CloseStatistic", t, func() {
		w := &WorkerInfo{statisticStopped: true, statisticSwitch: make(chan struct{})}

		convey.Convey(" chan not close when statisticStopped is true ", func() {
			w.CloseStatistic()
			go func() {
				w.statisticSwitch <- struct{}{}
			}()
			_, open := <-w.statisticSwitch
			convey.So(open, convey.ShouldEqual, true)
		})

	})
}

// TestVCJobWorkerStatistic test VCJobWorker_Statistic
func TestVCJobWorkerStatistic(t *testing.T) {
	convey.Convey("agent VCJobWorker_Statistic", t, func() {
		vc := &VCJobWorker{WorkerInfo: WorkerInfo{statisticSwitch: make(chan struct{}), statisticStopped: false}}
		const (
			TaskRep   = 2
			SleepTime = 3
		)

		convey.Convey(" chan will return when chan close ", func() {
			vc.taskReplicasTotal = TaskRep
			vc.cachedPodNum = 1
			go func() {
				time.Sleep(SleepTime * time.Second)
				vc.CloseStatistic()
			}()
			vc.Statistic(1 * time.Second)
		})

		convey.Convey(" chan will return when taskReplicasTotal==cachedPodNum ", func() {
			const CachePod = 2
			vc.taskReplicasTotal = TaskRep
			vc.cachedPodNum = CachePod
			vc.Statistic(1 * time.Second)
		})
	})
}

// TestValidateRank validate rank range
func TestValidateRank(t *testing.T) {
	convey.Convey("test validate rank", t, func() {
		convey.Convey("invalid rank too small", func() {
			err := validate(-1)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("invalid rank too large", func() {
			err := validate(maxRankIndex + 1)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("correct rank", func() {
			err := validate(1)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestGetPodIndex(t *testing.T) {
	convey.Convey("test get PodIndex", t, func() {
		worker := &WorkerInfo{
			cachedIndex:       &sync.Map{},
			taskReplicasTotal: 2,
		}
		convey.Convey("pod with not digital rankIndex will return err", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						PodRankIndexKey: "xxx",
					},
				},
			}
			rank, err := worker.getOrSetPodIndex(pod)
			convey.ShouldEqual(rank, "")
			convey.ShouldNotBeNil(err)
		})
		convey.Convey("pod with invalid rankIndex will return err", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						PodRankIndexKey: "-1",
					},
				},
			}
			rank, err := worker.getOrSetPodIndex(pod)
			convey.ShouldEqual(rank, "")
			convey.ShouldNotBeNil(err)
		})
		convey.Convey("pod with valid rankIndex will return normal rank and nil", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						PodRankIndexKey: "0",
					},
				},
			}
			rank, err := worker.getOrSetPodIndex(pod)
			convey.ShouldEqual(rank, 0)
			convey.ShouldBeNil(err)
		})
	})
}

func TestSetPodIndex(t *testing.T) {
	convey.Convey("test SetPodIndex", t, func() {
		worker := &WorkerInfo{
			cachedIndex:       &sync.Map{},
			taskReplicasTotal: 2,
		}
		convey.Convey("pod without rankIndex will return normal rank and nil", func() {
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Env: []corev1.EnvVar{
								{
									Name:  vcPodIndexKey,
									Value: "1",
								},
							},
						},
					},
				},
			}
			patch := gomonkey.ApplyPrivateMethod(new(WorkerInfo), "updatePod", func(_ *corev1.Pod,
				_ func(*corev1.Pod)) error {
				return nil
			})
			defer patch.Reset()
			rank, err := worker.getOrSetPodIndex(pod)
			convey.ShouldEqual(rank, 1)
			convey.ShouldBeNil(err)
		})
	})
}

func TestCacheRankTable(t *testing.T) {
	vc, fakePod1, fakePod2 := prepareForTest()
	convey.Convey("test cacheRankTable", t, func() {

		convey.Convey("cache pods not reach replicas will return false", func() {
			vc.cachedPodNum = 1
			res := vc.cacheRankTable()
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey("get pods from cache failed will return false", func() {
			patch := gomonkey.ApplyPrivateMethod(new(VCJobWorker), "getPodsFromCache",
				func(_ *VCJobWorker) ([]*corev1.Pod, error) {
					return nil, errors.New("get pods from cache failed")
				})
			defer patch.Reset()
			res := vc.cacheRankTable()
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey("there has pod who has been delete will return false", func() {
			patch := gomonkey.ApplyPrivateMethod(new(VCJobWorker), "getPodsFromCache",
				func(_ *VCJobWorker) ([]*corev1.Pod, error) {
					return []*corev1.Pod{fakePod1, fakePod2}, nil
				})
			defer patch.Reset()
			res := vc.cacheRankTable()
			convey.So(res, convey.ShouldBeFalse)
		})
		convey.Convey("cache ready pods failed will return false", func() {
			fakePod2.DeletionTimestamp = nil
			patch1 := gomonkey.ApplyPrivateMethod(new(VCJobWorker), "getPodsFromCache",
				func(_ *VCJobWorker) ([]*corev1.Pod, error) {
					return []*corev1.Pod{fakePod1, fakePod2}, nil
				})
			defer patch1.Reset()
			patch2 := gomonkey.ApplyPrivateMethod(new(VCJobWorker), "cacheReadyPods", func(_ *VCJobWorker,
				_ []*corev1.Pod) error {
				return errors.New("cacheReadyPods failed")
			})
			defer patch2.Reset()
			res := vc.cacheRankTable()
			convey.So(res, convey.ShouldBeFalse)
		})
	})
}

func TestCacheReadyPods(t *testing.T) {
	const (
		fakeReplicas = 1
	)
	convey.Convey("test cacheReadyPods", t, func() {
		vc := getVCJobWorker(fakeReplicas)
		instance := ranktablev1.Instance{
			Devices: []ranktablev1.Device{
				{
					DeviceID: "1",
					DeviceIP: "0.0.0.0",
				},
			},

			PodName:  "pod1",
			ServerID: "0.0.0.0",
		}
		device, err := json.Marshal(instance)
		if err != nil {
			return
		}
		fakePod := &corev1.Pod{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod",
				Namespace: "default",
				Annotations: map[string]string{
					PodDeviceKey: string(device),
				},
			},
		}

		pods := []*corev1.Pod{fakePod}
		convey.Convey("unmarshal configuration failed will return err", func() {
			patch := gomonkey.ApplyFunc(json.Unmarshal, func(_ []byte, _ interface{}) error {
				return errors.New("unmarshal configuration faild")
			})
			defer patch.Reset()
			err = vc.cacheReadyPods(pods)
			convey.ShouldNotBeNil(err)
		})
		convey.Convey("check configuration failed will return err", func() {
			patch := gomonkey.ApplyFunc(ranktablev1.CheckDeviceInfo, func(_ *ranktablev1.Instance) bool {
				return false
			})
			defer patch.Reset()
			err = vc.cacheReadyPods(pods)
			convey.ShouldNotBeNil(err)
		})
	})
}

func getVCJobWorker(fakeReplicas int32) *VCJobWorker {
	return &VCJobWorker{
		WorkerInfo: WorkerInfo{
			statisticSwitch:   make(chan struct{}),
			statisticStopped:  false,
			statisticMu:       sync.Mutex{},
			taskReplicasTotal: fakeReplicas,
			cachedPodNum:      fakeReplicas,
			configmapData: &v2.RankTable{ServerCount: "1", ServerList: []*v2.Server(nil),
				RankTableStatus: ranktablev1.RankTableStatus{Status: ConfigmapInitializing}, Version: "1.0"},
			cachedPods:    &sync.Map{},
			cachedIndex:   newCachedIndex(int(fakeReplicas)),
			kubeclientset: &kubernetes.Clientset{},
		},
	}
}

func prepareForTest() (*VCJobWorker, *corev1.Pod, *corev1.Pod) {
	const (
		fakeReplicas  = 2
		fakeNamespace = "default"
	)
	vc := &VCJobWorker{
		WorkerInfo: WorkerInfo{
			statisticSwitch:   make(chan struct{}),
			statisticStopped:  false,
			statisticMu:       sync.Mutex{},
			taskReplicasTotal: fakeReplicas,
			cachedPodNum:      fakeReplicas,
		},
	}
	fakePod1 := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: fakeNamespace,
		},
	}
	fakePod2 := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "pod2",
			Namespace:         fakeNamespace,
			DeletionTimestamp: &metav1.Time{},
		},
	}
	return vc, fakePod1, fakePod2
}
