// Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.

// Package kube a series of kube test function
package kube

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"volcano.sh/apis/pkg/apis/batch/v1alpha1"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"
	"volcano.sh/apis/pkg/client/informers/externalversions"

	"ascend-common/api"
	ascendv1 "ascend-common/api/ascend-operator/apis/batch/v1"
	ascendexternalversions "ascend-common/api/ascend-operator/client/informers/externalversions"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/node"
	"clusterd/pkg/domain/superpod"
)

var (
	testTwoDeviceFunc   = 2
	testTwoNodeFunc     = 2
	testTwoPingMeshFunc = 2
	testCmName          = "test-node-name"
	testNodeCheckCode   = "4c97cddcb947bd707778eb50b0986a69768afc2ef3e4f351db0b92e9d07d1fed"
	testDeviceCheckCode = "aaa60c794e2dbec298a2f3c18ea64dea9a1fd2ccdb0cc577b8dfe2c3c5966965"

	testFaultInfo = api.PubFaultInfo{
		Version: "1.0",
	}
)

func TestStopInformer(t *testing.T) {
	convey.Convey("TestStopInformer", t, func() {
		convey.So(StopInformer, convey.ShouldNotPanic)
	})
}

func TestCleanFuncs(t *testing.T) {
	convey.Convey("TestCleanFuncs", t, func() {
		CleanFuncs()
		convey.So(len(cmDeviceFuncs), convey.ShouldEqual, 0)
		convey.So(len(cmNodeFuncs), convey.ShouldEqual, 0)
	})
}

func TestAddCmDeviceFunc(t *testing.T) {
	convey.Convey("TestAddCmDeviceFunc", t, func() {
		convey.Convey("add one device func", func() {
			AddCmDeviceFunc(constant.Resource, func(info *constant.DeviceInfo, info2 *constant.DeviceInfo, s string) {})
			convey.So(len(cmDeviceFuncs[constant.Resource]), convey.ShouldEqual, 1)
		})
		convey.Convey("add two device func", func() {
			AddCmDeviceFunc(constant.Resource, func(info *constant.DeviceInfo, info2 *constant.DeviceInfo, s string) {})
			convey.So(len(cmDeviceFuncs[constant.Resource]), convey.ShouldEqual, testTwoDeviceFunc)
		})
		convey.Convey("add two different business func", func() {
			AddCmDeviceFunc(constant.Statistics, func(info *constant.DeviceInfo, info2 *constant.DeviceInfo, s string) {})
			convey.So(len(cmDeviceFuncs), convey.ShouldEqual, testTwoDeviceFunc)
		})
	})
}

func TestAddCmNodeFunc(t *testing.T) {
	convey.Convey("TestAddCmNodeFunc", t, func() {
		convey.Convey("add one node func", func() {
			AddCmNodeFunc(constant.Resource, func(info *constant.NodeInfo, info2 *constant.NodeInfo, s string) {})
			convey.So(len(cmNodeFuncs[constant.Resource]), convey.ShouldEqual, 1)
		})
		convey.Convey("add two node func", func() {
			AddCmNodeFunc(constant.Resource, func(info *constant.NodeInfo, info2 *constant.NodeInfo, s string) {})
			convey.So(len(cmNodeFuncs[constant.Resource]), convey.ShouldEqual, testTwoNodeFunc)
		})
		convey.Convey("add two different business func", func() {
			AddCmNodeFunc(constant.Statistics, func(info *constant.NodeInfo, info2 *constant.NodeInfo, s string) {})
			convey.So(len(cmNodeFuncs), convey.ShouldEqual, testTwoNodeFunc)
		})
	})
}

func TestAddVCJobFunc(t *testing.T) {
	convey.Convey("TestAddVCJobFunc", t, func() {
		convey.Convey("add one job func", func() {
			AddVCJobFunc(constant.Statistics, func(info *v1alpha1.Job, info2 *v1alpha1.Job, s string) {})
			convey.So(len(vcJobFuncs[constant.Statistics]), convey.ShouldEqual, 1)
		})
		convey.Convey("add two job func", func() {
			AddVCJobFunc(constant.Statistics, func(info *v1alpha1.Job, info2 *v1alpha1.Job, s string) {})
			convey.So(len(vcJobFuncs[constant.Statistics]), convey.ShouldEqual, testTwoNodeFunc)
		})
		convey.Convey("add two different business func", func() {
			AddVCJobFunc(constant.Job, func(info *v1alpha1.Job, info2 *v1alpha1.Job, s string) {})
			convey.So(len(vcJobFuncs), convey.ShouldEqual, testTwoNodeFunc)
		})
	})
}

