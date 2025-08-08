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

// Package kubeclient a series of k8s function
package kubeclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/rest"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/limiter"
)

const (
	// HostIPEnv represents the env for HOST_IP
	HostIPEnv = "HOST_IP"
	// KubeletPortEnv represents the env for KUBELET_PORT
	KubeletPortEnv = "KUBELET_PORT"
	// DefaultKubeletPort represents the default listening port for the kubelet service
	DefaultKubeletPort = "10250"
	// MaxPortIntValue represents the max value of port
	MaxPortIntValue = 65535
)

var (
	kubeConfig     *rest.Config = nil
	kubeConfigOnce sync.Once
)

func isValidPort(port string) error {
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return err
	}

	if portNum < 0 || portNum > MaxPortIntValue {
		return fmt.Errorf("port must be within the range of 0 to 65535")
	}
	return nil
}

// getKltPodsURL returns the URL used by kubelet to obtain pods information
func getKltPodsURL() (string, error) {
	envHostIP := os.Getenv(HostIPEnv)
	parseIP := net.ParseIP(envHostIP)
	if parseIP == nil {
		return "", fmt.Errorf("host ip is invalid")
	}
	hostIP := parseIP.String()
	kubeletPort := os.Getenv(KubeletPortEnv)
	if err := isValidPort(kubeletPort); err != nil {
		hwlog.RunLog.Debugf("kubelet port: %s is invalid, err: %s, use default port: %s",
			kubeletPort, err.Error(), DefaultKubeletPort)
		kubeletPort = DefaultKubeletPort
	}

	host := fmt.Sprintf("%s:%s", hostIP, kubeletPort)
	kltPodsURL := &url.URL{
		Scheme: "https",
		Host:   host,
		Path:   "/pods",
	}
	return kltPodsURL.String(), nil
}

// createKltPodsReqWithToken returns the http request that get Pods using kubelet with a token
func createKltPodsReqWithToken() (*http.Request, error) {
	kltPodsURL, err := getKltPodsURL()
	if err != nil {
		hwlog.RunLog.Errorf("get klt pods url failed: %v", err)
		return nil, err
	}
	req, err := http.NewRequest("GET", kltPodsURL, nil)
	if err != nil {
		hwlog.RunLog.Errorf("create http request failed: %v", err)
		return nil, err
	}
	kubeConfigOnce.Do(func() {
		kubeConfig, err = rest.InClusterConfig()
		if err != nil {
			hwlog.RunLog.Errorf("build kubeConfig err: %v", err)
		}
		hwlog.RunLog.Info("init kubeConfig success")
	})
	if kubeConfig == nil {
		return nil, fmt.Errorf("kubeConfig is nil")
	}
	req.Header.Add("Authorization", "Bearer "+kubeConfig.BearerToken)
	return req, nil
}

// getPodsByKltPort returns pods information obtained through the kubelet port
func (ki *ClientK8s) getPodsByKltPort() (*v1.PodList, error) {
	req, err := createKltPodsReqWithToken()
	if err != nil {
		hwlog.RunLog.Errorf("get kubelet http request failed: %v", err)
		return nil, err
	}
	resp, err := ki.KltClient.Do(req)
	if err != nil {
		hwlog.RunLog.Errorf("send kubelet http request failed: %v", err)
		return nil, err
	}
	defer func() {
		if err = resp.Body.Close(); err != nil {
			hwlog.RunLog.Errorf("close response body failed, err: %v", err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get kubelet http response failed, resp status code: %v is not %v",
			resp.StatusCode, http.StatusOK)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, limiter.DefaultDataLimit))
	if err != nil {
		if err == io.EOF {
			hwlog.RunLog.Errorf("kubelet http response size exceeds %d", limiter.DefaultDataLimit)
			return nil, fmt.Errorf("response size exceeds limit")
		}
		hwlog.RunLog.Errorf("read kubelet http response failed: %v", err)
		return nil, err
	}
	var podList v1.PodList
	if err = json.Unmarshal(body, &podList); err != nil {
		hwlog.RunLog.Errorf("unmarshal kubelet http response failed: %v, response body: %s", err, body)
		return nil, err
	}
	return &podList, nil
}
