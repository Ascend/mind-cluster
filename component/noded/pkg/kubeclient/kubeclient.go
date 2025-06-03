/* Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package kubeclient for k8s client
package kubeclient

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"nodeD/pkg/common"
)

const retryTime = 3

var k8sClient *ClientK8s

// ClientK8s k8s client include node name and node info name
type ClientK8s struct {
	ClientSet    kubernetes.Interface
	NodeName     string
	NodeInfoName string
}

// NewClientK8s create k8s client
func NewClientK8s() (*ClientK8s, error) {
	clientCfg, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		hwlog.RunLog.Errorf("build client config err: %v", err)
		return nil, err
	}

	client, err := kubernetes.NewForConfig(clientCfg)
	if err != nil {
		hwlog.RunLog.Errorf("get client err: %v", err)
		return nil, err
	}

	nodeName, err := GetNodeNameFromEnv()
	if err != nil {
		return nil, err
	}
	k8sClient = &ClientK8s{
		ClientSet:    client,
		NodeName:     nodeName,
		NodeInfoName: common.NodeInfoCMNamePrefix + nodeName,
	}
	return k8sClient, nil
}

// GetK8sClient get k8s client
func GetK8sClient() *ClientK8s {
	return k8sClient
}

// GetNodeNameFromEnv get node name from env
func GetNodeNameFromEnv() (string, error) {
	nodeName := os.Getenv(api.NodeNameEnv)
	if err := checkNodeName(nodeName); err != nil {
		return "", fmt.Errorf("check node name failed, err is %v", err)
	}
	return nodeName, nil
}

// checkNodeName check if node name is legal
func checkNodeName(nodeName string) error {
	if len(nodeName) == 0 {
		return fmt.Errorf("the env of 'NODE_NAME' must be set")
	}
	if len(nodeName) > common.KubeEnvMaxLength {
		return fmt.Errorf("node name length %d is bigger than k8s env max length %d",
			len(nodeName), common.KubeEnvMaxLength)
	}
	pattern := common.GetPattern()[common.RegexNodeNameKey]
	if match := pattern.MatchString(nodeName); !match {
		return fmt.Errorf("node name %s is illegal", nodeName)
	}
	return nil
}

// CreateConfigMap create a configmap
func (ck *ClientK8s) CreateConfigMap(cm *v1.ConfigMap) (*v1.ConfigMap, error) {
	if cm == nil {
		return nil, fmt.Errorf("param cm is nil")
	}
	newCM, err := ck.ClientSet.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).
		Create(context.TODO(), cm, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return newCM, nil
}

// GetConfigMap get config map by name and name space
func (ck *ClientK8s) GetConfigMap(cmName, cmNameSpace string) (*v1.ConfigMap, error) {
	newCM, err := ck.ClientSet.CoreV1().ConfigMaps(cmNameSpace).Get(context.TODO(), cmName, metav1.GetOptions{
		ResourceVersion: "0",
	})
	if err != nil {
		return nil, err
	}
	return newCM, nil
}

// UpdateConfigMap update config map
func (ck *ClientK8s) UpdateConfigMap(cm *v1.ConfigMap) (*v1.ConfigMap, error) {
	if cm == nil {
		return nil, fmt.Errorf("param cm is nil")
	}
	newCM, err := ck.ClientSet.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).
		Update(context.TODO(), cm, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return newCM, nil
}

// DeleteConfigMap delete configmap
func (ck *ClientK8s) DeleteConfigMap(cmNamespace, cmName string) error {
	return ck.ClientSet.CoreV1().ConfigMaps(cmNamespace).Delete(context.TODO(), cmName, metav1.DeleteOptions{})
}

// CreateOrUpdateConfigMap create config map when config map not found or update config map
func (ck *ClientK8s) CreateOrUpdateConfigMap(cm *v1.ConfigMap) error {
	_, err := ck.UpdateConfigMap(cm)
	if err == nil {
		return nil
	}
	if errors.IsNotFound(err) {
		if _, err := ck.CreateConfigMap(cm); err != nil {
			hwlog.RunLog.Errorf("create configmap err: %v", err)
			return fmt.Errorf("can not create config map, err is %v", err)
		}
		return nil
	}
	hwlog.RunLog.Errorf("update configmap err: %v", err)
	return fmt.Errorf("update config map failed, err is %v", err)

}

// AddAnnotation add annotation
func (ck *ClientK8s) AddAnnotation(key, value string) error {
	patchMap := map[string]string{
		"op":    "replace",
		"path":  "/metadata/annotations/" + key,
		"value": value,
	}
	patchMapByte, err := json.Marshal([]interface{}{patchMap})
	if err != nil {
		hwlog.RunLog.Errorf("marshal patchMap failed, err is %v", err)
		return err
	}
	for i := 0; i < retryTime; i++ {
		_, err = ck.ClientSet.CoreV1().Nodes().Patch(context.TODO(), ck.NodeName,
			types.JSONPatchType, patchMapByte, metav1.PatchOptions{})
		if err != nil {
			hwlog.RunLog.Errorf("patch node annotation failed, err is %v", err)
			time.Sleep(time.Second)
			continue
		}
		break
	}
	return err
}