func TestAddACJobFunc(t *testing.T) {
	convey.Convey("TestAddACJobFunc", t, func() {
		convey.Convey("add one job func", func() {
			AddACJobFunc(constant.Statistics, func(info *ascendv1.AscendJob, info2 *ascendv1.AscendJob, s string) {})
			convey.So(len(acJobFuncs[constant.Statistics]), convey.ShouldEqual, 1)
		})
		convey.Convey("add two job func", func() {
			AddACJobFunc(constant.Statistics, func(info *ascendv1.AscendJob, info2 *ascendv1.AscendJob, s string) {})
			convey.So(len(acJobFuncs[constant.Statistics]), convey.ShouldEqual, testTwoNodeFunc)
		})
		convey.Convey("add two different business func", func() {
			AddACJobFunc(constant.Job, func(info *ascendv1.AscendJob, info2 *ascendv1.AscendJob, s string) {})
			convey.So(len(acJobFuncs), convey.ShouldEqual, testTwoNodeFunc)
		})
	})
}

func TestCheckConfigMapIsDeviceInfo(t *testing.T) {
	convey.Convey("test checkConfigMapIsNodeInfo", t, func() {
		var obj interface{}
		mockMatchedFalse := gomonkey.ApplyFunc(util.IsNSAndNameMatched, func(obj interface{},
			namespace string, namePrefix string) bool {
			return false
		})
		defer mockMatchedFalse.Reset()
		cmCheck := checkConfigMapIsDeviceInfo(obj)
		convey.So(cmCheck, convey.ShouldBeFalse)
	})
}

func TestCheckConfigMapIsNodeInfo(t *testing.T) {
	convey.Convey("test checkConfigMapIsNodeInfo", t, func() {
		var obj interface{}
		mockMatchedTrue := gomonkey.ApplyFunc(util.IsNSAndNameMatched, func(obj interface{},
			namespace string, namePrefix string) bool {
			return true
		})
		defer mockMatchedTrue.Reset()
		nodeCheck := checkConfigMapIsNodeInfo(obj)
		convey.So(nodeCheck, convey.ShouldBeTrue)
	})
}

func TestInitClusterDevice(t *testing.T) {
	convey.Convey("Test initClusterDevice", t, func() {
		fakeNodes := []*v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node-1",
				},
			},
		}
		patchGetNodes := gomonkey.ApplyFunc(getNodesFromInformer, func() []*v1.Node {
			return fakeNodes
		})
		defer patchGetNodes.Reset()

		patchGetNodeDevice := gomonkey.ApplyFunc(node.GetNodeDeviceAndSuperPodID,
			func(node *v1.Node) (*api.NodeDevice, string) {
				return &api.NodeDevice{
					NodeName: node.Name,
				}, "test-superpod-id"
			})
		defer patchGetNodeDevice.Reset()

		var calledSaveNode bool
		patchSaveNode := gomonkey.ApplyFunc(superpod.SaveNode,
			func(superPodID string, node *api.NodeDevice) {
				calledSaveNode = true
				convey.So(superPodID, convey.ShouldEqual, "test-superpod-id")
				convey.So(node.NodeName, convey.ShouldEqual, "test-node-1")
			})
		defer patchSaveNode.Reset()

		initClusterDevice()
		convey.So(calledSaveNode, convey.ShouldBeFalse)
	})
}

