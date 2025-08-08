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

// Package slownode is a DT collection for func in config_map_watcher
package slownode

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/model/slownode"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/cluster"
	"ascend-faultdiag-online/pkg/service/servicefunc/slownode/constants"
	"ascend-faultdiag-online/pkg/utils/k8s"
)

var (
	zero          = 0
	one           = 1
	two           = 2
	three         = 3
	logLineLength = 256
	testCmName    = "ras-feature-slownode"
)

func init() {
	config := hwlog.LogConfig{
		OnlyToStdout:  true,
		MaxLineLength: logLineLength,
	}
	err := hwlog.InitRunLogger(&config, context.TODO())
	if err != nil {
		fmt.Println(err)
	}
}

func TestStopInformer(t *testing.T) {
	convey.Convey("TestStopInformer", t, func() {
		convey.So(stopInformer, convey.ShouldNotPanic)
	})
}

func TestCleanFuncs(t *testing.T) {
	convey.Convey("TestCleanFuncs", t, func() {
		cleanFunc()
		convey.So(len(jobFuncList), convey.ShouldEqual, zero)
	})
}

func TestAddCmSLFeatFunc(t *testing.T) {
	convey.Convey("TestAddCmSLFeatFunc", t, func() {
		convey.Convey("add one SLFeat func", func() {
			addCMHandler(&jobFuncList, func(info *slownode.Job, info2 *slownode.Job, s watch.EventType) {})
			convey.So(len(jobFuncList), convey.ShouldEqual, one)
		})
		convey.Convey("add two SLFeat func", func() {
			addCMHandler(&jobFuncList, func(info *slownode.Job, info2 *slownode.Job, s watch.EventType) {})
			convey.So(len(jobFuncList), convey.ShouldEqual, two)
		})
		convey.Convey("add two different business func", func() {
			addCMHandler(&jobFuncList, func(info *slownode.Job, info2 *slownode.Job, s watch.EventType) {})
			convey.So(len(jobFuncList), convey.ShouldEqual, three)
		})
	})
}

func TestCheckConfigMapIsSlowNodeFeatConf(t *testing.T) {
	convey.Convey("test filterSlowNodeJob", t, func() {
		var obj any
		mockMatchedTrue := gomonkey.ApplyFunc(isNameMatched, func(any, string) bool {
			return true
		})
		defer mockMatchedTrue.Reset()
		slowNodeCheck := filterClusterJob(obj)
		convey.So(slowNodeCheck, convey.ShouldBeTrue)
	})
}

func TestInitInformer(t *testing.T) {
	convey.Convey("TestInitInformer", t, func() {
		convey.Convey("get k8s client failed", func() {
			patch := gomonkey.ApplyFunc(k8s.GetClient, func() (*k8s.Client, error) {
				return nil, fmt.Errorf("mock get k8s client error")
			})
			defer patch.Reset()
			initCMInformer()
		})
		patch := gomonkey.ApplyFunc(k8s.GetClient, func() (*k8s.Client, error) {
			return &k8s.Client{}, nil
		})
		defer patch.Reset()
		convey.Convey("test normal case", func() {
			registerHandlers(filterClusterJob, clusterJobHandler)
			initCMInformer()
		})
	})
}

func TestGenericHandler(t *testing.T) {
	var f = func(*slownode.NodeAlgoResult, *slownode.NodeAlgoResult, watch.EventType) {}
	funcList := []func(*slownode.NodeAlgoResult, *slownode.NodeAlgoResult, watch.EventType){f}
	convey.Convey("Test genericHandler and handler", t, func() {
		newObj := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Data: map[string]string{
				constants.ClusterJobCMKey: `{"jobName":"test-job"}`,
			},
		}
		convey.Convey("test genericHandler", func() {
			// oldObj and new obj  is nil -> parse failed
			genericHandler(nil, nil, watch.Added, constants.ClusterJobCMKey, funcList)
			// oldObj is nil, newObj is valid
			genericHandler(nil, newObj, watch.Added, constants.ClusterJobCMKey, funcList)
			// oldObj is not nil, but parse failed
			oldObj := &corev1.ConfigMap{}
			genericHandler(oldObj, newObj, watch.Added, constants.ClusterJobCMKey, funcList)
		})
		convey.Convey("test handler", func() {
			// clusterJobHandler
			clusterJobHandler(nil, newObj, watch.Added)
			// nodeJobHandler
			nodeJobHandler(nil, newObj, watch.Added)
			// nodeAlgoHandler
			nodeAlgoHandler(nil, newObj, watch.Added)
			// nodeDataProfilingHandler
			nodeDataProfilingHandler(nil, newObj, watch.Added)
			// nodeRestartInfoHandler
			nodeRestartInfoHandler(nil, newObj, watch.Added)
		})
	})
}

