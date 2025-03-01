/*
Copyright(C)2020-2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package util is using for the total variable.
*/
package util

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
	topInt := make([]int, 0)
	var cardStr string
	var topStrArray []string

	if topStr == "" {
		return []int{}
	}

	topStrArray = strings.Split(topStr, ",")
	for _, cardStr = range topStrArray {
		// cannot use strings 's Trim
		v := strings.TrimPrefix(cardStr, npuCardPreName)
		cardInt, err := strconv.Atoi(v)
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

// GetConfigurationByKey called by GetConfigFromSchedulerConfigMap
func GetConfigurationByKey(configurations []config.Configuration) map[string]string {
	for _, cf := range configurations {
		if cf.Name == CMInitParamKey {
			return cf.Arguments
		}
	}
	return map[string]string{}
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

// IsSliceContain judges whether keyword in targetSlice
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
	res := make([]int, 0)
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

// GetSizeOfSuperPod get size of super pod
func GetSizeOfSuperPod(configurations map[string]string) int {
	superPodSize := getSuperPodInfoFromConfig(sizeOfSuperPodKey, configurations)
	if superPodSize == 0 {
		klog.V(LogWarningLev).Infof(" super-pod-size configuration should be a number bigger than 0, "+
			"set default super-pod-size: %d", defaultSuperPodSize)
		superPodSize = defaultSuperPodSize
	}
	return superPodSize
}

// GetReserveNodes get reserve nodes
func GetReserveNodes(configurations map[string]string, superPodSize int) int {
	reserve := getSuperPodInfoFromConfig(reserveNodesKey, configurations)
	if reserve == 0 {
		reserve = defaultReserveNodes
	}
	if reserve >= superPodSize {
		validRes := 0
		if superPodSize > defaultReserveNodes {
			validRes = defaultReserveNodes
		}
		klog.V(LogWarningLev).Infof("reserve-nodes(%d) is larger than super-pod-size(%d), set reserve-nodes: %d",
			reserve, superPodSize, validRes)
		reserve = validRes
	}
	return reserve
}

func getSuperPodInfoFromConfig(key string, configurations map[string]string) int {
	if len(configurations) == 0 {
		klog.V(LogWarningLev).Info("volcano scheduler config init-params map is nil")
		return 0
	}
	value, ok := configurations[key]
	if !ok {
		klog.V(LogWarningLev).Infof("%s configuration not exist", key)
		return 0
	}

	res, err := strconv.Atoi(value)
	if err != nil {
		klog.V(LogWarningLev).Infof("cannot convert %s configuration, err: %v", key, err)
		return 0
	}
	if res < 0 {
		klog.V(LogWarningLev).Infof(" %s configuration should not be negative number", key)
		return 0
	}
	return res
}

// checkGraceDeleteTimeValid used by GetGraceDeleteTime for validity checking
func checkGraceDeleteTimeValid(overTime int64) bool {
	if overTime < minGraceOverTime || overTime > maxGraceOverTime {
		klog.V(LogErrorLev).Infof("GraceOverTime value should be range [2, 3600], configured is [%d], "+
			"GraceOverTime will not be changed", overTime)
		return false
	}
	// use user's configuration to set grace over time
	klog.V(LogInfoLev).Infof("set GraceOverTime to new value [%d].", overTime)
	return true
}

// GetGraceDeleteTime get grace delete time
func GetGraceDeleteTime(conf map[string]string) int64 {
	klog.V(LogInfoLev).Infof("enter GetGraceDeleteTime ...")
	defer klog.V(LogInfoLev).Infof("leave GetGraceDeleteTime ...")
	if len(conf) == 0 {
		klog.V(LogErrorLev).Infof("GetGraceDeleteTime failed: %s, no conf", ArgumentError)
		return DefaultGraceOverTime
	}
	// get grace over time by user configuration
	overTimeStr, ok := conf[GraceOverTimeKey]
	if !ok {
		klog.V(LogErrorLev).Info("set GraceOverTime failed and will not be changed, " +
			"key grace-over-time doesn't exists.")
		return DefaultGraceOverTime
	}
	overTime, err := strconv.ParseInt(overTimeStr, Base10, BitSize64)
	if err != nil {
		klog.V(LogErrorLev).Infof("set GraceOverTime failed and will not be changed, "+
			"grace-over-time is invalid [%s].", SafePrint(overTimeStr))
		return DefaultGraceOverTime
	}
	// check time validity
	if !checkGraceDeleteTimeValid(overTime) {
		return DefaultGraceOverTime
	}
	return overTime
}

// GetUseClusterDConfig check use cluster info manager by config, default true
func GetUseClusterDConfig(conf map[string]string) bool {
	useClusterInfoManager, ok := conf[UseClusterInfoManager]
	if !ok {
		klog.V(LogDebugLev).Info("CheckUseCIMByConfig doesn't exist useClusterInfoManager.")
		return true
	}
	return useClusterInfoManager == "true"
}

// GetPresetVirtualDeviceConfig get VNPU segmentEnable by init plugin parameters, return true if static
func GetPresetVirtualDeviceConfig(conf map[string]string) bool {
	// get segmentEnable by user configuration
	segmentEnable, ok := conf[SegmentEnable]
	if !ok {
		klog.V(LogDebugLev).Info("checkVNPUSegmentEnable doesn't exist presetVirtualDevice.")
		return false
	}
	return segmentEnable == "true"
}

// GetShardTorNum get shared tor num from configmap
func GetShardTorNum(conf map[string]string) int {
	str := conf[keyOfSharedTorNum]
	sharedTorNum, err := strconv.Atoi(str)
	if err != nil {
		klog.V(LogWarningLev).Infof("getSharedTorNum %s.", err)
		return shareTorNum2
	}
	if sharedTorNum != shareTorNum1 && sharedTorNum != shareTorNum2 {
		klog.V(LogWarningLev).Infof("sharedTorNum is illegal. use default config")
		return shareTorNum2
	}
	return sharedTorNum
}

// GetNslbVersion get nslb version from config
func GetNslbVersion(conf map[string]string) string {
	nslbVersion := conf[keyOfNSLBVersion]
	if nslbVersion != defaultNSLBVersion && nslbVersion != NSLB2Version {
		klog.V(LogWarningLev).Infof("nslbVersion is illegal. use default config")
		return defaultNSLBVersion
	}
	return nslbVersion
}

// CheckStrInSlice return whether str in string slice
func CheckStrInSlice(str string, slice []string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

// DeepCopyCmData return a replica of the cmDate
func DeepCopyCmData(cmData map[string]string) map[string]string {
	newCmData := make(map[string]string, len(cmData))
	for k, v := range cmData {
		newCmData[k] = v
	}
	return newCmData
}

// IsNodeReady returns the node ready status
func IsNodeReady(node *v1.Node) bool {
	for _, cond := range node.Status.Conditions {
		if cond.Type == v1.NodeReady {
			return cond.Status == v1.ConditionTrue
		}
	}
	return false
}

// MakeDataHash check code for configmap
func MakeDataHash(data interface{}) string {
	var dataBuffer []byte
	if dataBuffer = marshalData(data); len(dataBuffer) == 0 {
		return ""
	}
	h := sha256.New()
	if _, err := h.Write(dataBuffer); err != nil {
		klog.V(LogErrorLev).Infof("hash data error")
		return ""
	}
	sum := h.Sum(nil)
	return hex.EncodeToString(sum)
}

func marshalData(data interface{}) []byte {
	dataBuffer, err := json.Marshal(data)
	if err != nil {
		klog.V(LogErrorLev).Infof("marshal data err: %s", SafePrint(err))
		return nil
	}
	return dataBuffer
}