func TestConfigMapIsEpRankTableInfo(t *testing.T) {
	convey.Convey("test ConfigMapIsEpRankTableInfo", t, func() {
		convey.Convey("when object is not a ConfigMap", func() {
			obj := "not a configmap"
			result := checkConfigMapIsEpRankTableInfo(obj)
			convey.So(result, convey.ShouldBeFalse)
		})
		convey.Convey("when ConfigMap name has the correct prefix", func() {
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: constant.MindIeRanktablePrefix + "example",
				},
			}
			result := checkConfigMapIsEpRankTableInfo(cm)
			convey.So(result, convey.ShouldBeTrue)
		})
		convey.Convey("when ConfigMap name does not have the correct prefix", func() {
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "wrong-prefix-example",
				},
			}
			result := checkConfigMapIsEpRankTableInfo(cm)
			convey.So(result, convey.ShouldBeFalse)
		})
	})
}

func TestAddPodGroupFunc(t *testing.T) {
	convey.Convey("Test AddPodGroupFunc", t, func() {
		business := "testBusiness"
		func1 := func(pg1, pg2 *v1beta1.PodGroup, s string) {}
		AddPodGroupFunc(business, func1)
		convey.So(len(podGroupFuncs[business]), convey.ShouldEqual, 1)
	})
}

func TestAddPodFunc(t *testing.T) {
	convey.Convey("Test AddPodFunc", t, func() {
		business := "testBusiness"
		func1 := func(p1, p2 *v1.Pod, s string) {}
		AddPodFunc(business, func1)
		convey.So(len(podFuncs[business]), convey.ShouldEqual, 1)
	})
}

func TestAddNodeFunc(t *testing.T) {
	convey.Convey("Test AddNodeFunc", t, func() {
		business := "testBusiness"
		func1 := func(n1, n2 *v1.Node, s string) {}
		AddNodeFunc(business, func1)
		convey.So(len(nodeFuncs[business]), convey.ShouldEqual, 1)
	})
}

func TestCheckConfigMapIsSwitchInfo(t *testing.T) {
	convey.Convey("Test checkConfigMapIsSwitchInfo", t, func() {
		patch := gomonkey.ApplyFuncReturn(util.IsNSAndNameMatched, true)
		value := checkConfigMapIsSwitchInfo(nil)
		convey.ShouldBeTrue(value)
		patch.Reset()

		patch = gomonkey.ApplyFuncReturn(util.IsNSAndNameMatched, false)
		value = checkConfigMapIsSwitchInfo(nil)
		convey.ShouldBeFalse(value)
		patch.Reset()
	})
}

func TestAddCmPubFaultFunc(t *testing.T) {
	convey.Convey("Test AddCmPubFaultFunc", t, func() {
		business := "testBusiness"
		func1 := func(pf1, pf2 *api.PubFaultInfo, s string) {}

		AddCmPubFaultFunc(business, func1)

		convey.So(len(cmPubFaultFuncs[business]), convey.ShouldEqual, 1)
	})
}

func TestAddCmRankTableFunc(t *testing.T) {
	convey.Convey("Test AddCmRankTableFunc", t, func() {
		business := "testBusiness"
		func1 := func(i1, i2 interface{}, s string) {}

		AddCmRankTableFunc(business, func1)

		convey.So(len(cmRankTableFuncs[business]), convey.ShouldEqual, 1)
	})
}

func TestInitVCJobInformer(t *testing.T) {
	convey.Convey("Test InitVCJobInformer", t, func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		_, err := InitClientVolcano()
		if err != nil {
			return
		}
		called := false
		patches.ApplyFunc(externalversions.SharedInformerFactory.Start, func(stopCh <-chan struct{}) {
			called = true
		})
		patches.ApplyFunc(externalversions.SharedInformerFactory.WaitForCacheSync,
			func(stopCh <-chan struct{}) map[reflect.Type]bool {
				return make(map[reflect.Type]bool)
			})
		InitVCJobInformer()
		convey.So(called, convey.ShouldEqual, true)
	})

}

func TestInitACJobInformer(t *testing.T) {
	convey.Convey("Test InitVCJobInformer", t, func() {
		patches := gomonkey.NewPatches()
		defer patches.Reset()
		_, err := InitClientVolcano()
		if err != nil {
			return
		}
		called := false
		patches.ApplyFunc(ascendexternalversions.SharedInformerFactory.Start, func(stopCh <-chan struct{}) {
			called = true
		})
		patches.ApplyFunc(ascendexternalversions.SharedInformerFactory.WaitForCacheSync,
			func(stopCh <-chan struct{}) map[reflect.Type]bool {
				return make(map[reflect.Type]bool)
			})
		InitACJobInformer()
		convey.So(called, convey.ShouldEqual, true)
	})

}

