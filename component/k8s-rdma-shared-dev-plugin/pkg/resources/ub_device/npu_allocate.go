/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package ub_device for ub device info
package ub_device

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apitypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	ascutils "ascend-common/common-utils/utils"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/types"
	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/utils"
)

const (
	podPredicateTimeAnnotation = "predicate-time"
	deviceStatusAnnotation     = "k8s.v1.cni.cncf.io/device-status"
	hostIPEnv                  = "HOST_IP"
	kubeletPortEnv             = "KUBELET_PORT"
	defaultKubeletPort         = "10250"
)

// kubeletClient queries the kubelet /pods endpoint (self-signed cert).
var (
	kubeletClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // #nosec
		},
	}
	npuIdPrefixes = []string{
		api.AscendMinuxPrefix,    // "npu-"
		api.Ascend910MinuxPrefix, // "Ascend910-"
	}
)

// parsePredicateTime returns the pod's predicate-time and whether it is valid.
func parsePredicateTime(pod *v1.Pod) (uint64, bool) {
	timeStr, ok := pod.Annotations[podPredicateTimeAnnotation]
	if !ok || len(timeStr) > 8 {
		return 0, false
	}
	predicateTime, err := strconv.ParseUint(timeStr, 10, 64)
	if err != nil {
		hwlog.RunLog.Errorf("parse predicate time failed: %v", err)
		return 0, false
	}
	return predicateTime, true
}

// podIsOlder reports whether pod a should be allocated before pod b.
func podIsOlder(a, b *v1.Pod) bool {
	pa, oka := parsePredicateTime(a)
	pb, okb := parsePredicateTime(b)
	switch {
	case oka && okb:
		return pa < pb
	case oka:
		return true
	case okb:
		return false
	default:
		if !a.CreationTimestamp.Equal(&b.CreationTimestamp) {
			return a.CreationTimestamp.Before(&b.CreationTimestamp)
		}
		return a.UID < b.UID
	}
}

// pendingNpuPods filters out already-allocated pods and sorts the rest oldest-first.
func pendingNpuPods(pods []v1.Pod) []*v1.Pod {
	var pending []*v1.Pod
	for i := range pods {
		if _, done := pods[i].Annotations[deviceStatusAnnotation]; done {
			continue
		}
		pending = append(pending, &pods[i])
	}
	sort.Slice(pending, func(i, j int) bool { return podIsOlder(pending[i], pending[j]) })
	return pending
}

// markPodAllocated records the allocated DPU IDs in the device-status annotation.
func (rs *ubResourceServer) markPodAllocated(ctx context.Context, pod *v1.Pod, deviceIDs []string) error {
	containerName := podResourceContainerName(pod, rs.resourceName)
	if containerName == "" {
		return fmt.Errorf("no container of pod %s requests resource %s", pod.Name, rs.resourceName)
	}
	status, err := json.Marshal(map[string][]string{containerName: deviceIDs})
	if err != nil {
		return fmt.Errorf("marshal device-status for pod %s failed: %v", pod.Name, err)
	}
	annotations := map[string]string{
		podPredicateTimeAnnotation: strconv.FormatUint(math.MaxUint64, 10),
		deviceStatusAnnotation:     string(status),
	}
	patchBytes, err := json.Marshal(map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": annotations,
		},
	})
	if err != nil {
		return fmt.Errorf("marshal annotation patch for pod %s failed: %v", pod.Name, err)
	}
	_, err = rs.k8sClient.CoreV1().Pods(pod.Namespace).Patch(ctx, pod.Name,
		apitypes.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("patch pod %s annotation failed: %v", pod.Name, err)
	}
	return nil
}

// podResourceContainerName returns the name of the container requesting the given resource.
func podResourceContainerName(pod *v1.Pod, resourceName string) string {
	for _, container := range pod.Spec.Containers {
		for resName := range container.Resources.Requests {
			if strings.HasPrefix(string(resName), resourceName) {
				return container.Name
			}
		}
	}
	return ""
}

// parseNpuIds parses a comma-separated list of NPU device ids, e.g. "npu-1,npu-2" -> [1,2].
func parseNpuIds(value string) ([]int, error) {
	if strings.TrimSpace(value) == "" {
		return []int{}, nil
	}
	var devices []int
	for _, part := range strings.Split(value, ",") {
		id, err := strconv.Atoi(stripNpuIdPrefix(strings.TrimSpace(part)))
		if err != nil {
			return nil, fmt.Errorf("invalid NPU device id %q in %q", part, value)
		}
		devices = append(devices, id)
	}
	sort.Slice(devices, func(i, j int) bool { return devices[i] < devices[j] })
	return ascutils.RemoveDuplicates(devices), nil
}

