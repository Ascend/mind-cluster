/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
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
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/component-helpers/node/util"

	"Ascend-device-plugin/pkg/common"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

const retryTime = 3

var defaultSafeCipherSuites = []uint16{
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
}

// ClientK8s include ClientK8sSet & nodeName & configmap name & kubelet http Client
type ClientK8s struct {
	Clientset      kubernetes.Interface
	NodeName       string
	DeviceInfoName string
	IsApiErr       bool
	PodInformer    cache.SharedIndexInformer
	Queue          workqueue.RateLimitingInterface
	KltClient      *http.Client
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

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			CipherSuites:       defaultSafeCipherSuites,
			MinVersion:         tls.VersionTLS13,
		},
	}
	kltClient := &http.Client{Transport: transport}

	return &ClientK8s{
		Clientset:      client,
		NodeName:       nodeName,
		DeviceInfoName: common.DeviceInfoCMNamePrefix + nodeName,
		Queue:          workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		IsApiErr:       false,
		KltClient:      kltClient,
	}, nil
}

// GetNode get node
func (ki *ClientK8s) GetNode() (*v1.Node, error) {
	return ki.Clientset.CoreV1().Nodes().Get(context.Background(), ki.NodeName, metav1.GetOptions{
		ResourceVersion: "0",
	})
}

// PatchNodeState patch node state
func (ki *ClientK8s) PatchNodeState(curNode, newNode *v1.Node) (*v1.Node, []byte, error) {
	node, patchBytes, err := util.PatchNodeStatus(ki.Clientset.CoreV1(), types.NodeName(ki.NodeName), curNode, newNode)
	if err != nil && strings.Contains(err.Error(), common.ApiServerPort) {
		ki.IsApiErr = true
	}

	return node, patchBytes, err
}