func TestAddCmConfigPingMeshFunc(t *testing.T) {
	convey.Convey("When adding functions for a new business", t, func() {
		defer CleanFuncs()
		business := "test_business"
		func1 := func(cm1, cm2 constant.ConfigPingMesh, s string) {}
		func2 := func(cm1, cm2 constant.ConfigPingMesh, s string) {}

		convey.So(cmPingMeshCMFuncs[business], convey.ShouldBeNil)

		AddCmConfigPingMeshFunc(business, func1)
		convey.So(cmPingMeshCMFuncs[business], convey.ShouldNotBeNil)
		convey.So(len(cmPingMeshCMFuncs[business]), convey.ShouldEqual, 1)
		convey.So(cmPingMeshCMFuncs[business][0], convey.ShouldEqual, func1)

		AddCmConfigPingMeshFunc(business, func2)
		convey.So(len(cmPingMeshCMFuncs[business]), convey.ShouldEqual, testTwoPingMeshFunc)
		convey.So(cmPingMeshCMFuncs[business][0], convey.ShouldEqual, func1)
		convey.So(cmPingMeshCMFuncs[business][1], convey.ShouldEqual, func2)
	})
}

func TestGetNodeFromIndexer(t *testing.T) {
	convey.Convey("Test GetNodeFromIndexer", t, func() {
		convey.Convey("when nodeInformer is not nil, but indexer is empty, should return err", func() {
			factory := informers.NewSharedInformerFactoryWithOptions(k8sClient.ClientSet, 0)
			nodeInformer = factory.Core().V1().Nodes().Informer()
			_, err := GetNodeFromIndexer("testNode")
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestNodeHandler(t *testing.T) {
	convey.Convey("Test nodeHandler", t, func() {
		isAdd := false
		AddNodeFunc("testType", func(_ *v1.Node, _ *v1.Node, _ string) {
			isAdd = true
		})
		defer CleanFuncs()
		convey.Convey("when newObj is nil, should not exec function", func() {
			nodeHandler(nil, nil, "add")
			convey.So(isAdd, convey.ShouldBeFalse)
		})
		convey.Convey("when newObj is not nil, should exec function", func() {
			fakeNode := v1.Node{}
			nodeHandler(nil, &fakeNode, "add")
			convey.So(isAdd, convey.ShouldBeTrue)
		})
	})
}

func TestPodHandler(t *testing.T) {
	convey.Convey("Test podHandler", t, func() {
		isAdd := false
		AddPodFunc("testType", func(_ *v1.Pod, _ *v1.Pod, _ string) {
			isAdd = true
		})
		defer CleanFuncs()
		convey.Convey("when newObj is nil, should not exec function", func() {
			podHandler(nil, nil, "add")
			convey.So(isAdd, convey.ShouldBeFalse)
		})
		convey.Convey("when newObj is not nil, should exec function", func() {
			fakePod := v1.Pod{}
			podHandler(nil, &fakePod, "add")
			convey.So(isAdd, convey.ShouldBeTrue)
		})
	})
}

func TestCmDeviceHandler(t *testing.T) {
	convey.Convey("Test cmDeviceHandler", t, func() {
		isAdd := false
		AddCmDeviceFunc("testType", func(_ *constant.DeviceInfo, _ *constant.DeviceInfo, _ string) {
			isAdd = true
		})
		defer CleanFuncs()
		convey.Convey("when newObj is nil, should not exec function", func() {
			cmDeviceHandler(nil, nil, "add")
			convey.So(isAdd, convey.ShouldBeFalse)
		})
		convey.Convey("when newObj is not nil, should exec function", func() {
			fakeCm := &v1.ConfigMap{}
			fakeCm.Name = testCmName
			devInfoCM := constant.DeviceInfoCM{}
			devInfoCM.CheckCode = testDeviceCheckCode
			devInfoCM.SuperPodID = 1
			devInfoCM.SuperPodID = 1
			devInfoCM.DeviceInfo = constant.DeviceInfoNoName{
				UpdateTime: 0,
			}
			fakeCm.Data = map[string]string{}
			fakeCm.Data[api.DeviceInfoCMDataKey] = util.ObjToString(devInfoCM)
			cmDeviceHandler(nil, fakeCm, "add")
			convey.So(isAdd, convey.ShouldBeTrue)
		})
	})
}

func TestCmNodeHandler(t *testing.T) {
	convey.Convey("Test cmNodeHandler", t, func() {
		isAdd := false
		AddCmNodeFunc("testType", func(_ *constant.NodeInfo, _ *constant.NodeInfo, _ string) {
			isAdd = true
		})
		defer CleanFuncs()
		convey.Convey("when newObj is nil, should not exec function", func() {
			cmNodeHandler(nil, nil, "add")
			convey.So(isAdd, convey.ShouldBeFalse)
		})
		convey.Convey("when newObj is not nil, should exec function", func() {
			fakeCm := &v1.ConfigMap{}
			fakeCm.Name = testCmName
			nodeInfoCM := constant.NodeInfoCM{}
			nodeInfoCM.CheckCode = testNodeCheckCode
			nodeInfoCM.NodeInfo = constant.NodeInfoNoName{}
			fakeCm.Data = map[string]string{}
			fakeCm.Data[api.NodeInfoCMDataKey] = util.ObjToString(nodeInfoCM)
			cmNodeHandler(nil, fakeCm, "add")
			convey.So(isAdd, convey.ShouldBeTrue)
		})
	})
}

func TestCmPingMeshConfigHandler(t *testing.T) {
	convey.Convey("Test cmPingMeshConfigHandler", t, func() {
		isAdd := false
		AddCmConfigPingMeshFunc("testType", func(_ constant.ConfigPingMesh,
			_ constant.ConfigPingMesh, _ string) {
			isAdd = true
		})
		defer CleanFuncs()
		convey.Convey("when newObj is nil, should not exec function", func() {
			cmPingMeshConfigHandler(nil, nil, "add")
			convey.So(isAdd, convey.ShouldBeFalse)
		})
		convey.Convey("when newObj is not nil, should exec function", func() {
			fakeCm := &v1.ConfigMap{}
			fakeCm.Name = "testCmName"
			configInfo := constant.ConfigPingMesh{}
			configInfo["global"] = &constant.HccspingMeshItem{Activate: "on"}
			fakeCm.Data = map[string]string{"global": util.ObjToString(configInfo)}
			cmPingMeshConfigHandler(nil, fakeCm, "add")
			convey.So(isAdd, convey.ShouldBeTrue)
		})
	})
}

func TestCmSwitchHandler(t *testing.T) {
	convey.Convey("Test cmSwitchHandler", t, func() {
		isAdd := false
		AddCmSwitchFunc("testType", func(_ *constant.SwitchInfo,
			_ *constant.SwitchInfo, _ string) {
			isAdd = true
		})
		defer CleanFuncs()
		convey.Convey("when newObj is not nil, should exec function", func() {
			swit := constant.SwitchFaultInfo{
				FaultInfo:  []constant.SimpleSwitchFaultInfo{},
				FaultLevel: "FaultLevel",
				UpdateTime: 0,
				NodeStatus: "Healthy",
			}
			bytes, err := json.Marshal(swit)
			convey.So(err, convey.ShouldBeNil)
			fakeCm := v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: constant.SwitchInfoPrefix + "testName",
				},
				Data: map[string]string{api.SwitchInfoCMDataKey: string(bytes)},
			}
			cmSwitchHandler(nil, &fakeCm, "add")
			convey.So(isAdd, convey.ShouldBeTrue)
		})
	})
}

func TestCmPubFaultHandler(t *testing.T) {
	convey.Convey("Test cmPubFaultHandler", t, func() {
		isAdd := false
		AddCmPubFaultFunc("testType", func(_ *api.PubFaultInfo,
			_ *api.PubFaultInfo, _ string) {
			isAdd = true
		})
		defer CleanFuncs()
		convey.Convey("when newObj is nil, should not exec function", func() {
			cmPubFaultHandler(nil, nil, "add")
			convey.So(isAdd, convey.ShouldBeFalse)
		})
		convey.Convey("when newObj is not nil, should exec function", func() {
			faultData, err := json.Marshal(testFaultInfo)
			if err != nil {
				t.Error(err)
			}
			fakeCm := v1.ConfigMap{
				Data: map[string]string{api.PubFaultCMDataKey: string(faultData)},
			}
			cmPubFaultHandler(nil, &fakeCm, "add")
			convey.So(isAdd, convey.ShouldBeTrue)
		})
	})
}

func TestListObjects(t *testing.T) {
	convey.Convey("Test ListObjects", t, func() {
		convey.Convey("with no indexers", func() {
			indexers = make(map[schema.GroupVersionKind]cache.Indexer)
			result := ListObjects(PodGVK())
			convey.So(len(result), convey.ShouldEqual, 0)
		})
		convey.Convey("with valid indexer", func() {
			indexers = make(map[schema.GroupVersionKind]cache.Indexer)
			mockIndexer := &mockIndexer{items: []interface{}{&v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test-pod"}}}}
			indexers[PodGVK()] = mockIndexer
			result := ListObjects(PodGVK())
			convey.So(len(result), convey.ShouldEqual, 1)
		})
	})
}

func TestGetObject(t *testing.T) {
	convey.Convey("Test GetObject", t, func() {
		convey.Convey("with no indexers", func() {
			indexers = make(map[schema.GroupVersionKind]cache.Indexer)
			_, err := GetObject(PodGVK(), "test-pod")
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldResemble, "gvk /v1, Kind=Pod not found")
		})
		convey.Convey("with key not exist", func() {
			indexers = make(map[schema.GroupVersionKind]cache.Indexer)
			mockIndexer := &mockIndexer{items: []interface{}{}}
			indexers[PodGVK()] = mockIndexer
			_, err := GetObject(PodGVK(), "test-pod")
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldResemble, "get /v1, Kind=Pod test-pod failed")
		})
		convey.Convey("with valid key", func() {
			indexers = make(map[schema.GroupVersionKind]cache.Indexer)
			testPod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test-pod", Namespace: "default"}}
			mockIndexer := &mockIndexer{items: []interface{}{testPod}, keyFunc: func(obj interface{}) (string, error) {
				return "default/test-pod", nil
			}}
			indexers[PodGVK()] = mockIndexer
			result, err := GetObject(PodGVK(), "default/test-pod")
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldNotBeNil)
		})
	})
}

type mockIndexer struct {
	items   []interface{}
	keyFunc func(obj interface{}) (string, error)
}

func (m *mockIndexer) List() []interface{} { return m.items }
func (m *mockIndexer) ListKeys() []string  { return nil }
func (m *mockIndexer) Get(obj interface{}) (interface{}, bool, error) {
	return nil, false, nil
}
func (m *mockIndexer) GetByKey(key string) (interface{}, bool, error) {
	for _, item := range m.items {
		itemKey, err := m.keyFunc(item)
		if err != nil {
			return nil, false, err
		}
		if itemKey == key {
			return item, true, nil
		}
	}
	return nil, false, nil
}
func (m *mockIndexer) Replace(items []interface{}, resourceVersion string) error { return nil }
func (m *mockIndexer) ByIndex(indexName, indexedValue string) ([]interface{}, error) {
	return nil, nil
}
func (m *mockIndexer) Index(indexName string, obj interface{}) ([]interface{}, error) {
	return nil, nil
}
func (m *mockIndexer) IndexKeys(indexName, indexedValue string) ([]string, error) { return nil, nil }
func (m *mockIndexer) ListIndexFuncValues(indexName string) []string              { return nil }
func (m *mockIndexer) ByIndexKeys(indexName string, indexKeys ...string) ([]interface{}, error) {
	return nil, nil
}
func (m *mockIndexer) Add(obj interface{}) error                 { return nil }
func (m *mockIndexer) Update(obj interface{}) error              { return nil }
func (m *mockIndexer) Delete(obj interface{}) error              { return nil }
func (m *mockIndexer) AddIndexers(indexers cache.Indexers) error { return nil }
func (m *mockIndexer) GetIndexers() cache.Indexers               { return nil }
func (m *mockIndexer) Resync() error                             { return nil }
