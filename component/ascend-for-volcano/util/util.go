/*
Copyright(C)2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package util is using for the total variable.
package util

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/config"
)

// ChangeTopToIntArray Change npu card ids from string to int array.
func ChangeTopToIntArray(topStr string, npuCardPreName string) []int {
	var topInt []int
	var cardInt int
	var cardStr string
	var err error
	var topStrArray []string

	if topStr == "" {
		return []int{}
	}

	topStrArray = strings.Split(topStr, ",")
	for _, cardStr = range topStrArray {
		// cannot use strings 's Trim
		v := strings.TrimPrefix(cardStr, npuCardPreName)
		cardInt, err = strconv.Atoi(v)
		if err != nil {
			klog.V(LogErrorLev).Infof("ChangeTopToIntArray conv failed %v.", err)
			return nil
		}

		topInt = append(topInt, cardInt)
	}
	klog.V(LogDebugLev).Infof("ChangeTopToIntArray %v.", topInt)
	return topInt
}

// IsMapHasNPUResource Determines whether a target string exists in the map.
func IsMapHasNPUResource(resMap map[v1.ResourceName]float64, npuName string) bool {
	for k := range resMap {
		// must contain "huawei.com"
		if strings.Contains(string(k), npuName) {
			return true
		}
	}
	return false
}

// ChangeIntArrToStr Covert []int to string. Like [0,1] -> "Ascend910-0,Ascend910-1".
func ChangeIntArrToStr(top []int, npuCardPreName string) string {
	var tmp int
	var str string

	i := 0
	for i, tmp = range top {
		str += npuCardPreName + strconv.Itoa(tmp)
		if i+1 < len(top) {
			str += ","
		}
	}

	return str
}

// GetConfigFromSchedulerConfigMap get config info from yaml
func GetConfigFromSchedulerConfigMap(configKey string, configurations []config.Configuration) (*config.Configuration,
	error) {
	if len(configurations) == 0 {
		return nil, errors.New("no configurations in scheduler configmap")
	}

	// in the new version, the configuration is obtained based on the configured name field.
	if cfg := getConfigurationByKey(configKey, configurations); cfg != nil {
		klog.V(LogDebugLev).Infof("get the configurations by name [%s] successful.", configKey)
		return cfg, nil
	}

	// compatible with old versions, because of the name field is not configured in the old versions.
	if configKey == CMSelectorKey {
		// if user removes configuration name and changes the order, will make mistakes.
		return getConfigurationOldVersion(configurations), nil
	}

	return nil, fmt.Errorf("cannot get configurations by name [%s], name not in configurations", configKey)
}

// getConfigurationByKey called by GetConfigFromSchedulerConfigMap
func getConfigurationByKey(configKey string, configurations []config.Configuration) *config.Configuration {
	for _, cf := range configurations {
		if cf.Name == configKey {
			return &cf
		}
	}

	return nil
}

// getConfigurationOldVersion called by GetConfigFromSchedulerConfigMap
func getConfigurationOldVersion(configurations []config.Configuration) *config.Configuration {
	// if user removes configuration name and changes the order, will make mistakes.
	klog.V(LogDebugLev).Info("compatible with old versions, get the selector configuration successful.")
	return &configurations[0]
}

// Max return the bigger one
func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// Min return the smaller one
func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// IsSliceContain judges whether keyword in tasgetSlice
func IsSliceContain(keyword interface{}, targetSlice interface{}) bool {
	if targetSlice == nil {
		klog.V(LogErrorLev).Infof("IsSliceContain :%s", ArgumentError)
		return false
	}
	kind := reflect.TypeOf(targetSlice).Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		klog.V(LogErrorLev).Infof(
			"the input %#v of type %T isn't a slice or array", targetSlice, targetSlice)
		return false
	}

	v := reflect.ValueOf(targetSlice)
	m := make(map[interface{}]struct{}, v.Len())
	for j := 0; j < v.Len(); j++ {
		m[v.Index(j).Interface()] = struct{}{}
	}

	_, ok := m[keyword]
	return ok
}

// RemoveSliceDuplicateElement remove duplicate element in slice
func RemoveSliceDuplicateElement(languages []string) []string {
	result := make([]string, 0, len(languages))
	temp := map[string]struct{}{}
	for _, item := range languages {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

// RemoveCommonElement remove common element from s1
func RemoveCommonElement(s1, s2 []int) []int {
	var res []int
	for _, e1 := range s1 {
		existFlag := false
		for _, e2 := range s2 {
			if e1 == e2 {
				existFlag = true
				break
			}
		}
		if !existFlag {
			res = append(res, e1)
		}
	}
	return res
}

// Add add resource
func (vResource *VResource) Add(resource VResource) {
	vResource.Aicore += resource.Aicore
	vResource.Aicpu += resource.Aicpu
}

// Sub sub resource
func (vResource *VResource) Sub(resource VResource) {
	vResource.Aicore -= resource.Aicore
	vResource.Aicpu -= resource.Aicpu
}

// BeGreater judge resource greater or equal to
func (vResource VResource) BeGreater(resource VResource) bool {
	return vResource.Aicore >= resource.Aicore && vResource.Aicpu >= resource.Aicpu
}

// ConvertErrSliceToError convert []error to one error.
func ConvertErrSliceToError(reErrors []error) error {
	var reE error

	for _, value := range reErrors {
		if reE == nil {
			reE = value
			continue
		}
		reE = fmt.Errorf("%s %s", reE, value)
	}

	return reE
}

// SafePrint safe print error
func SafePrint(args ...interface{}) string {
	msg := fmt.Sprint(args...)
	trimMsg := strings.Replace(msg, "\r", " ", -1)
	trimMsg = strings.Replace(trimMsg, "\n", " ", -1)
	return trimMsg
}

// ChangeNodesToNodeMaps change nodes slice into node maps
func ChangeNodesToNodeMaps(nodes []*api.NodeInfo) map[string]*api.NodeInfo {
	if len(nodes) == 0 {
		return nil
	}
	tmpNodes := make(map[string]*api.NodeInfo, len(nodes))
	for _, node := range nodes {
		tmpNodes[node.Name] = node
	}
	return tmpNodes
}

// GetNpuNameFromJobRequire get npuName,if job require name is npu-core return huawei.com/Ascend310P
func GetNpuNameFromJobRequire(npuName string) string {
	if npuName == AscendNPUCore {
		return NPU310PCardName
	}
	return npuName
}
