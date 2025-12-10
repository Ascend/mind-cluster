/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
*/

// Package v2dot0 is using for v2.0 Ranktable.
package v2dot0

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	corev1 "k8s.io/api/core/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-operator/pkg/api/v1"
	"ascend-operator/pkg/ranktable/common"
	"ascend-operator/pkg/ranktable/utils"
)

const (
	rankTableVersion = "2.0"
)

const (
	// level0 for ranktable
	level0 = 0
	// level1 for ranktable
	level1 = 1
	// level2 for ranktable
	level2 = 2
	// level3 for ranktable
	level3 = 3
	// index of the first device in the devices list
	firstDevice = 0
	// represents zero length or empty value check
	zero = 0
)

// NetInfo for net info of the item in rank table
type NetInfo struct {
	PortAddrType string
	ScaleOutType string
	RankAddrType string
}

var defaultPriorityScaleOutType = []string{
	v1.PortAddrTypeRoCE,
	v1.PortAddrTypeUBoE,
	v1.PortAddrTypeUBG,
}

var portTypeMappings = map[string]NetInfo{
	v1.PortAddrTypeRoCE: {
		PortAddrType: v1.PortAddrTypeRoCE, ScaleOutType: v1.ScaleOutTypeRoCE, RankAddrType: v1.RankAddrTypeIP,
	},
	v1.PortAddrTypeUBoE: {
		PortAddrType: v1.PortAddrTypeUBoE, ScaleOutType: v1.ScaleOutTypeUBoE, RankAddrType: v1.RankAddrTypeIP,
	},
	v1.PortAddrTypeUBG: {
		PortAddrType: v1.PortAddrTypeUBG, ScaleOutType: v1.ScaleOutTypeUBoE, RankAddrType: v1.RankAddrTypeEID,
	},
}

var scaleOutTypeMappings = map[string][]NetInfo{
	v1.ScaleOutTypeRoCE: {
		{PortAddrType: v1.PortAddrTypeRoCE, ScaleOutType: v1.ScaleOutTypeRoCE, RankAddrType: v1.RankAddrTypeIP},
	},
	v1.ScaleOutTypeUBoE: {
		{PortAddrType: v1.PortAddrTypeUBoE, ScaleOutType: v1.ScaleOutTypeUBoE, RankAddrType: v1.RankAddrTypeIP},
		{PortAddrType: v1.PortAddrTypeUBG, ScaleOutType: v1.ScaleOutTypeUBoE, RankAddrType: v1.RankAddrTypeEID},
	},
}

// GetNetInfoByDefault get rank info generating policy by default priority
func (r *RankTable) GetNetInfoByDefault() (NetInfo, error) {
	for _, sot := range defaultPriorityScaleOutType {
		if _, exist := r.portAddrTypes.Load(sot); !exist {
			continue
		}

		if netInfo, exist := portTypeMappings[sot]; exist {
			return netInfo, nil
		}
	}
	// if no matching network type is found, log a warning and still generate the rank table with empty fields.
	hwlog.RunLog.Warn("no suitable port addr type found")
	return NetInfo{}, nil
}

// GetNetInfoByCustom get rank info generating policy by user custom
func (r *RankTable) GetNetInfoByCustom() (NetInfo, error) {
	infoList, exist := scaleOutTypeMappings[r.customScaleOutType]
	if !exist {
		errMsg := fmt.Sprintf("the value of label %s is invalid, which should be %s or %s",
			v1.ScaleOutTypeLabel, v1.ScaleOutTypeRoCE, v1.ScaleOutTypeUBoE)
		// In a priority scenario, if the scale-out type is invalid, log an error and skip generating the rank table.
		hwlog.RunLog.Error(errMsg)
		return NetInfo{}, errors.New(errMsg)
	}
	for _, item := range infoList {
		if _, ok := r.portAddrTypes.Load(item.PortAddrType); ok {
			return item, nil
		}
	}
	// if no matching network type is found, log a warning and still generate the rank table with empty fields.
	hwlog.RunLog.Warnf("no suitable port addr type found in device for the custom %v label value %s",
		v1.ScaleOutTypeLabel, r.customScaleOutType)
	return NetInfo{}, nil
}

// GetNetInfo get rank info generating policy
func (r *RankTable) GetNetInfo() (NetInfo, error) {
	if len(r.customScaleOutType) == 0 {
		return r.GetNetInfoByDefault()
	}
	return r.GetNetInfoByCustom()
}