// stripNpuIdPrefix removes the known device prefix from a value, e.g. "npu-0" -> "0".
func stripNpuIdPrefix(value string) string {
	for _, prefix := range npuIdPrefixes {
		if strings.HasPrefix(value, prefix) {
			return strings.TrimPrefix(value, prefix)
		}
	}
	return value
}

// podRequestsResource returns true if any container of the pod requests the given resource
func podRequestsResource(pod *v1.Pod, resourceName string) bool {
	for _, container := range pod.Spec.Containers {
		for resName := range container.Resources.Requests {
			if strings.HasPrefix(string(resName), resourceName) {
				return true
			}
		}
	}
	return false
}

// getNpuPodsOnNode lists active pods on this node requesting the resource.
func getNpuPodsOnNode(nodeName, resourceName string) []v1.Pod {
	url, err := kubeletPodsURL()
	if err != nil {
		hwlog.RunLog.Errorf("build kubelet /pods url failed: %v", err)
		return nil
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		hwlog.RunLog.Errorf("create kubelet /pods request failed: %v", err)
		return nil
	}
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		hwlog.RunLog.Errorf("get in-cluster config for kubelet auth failed: %v", err)
		return nil
	}
	req.Header.Add("Authorization", "Bearer "+kubeConfig.BearerToken)
	resp, err := kubeletClient.Do(req)
	if err != nil {
		hwlog.RunLog.Errorf("query kubelet /pods failed: %v", err)
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		hwlog.RunLog.Errorf("kubelet /pods returned status %d", resp.StatusCode)
		return nil
	}
	var podList v1.PodList
	if err := json.NewDecoder(resp.Body).Decode(&podList); err != nil {
		hwlog.RunLog.Errorf("decode kubelet /pods response failed: %v", err)
		return nil
	}
	var result []v1.Pod
	for _, pod := range podList.Items {
		if pod.Spec.NodeName != nodeName ||
			pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
			continue
		}
		if podRequestsResource(&pod, resourceName) {
			result = append(result, pod)
		}
	}
	return result
}

// kubeletPodsURL builds the kubelet /pods URL.
func kubeletPodsURL() (string, error) {
	hostIP := os.Getenv(hostIPEnv)
	if hostIP == "" {
		return "", fmt.Errorf("env %s is not set", hostIPEnv)
	}
	port := os.Getenv(kubeletPortEnv)
	if port == "" {
		port = defaultKubeletPort
	}
	return "https://" + net.JoinHostPort(hostIP, port) + "/pods", nil
}

// allocateByNpu records the allocated DPU device IDs on the owning pod.
func (rs *ubResourceServer) allocateByNpu(ctx context.Context, deviceIDs []string) error {
	if rs.k8sClient == nil {
		return fmt.Errorf("k8s client is nil, cannot record NPU allocation")
	}

	nodeName, err := utils.GetNodeName()
	if err != nil {
		return fmt.Errorf("get node name failed: %v", err)
	}

	pods := getNpuPodsOnNode(nodeName, rs.resourceName)

	// Re-allocation: pod already holds these devices, skip.
	for i := range pods {
		annotated, ok := pods[i].Annotations[deviceStatusAnnotation]
		if ok && sameDeviceSet(annotated, deviceIDs) {
			hwlog.RunLog.Infof("pod %s already holds devices %v, skip re-allocation", pods[i].Name, deviceIDs)
			return nil
		}
	}

	pending := pendingNpuPods(pods)
	// Refuse to bind arbitrary DPUs while a pending pod lacks an NPU annotation.
	for _, pod := range pending {
		if _, err := getPodNpuIds(pod); err != nil {
			return fmt.Errorf("reject allocation: %v", err)
		}
	}
	for _, pod := range pending {
		if err := rs.markPodAllocated(ctx, pod, deviceIDs); err != nil {
			return fmt.Errorf("mark pod %s allocated failed: %v", pod.Name, err)
		}
		hwlog.RunLog.Infof("recorded allocated devices %v on pod %s", deviceIDs, pod.Name)
		return nil
	}
	// No pending pod and no pod owns these devices.
	hwlog.RunLog.Warnf("no pending pod on node %s to record allocation for devices %v", nodeName, deviceIDs)
	return nil
}

// sameDeviceSet reports whether the device-status annotation matches the given device IDs.
func sameDeviceSet(annotated string, deviceIDs []string) bool {
	setA := deviceStatusSet(annotated)
	if len(setA) != len(deviceIDs) {
		return false
	}
	for _, id := range deviceIDs {
		if _, ok := setA[id]; !ok {
			return false
		}
	}
	return true
}

// deviceStatusSet flattens the per-container device-status JSON into a device ID set.
func deviceStatusSet(annotated string) map[string]struct{} {
	var status map[string][]string
	if err := json.Unmarshal([]byte(annotated), &status); err != nil {
		return nil
	}
	set := make(map[string]struct{})
	for _, ids := range status {
		for _, id := range ids {
			set[id] = struct{}{}
		}
	}
	return set
}