// AddAnnotation add annotation
func (ki *ClientK8s) AddAnnotation(key, value string) error {
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
		_, err = ki.Clientset.CoreV1().Nodes().Patch(context.TODO(), ki.NodeName,
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

// GetPod get pod by namespace and name
func (ki *ClientK8s) GetPod(pod *v1.Pod) (*v1.Pod, error) {
	if pod == nil {
		return nil, fmt.Errorf("param pod is nil")
	}

	v1Pod, err := ki.Clientset.CoreV1().Pods(pod.Namespace).Get(context.Background(), pod.Name, metav1.GetOptions{
		ResourceVersion: "0",
	})
	if err != nil && strings.Contains(err.Error(), common.ApiServerPort) {
		ki.IsApiErr = true
	}

	return v1Pod, err
}

// PatchPod patch pod information
func (ki *ClientK8s) PatchPod(pod *v1.Pod, data []byte) (*v1.Pod, error) {
	return ki.Clientset.CoreV1().Pods(pod.Namespace).Patch(context.Background(),
		pod.Name, types.StrategicMergePatchType, data, metav1.PatchOptions{})
}

// GetActivePodList is to get active pod list
func (ki *ClientK8s) GetActivePodList() ([]v1.Pod, error) {
	fieldSelector, err := fields.ParseSelector("spec.nodeName=" + ki.NodeName + "," +
		"status.phase!=" + string(v1.PodSucceeded) + ",status.phase!=" + string(v1.PodFailed))
	if err != nil {
		return nil, err
	}
	podList, err := ki.getPodListByCondition(fieldSelector)
	if err != nil {
		return nil, err
	}
	return checkPodList(podList)
}

// GetAllPodList get pod list by field selector
func (ki *ClientK8s) GetAllPodList() (*v1.PodList, error) {
	selector := fields.SelectorFromSet(fields.Set{"spec.nodeName": ki.NodeName})
	v1PodList, err := ki.getPodListByCondition(selector)
	if err != nil {
		hwlog.RunLog.Errorf("get pod list failed, err: %v", err)
		return nil, err
	}
	if len(v1PodList.Items) >= common.MaxPodLimit {
		hwlog.RunLog.Error("The number of pods exceeds the upper limit")
		return nil, fmt.Errorf("pod list count invalid")
	}
	return v1PodList, nil
}

// getPodListByCondition get pod list by field selector
func (ki *ClientK8s) getPodListByCondition(selector fields.Selector) (*v1.PodList, error) {
	newPodList, err := ki.Clientset.CoreV1().Pods(v1.NamespaceAll).List(context.Background(), metav1.ListOptions{
		FieldSelector:   selector.String(),
		ResourceVersion: "0",
	})
	if err != nil && strings.Contains(err.Error(), common.ApiServerPort) {
		ki.IsApiErr = true
	}

	return newPodList, err
}

// GetContainerRuntime get container runtime by node status info
func (ki *ClientK8s) GetContainerRuntime() (string, error) {
	node, err := ki.GetNode()
	if err != nil {
		return "", fmt.Errorf("failed to get node: %w", err)
	}

	runtimeVersion := node.Status.NodeInfo.ContainerRuntimeVersion
	if runtimeVersion == "" {
		return "", fmt.Errorf("container runtime version not found in node status")
	}
	if strings.HasPrefix(runtimeVersion, "docker://") {
		return DockerRuntime, nil
	}
	if strings.HasPrefix(runtimeVersion, "containerd://") {
		return ContainerdRuntime, nil
	}

	return "", fmt.Errorf("unknown container runtime: %s", runtimeVersion)
}

// checkPodList check each pod and return podList
func checkPodList(podList *v1.PodList) ([]v1.Pod, error) {
	if podList == nil {
		return nil, fmt.Errorf("pod list is invalid")
	}
	if len(podList.Items) >= common.MaxPodLimit {
		return nil, fmt.Errorf("the number of pods exceeds the upper limit")
	}
	var pods = make([]v1.Pod, 0)
	for _, pod := range podList.Items {
		if err := common.CheckPodNameAndSpace(pod.Name, common.PodNameMaxLength); err != nil {
			hwlog.RunLog.Warnf("pod name syntax illegal, err: %v", err)
			continue
		}
		if err := common.CheckPodNameAndSpace(pod.Namespace, common.PodNameSpaceMaxLength); err != nil {
			hwlog.RunLog.Warnf("pod namespace syntax illegal, err: %v", err)
			continue
		}
		pods = append(pods, pod)
	}
	return pods, nil
}

// CreateConfigMap create device info, which is cm
func (ki *ClientK8s) CreateConfigMap(cm *v1.ConfigMap) (*v1.ConfigMap, error) {
	if cm == nil {
		return nil, fmt.Errorf("param cm is nil")
	}

	newCM, err := ki.Clientset.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).
		Create(context.TODO(), cm, metav1.CreateOptions{})
	if err != nil && strings.Contains(err.Error(), common.ApiServerPort) {
		ki.IsApiErr = true
	}

	return newCM, err
}

// GetConfigMap get config map by name and namespace
func (ki *ClientK8s) GetConfigMap(cmName, cmNameSpace string) (*v1.ConfigMap, error) {
	newCM, err := ki.Clientset.CoreV1().ConfigMaps(cmNameSpace).Get(context.TODO(), cmName, metav1.GetOptions{
		ResourceVersion: "0",
	})
	if err != nil && strings.Contains(err.Error(), common.ApiServerPort) {
		ki.IsApiErr = true
	}

	return newCM, err
}

// UpdateConfigMap update device info, which is cm
func (ki *ClientK8s) UpdateConfigMap(cm *v1.ConfigMap) (*v1.ConfigMap, error) {
	if cm == nil {
		return nil, fmt.Errorf("param cm is nil")
	}
	newCM, err := ki.Clientset.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).
		Update(context.TODO(), cm, metav1.UpdateOptions{})
	if err != nil && strings.Contains(err.Error(), common.ApiServerPort) {
		ki.IsApiErr = true
	}

	return newCM, err
}

func (ki *ClientK8s) resetNodeAnnotations(node *v1.Node) {
	for k := range common.GetAllDeviceInfoTypeList() {
		delete(node.Annotations, k)
	}

	if common.ParamOption.AutoStowingDevs {
		delete(node.Labels, common.HuaweiRecoverAscend910)
		delete(node.Labels, common.HuaweiNetworkRecoverAscend910)
	}
}