// RankTable is the struct for rank table file
type RankTable struct {
	*common.BaseGenerator
	ranks     *sync.Map
	RankList  []*common.Rank `json:"rank_list,omitempty"`  // for latest 910A5 ranktable
	RankCount int            `json:"rank_count,omitempty"` // for latest 910A5 ranktable
	// portAddrTypes for collect the port addr type info of the rank table 910A5
	portAddrTypes      *sync.Map
	customScaleOutType string
}

// New construct the instance of RankTable
func New(job *v1.AscendJob) *RankTable {
	r := &RankTable{
		ranks:              &sync.Map{},
		RankList:           []*common.Rank{},
		portAddrTypes:      &sync.Map{},
		customScaleOutType: "",
	}
	r.BaseGenerator = common.NewBaseGenerator(job, rankTableVersion, r)
	if scaleOutType, ok := job.Labels[v1.ScaleOutTypeLabel]; ok {
		r.customScaleOutType = strings.ToUpper(strings.TrimSpace(scaleOutType))
	}

	// stacking server only applicable port type RoCE
	replicaSpec := job.Spec.ReplicaSpecs[v1.PytorchReplicaTypeMaster]
	if replicaSpec.Template.Spec.NodeSelector[api.AcceleratorTypeKey] == api.Ascend800ia5Stacking {
		if r.customScaleOutType == v1.PortAddrTypeUBoE {
			hwlog.RunLog.Warnf("job(%s) custom scale-out type is UBoE, but stacking servers only support RoCE",
				job.Name)
		}
		r.customScaleOutType = v1.PortAddrTypeRoCE
	}

	return r
}

// AddPod for add pod in rank table
func (r *RankTable) AddPod(pod *corev1.Pod) error {
	if pod == nil {
		hwlog.RunLog.Error("illegal input, pod is nil")
		return errors.New("illegal input, pod is nil")
	}
	deviceInfo, ok := pod.Annotations[api.Pod910DeviceAnno]
	if !ok {
		// Key does not exist, handle explicitly
		return fmt.Errorf("annotation %s not found in pod", api.Pod910DeviceAnno)
	}
	var instance common.Instance
	if err := json.Unmarshal([]byte(deviceInfo), &instance); err != nil {
		hwlog.RunLog.Errorf("unmarshal pod(%s/%s) deviceInfo(%s) failed: %v", pod.Namespace, pod.Name,
			deviceInfo, err)
		return err
	}
	if pod.Status.PodIP == "" {
		hwlog.RunLog.Errorf("pod(%s/%s) ip is empty", pod.Namespace, pod.Name)
		return fmt.Errorf("pod(%s/%s) ip is empty", pod.Namespace, pod.Name)
	}
	// a5 standard card no mesh scene, there is no urma device and eid info, should generate rank table level_list
	// Save the server_list value to baseGenerator
	if err := r.BaseGenerator.AddPod(pod); err != nil {
		hwlog.RunLog.Errorf("pod(%s/%s) add failed.Error: %s", pod.Namespace, pod.Name, err)
		return fmt.Errorf("pod(%s/%s) add failed.Error: %s", pod.Namespace, pod.Name, err)
	}

	return addPodInSuperPodScene(pod, r, instance)
}

func recordDevicePortAddrTypes(r *RankTable, dev common.Dev) {
	for _, rankLevel := range dev.LevelList {
		for netType := range rankLevel.Info {
			if _, exist := r.portAddrTypes.Load(netType); !exist {
				hwlog.RunLog.Infof("the rank table file contains device port addr type: %v", netType)
				r.portAddrTypes.Store(netType, struct{}{})
			}
		}
	}
}

func addPodInSuperPodScene(pod *corev1.Pod, r *RankTable, instance common.Instance) error {
	rankIndex, err := strconv.Atoi(pod.Annotations[api.PodRankIndexAnno])
	if err != nil {
		hwlog.RunLog.Errorf("parse pod(%s/%s) rankIndex(%s) failed: %v", pod.Namespace, pod.Name,
			pod.Annotations[api.PodRankIndexAnno], err)
		return err
	}

	devices := instance.Devices
	rankFactor := len(devices)
	podRank := make([]*common.Rank, 0, rankFactor)

	for _, dev := range devices {
		recordDevicePortAddrTypes(r, dev)
	}

	for index := range devices {
		var rank common.Rank
		if errTidy := GenRankList(&rank, &instance, index); errTidy != nil {
			hwlog.RunLog.Errorf("parse data failed: %v", errTidy)
			return errTidy
		}
		rank.RankID = rankIndex*rankFactor + index
		podRank = append(podRank, &rank)
	}

	if len(podRank) < 1 {
		return fmt.Errorf("%s/%s get device list failed", pod.Namespace, pod.Name)
	}
	r.ranks.Store(pod.UID, podRank)
	return nil
}