func TestNodeRestartInfo(t *testing.T) {
	convey.Convey("test node restart info", t, func() {
		// create cm key
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Data: map[string]string{
				constants.NodeRestartInfoCMKey: `127.0.0.1`,
			},
		}
		funcList := []func(*string, *string, watch.EventType){cluster.JobRestartProcessor}
		genericHandler(nil, cm, watch.Added, constants.NodeRestartInfoCMKey, funcList)
	})
}

func TestIsNameMatched(t *testing.T) {
	convey.Convey("test isNameMatched", t, func() {
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ras-feature-slownode-job-13344455",
				Namespace: "default",
			},
		}
		convey.Convey("test name matched", func() {
			matched := isNameMatched(cm, constants.ClusterJobPrefix)
			convey.So(matched, convey.ShouldBeTrue)
		})
		convey.Convey("test convert cm failed", func() {
			name := "ras-feature-slownode"
			matched := isNameMatched(name, constants.ClusterJobPrefix)
			convey.So(matched, convey.ShouldBeFalse)
		})
		convey.Convey("test name unmatch", func() {
			matched := isNameMatched(cm, constants.NodeJobPrefix)
			convey.So(matched, convey.ShouldBeFalse)
		})
		convey.Convey("test filter func", func() {
			// filter cluster job
			cm.Name = "slow-node-job-13344455"
			convey.So(filterClusterJob(cm), convey.ShouldBeFalse)
			cm.Name = "ras-feature-slownode-job-13344455"
			convey.So(filterClusterJob(cm), convey.ShouldBeTrue)
			// filter node job
			convey.So(filterNodeJob(cm), convey.ShouldBeFalse)
			cm.Name = "slow-node-job-13344455"
			convey.So(filterNodeJob(cm), convey.ShouldBeTrue)
			// filter algo result
			convey.So(filterAlgoResult(cm), convey.ShouldBeFalse)
			cm.Name = "slow-node-algo-result-13344455"
			convey.So(filterAlgoResult(cm), convey.ShouldBeTrue)
			// filter data profiling result
			convey.So(filterDataProfilingResult(cm), convey.ShouldBeFalse)
			cm.Name = "data-profiling-result-13344455"
			convey.So(filterDataProfilingResult(cm), convey.ShouldBeTrue)
		})
	})
}

func marshalData(data any) []byte {
	dataBuffer, err := json.Marshal(data)
	if err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]json marshal data failed: %v", err)
		return nil
	}
	return dataBuffer
}

func TestParseCMResult(t *testing.T) {
	convey.Convey("TestParseCMResult", t, func() {

		convey.Convey("obj is nil", func() {
			result := ""
			err := parseCMResult(nil, constants.ClusterJobCMKey, &result)
			convey.So(err.Error(), convey.ShouldEqual, "convert to ConfigMap object failed")
		})
		convey.Convey("obj without FeatConf key", func() {
			cm := &corev1.ConfigMap{}
			cm.Name = testCmName
			result := ""
			err := parseCMResult(cm, constants.ClusterJobCMKey, &result)
			convey.So(err.Error(), convey.ShouldEqual, "no cmKey[FeatConf] found")
		})
		convey.Convey("obj is valid", func() {
			cm := &corev1.ConfigMap{}
			cm.Name = testCmName
			job := slownode.Job{}
			job.SlowNode = 1
			cm.Data = map[string]string{}
			cm.Data[constants.ClusterJobCMKey] = string(marshalData(job))
			err := parseCMResult(cm, constants.ClusterJobCMKey, &job)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("parse string", func() {
			var ip = `127.0.0.1`
			cm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
					Labels:    map[string]string{constants.CmConsumer: constants.CmConsumerValue},
				},
				Data: map[string]string{
					constants.NodeRestartInfoCMKey: ip,
				},
			}
			var nodeIp string
			err := parseCMResult(cm, constants.NodeRestartInfoCMKey, &nodeIp)
			convey.So(err, convey.ShouldBeNil)
			convey.So(ip, convey.ShouldEqual, nodeIp)
		})
	})
}
