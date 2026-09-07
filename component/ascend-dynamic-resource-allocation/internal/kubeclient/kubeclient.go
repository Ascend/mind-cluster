/*
 * Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package kubeclient provides the k8s client used by the DRA driver for node
// operations, such as writing the component version annotation.
package kubeclient

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	patchType "k8s.io/apimachinery/pkg/types"
	coreclientset "k8s.io/client-go/kubernetes"

	"ascend-common/common-utils/hwlog"
	ver "ascend-common/common-utils/version"
	"ascend-common/devmanager/common"
	draFlags "ascend-dynamic-resource-allocation/internal/flags"
)

// retryTime is the retry count for patching the annotation onto the node.
const retryTime = 3

// ClientK8s wraps the core k8s clientset with the node name and implements
// the ver.AnnotationAdder interface for node annotation operations.
type ClientK8s struct {
	ClientSet coreclientset.Interface
	NodeName  string
}

// NewClientK8s builds a ClientK8s from the kube client config and node name.
func NewClientK8s(config *draFlags.KubeClientConfig, nodeName string) (*ClientK8s, error) {
	clientSets, err := config.NewClientSets()
	if err != nil {
		return nil, err
	}
	return &ClientK8s{ClientSet: clientSets.Core, NodeName: nodeName}, nil
}

// AddAnnotation patches the annotation onto the node with retries.
func (c *ClientK8s) AddAnnotation(key, value string) error {
	escapedKey := strings.ReplaceAll(key, "~", "~0")
	escapedKey = strings.ReplaceAll(escapedKey, "/", "~1")
	patchMap := map[string]string{
		"op":    common.AddOP,
		"path":  common.PathForAnnotations + escapedKey,
		"value": value,
	}
	patchMapByte, err := json.Marshal([]interface{}{patchMap})
	if err != nil {
		hwlog.RunLog.Errorf("marshal patchMap failed, err is %v", err)
		return err
	}
	for i := 0; i < retryTime; i++ {
		_, err = c.ClientSet.CoreV1().Nodes().Patch(context.TODO(), c.NodeName,
			patchType.JSONPatchType, patchMapByte, metav1.PatchOptions{})
		if err != nil {
			hwlog.RunLog.Errorf("patch node annotation failed, err is %v", err)
			time.Sleep(time.Second)
			continue
		}
		break
	}
	return err
}

// ReportVersion reports the component version to the node annotation
// (huawei.com/<componentName>.version) for version query and ClusterD aggregation.
func (c *ClientK8s) ReportVersion(info ver.Info, componentName string) error {
	return ver.ReportVersionToNodeAnnotation(c, info, componentName)
}