// shouldInclude determines whether the given level should be retained.
func shouldInclude(level int, portAddrType, customScaleOutType string) bool {
	// If portAddrType is empty and customScaleOutType is set, use customScaleOutType instead.
	if portAddrType == "" && customScaleOutType != "" {
		portAddrType = customScaleOutType
	}

	switch level {
	case level2:
		// Retain only if portAddrType is UBoE or UBG.
		return portAddrType == v1.PortAddrTypeUBoE || portAddrType == v1.PortAddrTypeUBG
	case level3:
		// Retain only if portAddrType is RoCE.
		return portAddrType == v1.PortAddrTypeRoCE
	default:
		// Retain by default.
		return true
	}
}

// getElement retrieves the element for a given level.
func getElement(curDevice common.Dev, level int, portAddrType string) (api.LevelElement, bool) {
	// Iterate through LevelList to find the layer where Level == target level.
	for _, lvl := range curDevice.LevelList {
		if lvl.Level != level {
			continue
		}

		// For level 0 and 1: directly return the first element.
		if level == level0 || level == level1 {
			for _, v := range lvl.Info {
				return v, true
			}
			return api.LevelElement{}, false
		}

		// For other levels: retrieve by portAddrType.
		elem, ok := lvl.Info[portAddrType]
		return elem, ok
	}

	return api.LevelElement{}, false
}

// GenRankList initializes and populates the rank structure with level0 and level1 information.
func GenRankList(rank *common.Rank, instance *common.Instance, index int) error {
	curDevice := instance.Devices[index]
	rank.Device = curDevice

	// Parse DeviceID.
	localID, err := strconv.Atoi(curDevice.DeviceID)
	if err != nil {
		hwlog.RunLog.Errorf("parse device id(%s) failed: %v", curDevice.DeviceID, err)
		return err
	}
	rank.LocalID = localID
	rank.DeviceID = localID

	// Process only level0 and level1.
	for level := level0; level <= level1; level++ {
		elem, ok := getElement(curDevice, level, "")
		if !ok {
			// Handle missing level.
			hwlog.RunLog.Warnf("device %s level=%d has no valid element, skip append",
				curDevice.DeviceID, level)
			continue
		}
		rank.LevelList = append(rank.LevelList, elem)
	}

	return nil
}

// DeletePod clears rank-related data.
func (r *RankTable) DeletePod() {
	r.ranks = &sync.Map{}
	r.portAddrTypes = &sync.Map{}
	r.SetStatus(utils.InitialRTStatus)
}

func (r *RankTable) GatherServerList() {
	r.RankList = make([]*common.Rank, 0)
	r.ranks.Range(func(key, value interface{}) bool {
		if ranks, ok := value.([]*common.Rank); ok {
			for _, rank := range ranks {
				r.RankList = append(r.RankList, rank)
			}
		} else {
			hwlog.RunLog.Warnf("unexpected type in ranks map for key=%v: %T", key, value)
		}
		return true
	})

	// Retrieve portAddrType.
	var portAddrType string
	if netInfo, err := r.GetNetInfo(); err != nil {
		hwlog.RunLog.Warnf("GetNetInfo failed in GatherServerList: %v", err)
		r.SetNeedGenerate(false)
		portAddrType = ""
	} else {
		portAddrType = netInfo.PortAddrType
		hwlog.RunLog.Infof("finally selected result in GatherServerList: scaleOutType=%s, portAddrType=%s",
			netInfo.ScaleOutType, portAddrType)
	}

	// Iterate through ranks and supplement level2 / level3.
	for _, rank := range r.RankList {
		curDevice := rank.Device

		for _, level := range []int{level2, level3} {
			if !shouldInclude(level, portAddrType, r.customScaleOutType) {
				continue
			}

			// Retrieve element; if not found, use default values.
			elem, ok := getElement(curDevice, level, portAddrType)
			if !ok {
				hwlog.RunLog.Warnf("device %s level=%d has no valid element, using empty default",
					curDevice.DeviceID, level)
				elem = api.LevelElement{
					NetLayer:      level,
					NetInstanceID: api.DefaultClusterName,
					NetType:       api.NetTypeCLOS,
					NetAttr:       api.NetAttrEmpty,
					RankAddrList:  []api.RankAddrItem{},
				}
			}
			rank.LevelList = append(rank.LevelList, elem)
		}
	}

	// Sort and count ranks.
	sort.Slice(r.RankList, func(i, j int) bool {
		iRankID := r.RankList[i].RankID
		jRankID := r.RankList[j].RankID
		return iRankID < jRankID
	})
	r.RankCount = len(r.RankList)
}
