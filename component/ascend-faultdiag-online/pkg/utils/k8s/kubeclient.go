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

// Package k8s for k8s client
package k8s

import (
	"context"
	"fmt"
	"os"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/utils/constants"
)

const (
	workerselector = "workerselector"
	workerNodeTag  = "dls-worker-node"
)

var (
	err    error
	once   sync.Once
	client *Client
)

// Client k8s client include node name and node info name
type Client struct {
	ClientSet kubernetes.Interface
}

// GetClientK8s get the singleton instance of ClientK8s
func GetClient() (*Client, error) {
	once.Do(func() {
		var clientCfg *rest.Config
		var clientSet *kubernetes.Clientset
		clientCfg, err = clientcmd.BuildConfigFromFlags("", "")
		if err != nil {
			return
		}
		clientSet, err = kubernetes.NewForConfig(clientCfg)
		if err != nil {
			return
		}
		client = &Client{
			ClientSet: clientSet,
		}
	})
	return client, err
}

// CreateConfigMap create a configmap
func (c *Client) CreateConfigMap(cm *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	if cm == nil {
		return nil, fmt.Errorf("param cm is nil")
	}
	cmNum, err := c.GetConfigMapNum()
	if err != nil {
		return nil, err
	}
	if cmNum >= constants.MaxConfigMapNum {
		return nil, fmt.Errorf("config map number: %d reaches the max number: %d", cmNum, constants.MaxConfigMapNum)
	}
	newCM, err := c.ClientSet.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).
		Create(context.TODO(), cm, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return newCM, nil
}

// GetConfigMap get config map by name and name space
func (c *Client) GetConfigMap(cmName, cmNamespace string) (*corev1.ConfigMap, error) {
	newCM, err := c.ClientSet.CoreV1().ConfigMaps(cmNamespace).Get(context.TODO(), cmName, metav1.GetOptions{
		ResourceVersion: "0",
	})
	if err != nil {
		return nil, err
	}
	return newCM, nil
}

// UpdateConfigMap update config map
func (c *Client) UpdateConfigMap(cm *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	if cm == nil {
		return nil, fmt.Errorf("param cm is nil")
	}
	newCM, err := c.ClientSet.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).
		Update(context.TODO(), cm, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return newCM, nil
}

// CreateOrUpdateConfigMap create config map when config map not found or update config map
func (c *Client) CreateOrUpdateConfigMap(cm *corev1.ConfigMap) error {
	_, err := c.UpdateConfigMap(cm)
	if err == nil {
		return nil
	}
	if errors.IsNotFound(err) {
		if _, err := c.CreateConfigMap(cm); err != nil {
			hwlog.RunLog.Errorf("create configmap err: %v", err)
			return fmt.Errorf("can not create config map, err is %v", err)
		}
		return nil
	}
	hwlog.RunLog.Errorf("update configmap err: %v", err)
	return fmt.Errorf("update config map failed, err is %v", err)
}

// DeleteConfigMap delete a cm by giving cm name and namespace
func (c *Client) DeleteConfigMap(cmName, cmNamespace string) error {
	err := c.ClientSet.CoreV1().ConfigMaps(cmNamespace).Delete(context.TODO(), cmName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	return nil
}

// GetWorkerNodesIPByLabel get all the ips of all the nodes
func (c *Client) GetWorkerNodesIPByLabel(labelName, lableValue string) ([]string, error) {
	nodes, err := c.ClientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var ips []string
	for _, node := range nodes.Items {
		labels := node.Labels
		if labels[workerselector] != workerNodeTag || labels[labelName] != lableValue {
			continue
		}
		for _, addr := range node.Status.Addresses {
			if addr.Type == corev1.NodeInternalIP || addr.Type == corev1.NodeExternalIP {
				ips = append(ips, addr.Address)
			}
		}
	}
	return ips, nil
}

// GetLabels get all the labels of current node
func (c *Client) GetLabels() (map[string]string, error) {
	envNodeName := "NODE_NAME"
	nodeName := os.Getenv(envNodeName)
	if nodeName == "" {
		return nil, fmt.Errorf("no %s found in env", envNodeName)
	}
	node, err := c.ClientSet.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return node.Labels, nil
}

// GetConfigMapNum get the current number of config map, include all namespaces
func (c *Client) GetConfigMapNum() (int, error) {
	cms, err := c.ClientSet.CoreV1().ConfigMaps("").List(context.TODO(), metav1.ListOptions{
		ResourceVersion: "0", // read from cache
		Limit:           1,   // only need total numbers, no need to get the content
	})
	if err != nil {
		return 0, err
	}
	return len(cms.Items), nil
}
