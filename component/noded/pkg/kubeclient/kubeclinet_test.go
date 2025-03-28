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

// Package config for k8s client test
package kubeclient

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/clientcmd"

	"ascend-common/common-utils/hwlog"
	"nodeD/pkg/common"
)

const testNodeName = "test-node-name"

var (
	testErr       = errors.New("test error")
	testK8sClient *ClientK8s
)

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		return
	}
	code := m.Run()
	fmt.Printf("exit_code = %v\n", code)
}

func setup() error {
	return initLog()
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

func generateRandomString(length int) string {
	const letters = "abcdefg"
	if length < 0 {
		return ""
	}
	bytes := make([]byte, length)
	for i := range bytes {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			panic(err)
		}
		bytes[i] = letters[index.Int64()]
	}
	return string(bytes)
}

func TestNewClientK8s(t *testing.T) {
	var patches = gomonkey.ApplyFuncReturn(clientcmd.BuildConfigFromFlags, nil, nil).
		ApplyFuncReturn(kubernetes.NewForConfig, &kubernetes.Clientset{}, nil).
		ApplyFuncReturn(os.Getenv, testNodeName)
	defer patches.Reset()
	convey.Convey("test func NewClientK8s success", t, testNewClientK8s)
	convey.Convey("test func NewClientK8s failed, build client config error", t, testNewClientErrClientCfg)
	convey.Convey("test func NewClientK8s failed, new for config error", t, testNewClientErrNewCfg)
	convey.Convey("test func NewClientK8s failed, get node name error", t, testNewClientErrGetNodeName)
}

func testNewClientK8s() {
	var err error
	testK8sClient, err = NewClientK8s()
	convey.So(testK8sClient.NodeName, convey.ShouldEqual, testNodeName)
	convey.So(testK8sClient.NodeInfoName, convey.ShouldEqual, common.NodeInfoCMNamePrefix+testNodeName)
	convey.So(err, convey.ShouldBeNil)
}

func testNewClientErrClientCfg() {
	var p1 = gomonkey.ApplyFuncReturn(clientcmd.BuildConfigFromFlags, nil, testErr)
	defer p1.Reset()
	k8sClient, err := NewClientK8s()
	convey.So(k8sClient, convey.ShouldBeNil)
	convey.So(err, convey.ShouldResemble, testErr)
}

func testNewClientErrNewCfg() {
	var p2 = gomonkey.ApplyFuncReturn(kubernetes.NewForConfig, nil, testErr)
	defer p2.Reset()
	k8sClient, err := NewClientK8s()
	convey.So(k8sClient, convey.ShouldBeNil)
	convey.So(err, convey.ShouldResemble, testErr)
}

func testNewClientErrGetNodeName() {
	const errNodeNameLen = 250
	output := []gomonkey.OutputCell{
		{Values: gomonkey.Params{""}},
		{Values: gomonkey.Params{generateRandomString(errNodeNameLen)}},
		{Values: gomonkey.Params{"wrong node name"}},
	}
	var p3 = gomonkey.ApplyFuncSeq(os.Getenv, output)
	defer p3.Reset()

	innerErr := fmt.Errorf("the env of 'NODE_NAME' must be set")
	expErr := fmt.Errorf("check node name failed, err is %v", innerErr)
	k8sClient, err := NewClientK8s()
	convey.So(k8sClient, convey.ShouldBeNil)
	convey.So(err, convey.ShouldResemble, expErr)

	innerErr = fmt.Errorf("node name length %d is bigger than k8s env max length %d",
		errNodeNameLen, common.KubeEnvMaxLength)
	expErr = fmt.Errorf("check node name failed, err is %v", innerErr)
	k8sClient, err = NewClientK8s()
	convey.So(k8sClient, convey.ShouldBeNil)
	convey.So(err, convey.ShouldResemble, expErr)

	innerErr = fmt.Errorf("node name %s is illegal", "wrong node name")
	expErr = fmt.Errorf("check node name failed, err is %v", innerErr)
	k8sClient, err = NewClientK8s()
	convey.So(k8sClient, convey.ShouldBeNil)
	convey.So(err, convey.ShouldResemble, expErr)
}

func TestClientK8s(t *testing.T) {
	testK8sClient = &ClientK8s{
		ClientSet:    fake.NewSimpleClientset(),
		NodeName:     testNodeName,
		NodeInfoName: common.NodeInfoCMNamePrefix + testNodeName,
	}
	convey.Convey("test ClientK8s method 'CreateConfigMap'", t, testCreateConfigMap)
	convey.Convey("test ClientK8s method 'GetConfigMap'", t, testGetConfigMap)
	convey.Convey("test ClientK8s method 'UpdateConfigMap'", t, testUpdateConfigMap)
	convey.Convey("test ClientK8s method 'CreateOrUpdateConfigMap'", t, testCreateOrUpdateCM)

	convey.Convey("test ClientK8s method 'AddAnnotation'", t, testAddAnnotation)
}

func testAddAnnotation() {
	if testK8sClient == nil {
		panic("testK8sClient is nil")
	}
	err := testK8sClient.AddAnnotation("testKey", "testValue")
	convey.So(err.Error(), convey.ShouldEqual, `nodes "`+testNodeName+`" not found`)
}

func testCreateConfigMap() {
	if testK8sClient == nil {
		panic("testK8sClient is nil")
	}
	_, err := testK8sClient.CreateConfigMap(&v1.ConfigMap{})
	convey.So(err, convey.ShouldBeNil)
}

func testGetConfigMap() {
	if testK8sClient == nil {
		panic("testK8sClient is nil")
	}
	_, err := testK8sClient.GetConfigMap("", "")
	convey.So(err, convey.ShouldBeNil)
}

func testUpdateConfigMap() {
	if testK8sClient == nil {
		panic("testK8sClient is nil")
	}
	_, err := testK8sClient.UpdateConfigMap(&v1.ConfigMap{})
	convey.So(err, convey.ShouldBeNil)
}

func testCreateOrUpdateCM() {
	if testK8sClient == nil {
		panic("testK8sClient is nil")
	}

	convey.Convey("test method CreateOrUpdateConfigMap success", func() {
		err := testK8sClient.CreateOrUpdateConfigMap(&v1.ConfigMap{})
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test method CreateOrUpdateConfigMap success, cm not found", func() {
		// create success
		err := testK8sClient.CreateOrUpdateConfigMap(&v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-name-1",
				Namespace: "test-namespace-1",
			},
			Data: nil,
		})
		convey.So(err, convey.ShouldBeNil)

		// create error
		var p1 = gomonkey.ApplyMethodReturn(&ClientK8s{}, "CreateConfigMap", nil, testErr)
		defer p1.Reset()
		err = testK8sClient.CreateOrUpdateConfigMap(&v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-name-2",
				Namespace: "test-namespace-2",
			},
			Data: nil,
		})
		convey.So(err, convey.ShouldResemble, fmt.Errorf("can not create config map, err is %v", testErr))
	})

	convey.Convey("test method CreateOrUpdateConfigMap failed, update and create error", func() {
		var p2 = gomonkey.ApplyMethodReturn(&ClientK8s{}, "UpdateConfigMap", nil, testErr)
		defer p2.Reset()
		err := testK8sClient.CreateOrUpdateConfigMap(&v1.ConfigMap{})
		expErr := fmt.Errorf("update config map failed, err is %v", testErr)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}
