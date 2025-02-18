package kubeclient

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/limiter"
)

const (
	// HostIPEnv represents the env for HOST_IP
	HostIPEnv = "HOST_IP"
	// KubeletPortEnv represents the env for KUBELET_PORT
	KubeletPortEnv = "KUBELET_PORT"
	// KubeletPort represents the default listening port for the kubelet service
	DefaultKubeletPort = "10250"
	// MaxPortIntValue represents the max value of port
	MaxPortIntValue = 65535
)

func isValidPort(port string) error {
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("host ip is invalid")
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
		hwlog.RunLog.Errorf("host ip is invalid")
		return "", fmt.Errorf("host ip is invalid")
	}
	hostIP := parseIP.String()
	kubeletPort := os.Getenv(KubeletPortEnv)
	if err := isValidPort(kubeletPort); err != nil {
		hwlog.RunLog.Warnf("kubelet port:%s is not valid, use default port: %s",
			kubeletPort, DefaultKubeletPort)
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
	clientCfg, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		hwlog.RunLog.Errorf("build client config err: %v", err)
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+clientCfg.BearerToken)
	return req, nil
}

// getPodsByKltPort returns pods information obtained through the kubelet port
func getPodsByKltPort() (*v1.PodList, error) {
	req, err := createKltPodsReqWithToken()
	if err != nil {
		hwlog.RunLog.Errorf("get kubelet http request failed: %v", err)
		return nil, err
	}
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transport}
	resp, err := client.Do(req)
	if err != nil {
		hwlog.RunLog.Errorf("send kubelet http request failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get kubelet http response failed: %v", resp.StatusCode)
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
