/* Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package pkg for noded
package pkg

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"huawei.com/npu-exporter/v5/common-utils/hwlog"
)

const (
	kubeEnvMaxLength = 253
	// DefaultHeartbeatInterval is 5 * time.Second
	DefaultHeartbeatInterval = 5
	// MaxHeartbeatInterval is 300 * time.Second
	MaxHeartbeatInterval = 300
)

// NewClientK8s create k8s client
func newClientK8s() (*kubernetes.Clientset, error) {
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

	return client, nil
}

// CheckNodeName check node name
func checkNodeName(nodeName string) error {
	if len(nodeName) > kubeEnvMaxLength {
		return fmt.Errorf("node name length %d is bigger than %d", len(nodeName), kubeEnvMaxLength)
	}
	pattern := `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`
	reg := regexp.MustCompile(pattern)
	if !reg.MatchString(nodeName) {
		return fmt.Errorf("node name is illegal")
	}
	return nil
}

// ValidHeartbeatInterval valid interval
func ValidHeartbeatInterval(interval int) error {
	if interval > MaxHeartbeatInterval || interval <= 0 {
		return fmt.Errorf("heartbeat interval id invalid")
	}
	return nil
}

// makeDataHash make data hash
func makeDataHash(data interface{}) string {
	dataBuffer, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	encode := sha256.New()
	if _, err := encode.Write(dataBuffer); err != nil {
		hwlog.RunLog.Errorf("hash data failed, err is %v", err)
		return ""
	}
	sum := encode.Sum(nil)
	return hex.EncodeToString(sum)
}
