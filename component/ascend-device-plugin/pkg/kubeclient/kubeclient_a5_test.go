/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package kubeclient a series of k8s function
package kubeclient

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func TestParseFaultNetworkInfoCM(t *testing.T) {
	convey.Convey("test ParseFaultNetworkInfoCM case 1 return err", t, func() {
		item := HccspingMeshItem{
			Activate: "123",
		}
		jsonData, err := json.Marshal(item)
		cm := &v1.ConfigMap{
			Data: map[string]string{"123": string(jsonData)},
		}
		_, err = ParseFaultNetworkInfoCM(cm)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestParseFaultNetworkInfoCMErrorCases(t *testing.T) {
	convey.Convey("test error input type", t, func() {
		errorTypeObj := "ErrorType"
		_, err := ParseFaultNetworkInfoCM(errorTypeObj)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "not fault network of ras feature configmap")
	})
	convey.Convey("test error unmarshal", t, func() {
		unMarshalPatch := gomonkey.ApplyFuncReturn(json.Unmarshal, fmt.Errorf("invalid input data"))
		defer unMarshalPatch.Reset()
		configMap := &v1.ConfigMap{}
		configMap.Data = make(map[string]string)
		configMap.Data["123"] = "123"
		_, err := ParseFaultNetworkInfoCM(configMap)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "unmarshal failed")
	})
}

func TestIsConfigMapExists(t *testing.T) {
	convey.Convey("test get ConfigMap error", t, func() {
		client := kubernetes.Clientset{}
		namespace := "default"
		name := "test"
		patch := gomonkey.ApplyMethodReturn(client.CoreV1().ConfigMaps(v1.NamespaceAll), "Get",
			&v1.ConfigMap{}, fmt.Errorf("configMap %s not found", name))
		defer patch.Reset()
		configMap, isExist := IsConfigMapExists(&client, namespace, name)
		convey.So(configMap, convey.ShouldBeNil)
		convey.So(isExist, convey.ShouldBeFalse)
	})
	convey.Convey("test get nil ConfigMap", t, func() {
		client := kubernetes.Clientset{}
		namespace := "default"
		name := "test"
		patch := gomonkey.ApplyMethodReturn(client.CoreV1().ConfigMaps(v1.NamespaceAll), "Get",
			nil, nil)
		defer patch.Reset()
		configMap, isExist := IsConfigMapExists(&client, namespace, name)
		convey.So(configMap, convey.ShouldBeNil)
		convey.So(isExist, convey.ShouldBeFalse)
	})
	convey.Convey("test get ConfigMap success", t, func() {
		client := kubernetes.Clientset{}
		namespace := "default"
		name := "test"
		patch := gomonkey.ApplyMethodReturn(client.CoreV1().ConfigMaps(v1.NamespaceAll), "Get",
			&v1.ConfigMap{}, nil)
		defer patch.Reset()
		configMap, isExist := IsConfigMapExists(&client, namespace, name)
		convey.So(configMap, convey.ShouldNotBeNil)
		convey.So(isExist, convey.ShouldBeTrue)
	})
}