// deviceSpecByID returns the RDMA device specs of the UB device with the given ID.
func (rs *ubResourceServer) deviceSpecByID(id string) ([]*pluginapi.DeviceSpec, bool) {
	for _, dev := range rs.ubDevices {
		if dev.GetName() == id {
			return dev.GetRdmaSpec(), true
		}
	}
	return nil, false
}

// getPodNpuIds returns the pod's NPU IDs from the huawei.com/npu annotation.
func getPodNpuIds(pod *v1.Pod) ([]int, error) {
	for _, key := range npuAnnotationKeys(pod) {
		value, ok := pod.Annotations[key]
		if !ok {
			continue
		}
		ids, err := parseNpuIds(value)
		if err != nil {
			hwlog.RunLog.Warnf("invalid NPU annotation %s=%q: %v", key, value, err)
			continue
		}
		if len(ids) > 0 {
			hwlog.RunLog.Infof("got NPU IDs %v from pod %s annotation %s", ids, pod.Name, key)
			return ids, nil
		}
	}
	return nil, fmt.Errorf("no NPU device annotation found for pod %s", pod.Name)
}

// npuAnnotationKeys returns the annotation key carrying the assigned NPU IDs: huawei.com/npu.
func npuAnnotationKeys(*v1.Pod) []string {
	return []string{api.HuaweiNPU}
}

// npuMappedDeviceIDs maps NPU IDs to DPU device IDs via npu-nic-mapping.json.
func (rs *ubResourceServer) npuMappedDeviceIDs(npuIds []int, podName string) ([]string, error) {
	var deviceIDs []string
	matchedNpu := make(map[int]bool)
	for _, npuId := range npuIds {
		if matchedNpu[npuId] {
			continue
		}
		nicNames, err := utils.GetNicNames(npuId)
		if err != nil {
			return nil, fmt.Errorf("get nic names for npu %d failed: %v", npuId, err)
		}
		if len(nicNames) == 0 {
			return nil, fmt.Errorf("no nic name mapped for npu %d", npuId)
		}
		dev := rs.findDeviceByIfName(nicNames[0])
		if dev == nil {
			return nil, fmt.Errorf("no UB device found for npu %d ifName %s", npuId, nicNames[0])
		}
		deviceIDs = append(deviceIDs, dev.GetName())
		matchedNpu[npuId] = true
		hwlog.RunLog.Infof("mapped npu %d to UB device %s (ifName %s) for pod %s", npuId, dev.GetName(), nicNames[0], podName)
	}
	return deviceIDs, nil
}

// findDeviceByIfName finds the UB device matching the given interface name
func (rs *ubResourceServer) findDeviceByIfName(ifName string) types.Device {
	for _, dev := range rs.ubDevices {
		if ubDev, ok := dev.(types.UbDevice); ok && ubDev.GetIfName() == ifName {
			return dev
		}
	}
	return nil
}

// preferredDeviceIDs computes the preferred device allocation for one container.
func (rs *ubResourceServer) preferredDeviceIDs(ctx context.Context, req *pluginapi.ContainerPreferredAllocationRequest) ([]string, error) {
	size := int(req.AllocationSize)
	if size <= 0 {
		return nil, nil
	}

	// NPU-mapped DPUs of the allocating pod, if determinable.
	var preferred []string
	if rs.k8sClient != nil {
		if nodeName, err := utils.GetNodeName(); err == nil {
			pods := pendingNpuPods(getNpuPodsOnNode(nodeName, rs.resourceName))
			available := stringSet(req.AvailableDeviceIDs)
			for _, pod := range pods {
				npuIds, err := getPodNpuIds(pod)
				if err != nil {
					return nil, err
				}
				ids, err := rs.npuMappedDeviceIDs(npuIds, pod.Name)
				if err != nil {
					return nil, err
				}
				if len(ids) == 0 || !setSubset(stringSet(ids), available) {
					continue
				}
				preferred = ids
				break
			}
		}
	}

	used := make(map[string]struct{})
	result := make([]string, 0, size)
	add := func(ids []string) {
		for _, id := range ids {
			if len(result) >= size {
				return
			}
			if _, ok := used[id]; ok {
				continue
			}
			used[id] = struct{}{}
			result = append(result, id)
		}
	}
	add(req.MustIncludeDeviceIDs)
	add(preferred)
	add(req.AvailableDeviceIDs)
	return result, nil
}

// stringSet converts a device ID slice into a set.
func stringSet(ids []string) map[string]struct{} {
	set := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if id != "" {
			set[id] = struct{}{}
		}
	}
	return set
}

// setSubset reports whether sub is a subset of super.
func setSubset(sub, super map[string]struct{}) bool {
	for id := range sub {
		if _, ok := super[id]; !ok {
			return false
		}
	}
	return true
}