// ResetDeviceInfo reset device info
func (ki *ClientK8s) ResetDeviceInfo() {
	nodeDeviceData := &common.NodeDeviceInfoCache{
		DeviceInfo:  common.NodeDeviceInfo{DeviceList: make(map[string]string, 1)},
		SuperPodID:  common.DefaultSuperPodID,
		ServerIndex: common.DefaultServerIndex,
	}
	if common.ParamOption.RealCardType == api.Ascend910A5 {
		var rackID int32 = common.DefaultRackID
		nodeDeviceData.RackID = &rackID
	}
	if err := ki.WriteDeviceInfoDataIntoCMCache(nodeDeviceData, "",
		common.GetSwitchFaultInfo(), common.DpuInfo{}, ""); err != nil {
		hwlog.RunLog.Errorf("write device info failed, error is %v", err)
	}
}

// ClearResetInfo clear reset info
func (ki *ClientK8s) ClearResetInfo(taskName, namespace string) error {
	taskInfo := &common.TaskResetInfo{
		RankList: make([]*common.TaskDevInfo, 0),
	}
	if _, err := ki.WriteResetInfoDataIntoCM(taskName, namespace, taskInfo, false); err != nil {
		hwlog.RunLog.Errorf("failed to clear reset info, err: %v", err)
		return err
	}
	return nil
}

// CreateEvent create event resource
func (ki *ClientK8s) CreateEvent(evt *v1.Event) (*v1.Event, error) {
	if evt == nil {
		return nil, fmt.Errorf("param event is nil")
	}
	return ki.Clientset.CoreV1().Events(evt.ObjectMeta.Namespace).Create(context.TODO(), evt, metav1.CreateOptions{})
}

// GetNodeNameFromEnv get current node name from env
func GetNodeNameFromEnv() (string, error) {
	nodeName := os.Getenv(api.NodeNameEnv)
	if err := checkNodeName(nodeName); err != nil {
		return "", fmt.Errorf("check node name failed: %v", err)
	}
	return nodeName, nil
}

func checkNodeName(nodeName string) error {
	if len(nodeName) == 0 {
		return fmt.Errorf("the env variable whose key is NODE_NAME must be set")
	}
	if len(nodeName) > common.KubeEnvMaxLength {
		return fmt.Errorf("node name length %d is bigger than %d", len(nodeName), common.KubeEnvMaxLength)
	}
	pattern := common.GetPattern()["nodeName"]
	if match := pattern.MatchString(nodeName); !match {
		return fmt.Errorf("node name %s is illegal", nodeName)
	}
	return nil
}

// ResourceEventHandler handle the configmap resource event
func (ki *ClientK8s) ResourceEventHandler(res ResourceType, filter func(obj interface{}) bool) cache.
	ResourceEventHandler {
	enqueue := func(obj interface{}, event EventType) {
		if res == PodResource && event == EventTypeUpdate {
			return
		}
		key, err := cache.MetaNamespaceKeyFunc(obj)
		if err != nil {
			hwlog.RunLog.Warnf("get key from obj failed, %v", err)
			return
		}
		hwlog.RunLog.Infof("%s %s(%s) to work queue", event, res, key)
		ki.Queue.AddRateLimited(Event{
			Resource: res,
			Key:      key,
			Type:     event,
		})
	}
	return cache.FilteringResourceEventHandler{
		FilterFunc: filter,
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				enqueue(obj, EventTypeAdd)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				if reflect.DeepEqual(oldObj, newObj) {
					return
				}
				enqueue(newObj, EventTypeUpdate)
			},
			DeleteFunc: func(obj interface{}) {
				enqueue(obj, EventTypeDelete)
			},
		},
	}
}

// FlushPodCacheNextQuerying next time querying pod, flush cache
func (ki *ClientK8s) FlushPodCacheNextQuerying() {
	ki.IsApiErr = true
}
