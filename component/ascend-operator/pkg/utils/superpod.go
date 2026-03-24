/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
*/

// Package utils is common utils
package utils

import (
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"ascend-operator/pkg/api/v1"
)

const (
	// AnnoKeyOfSuperPod annotation key of utils
	AnnoKeyOfSuperPod = "sp-block"
	// Level1BlockKey level1 key of affinity config
	Level1BlockKey = "level1"
	// KvPairNum number of key value pair
	KvPairNum = 2
)

const (
	// SuperPodEnvPath super pod env path
	SuperPodEnvPath = `metadata.annotations['super-pod-rank']`
	// SuperPodAffinity super pod affinity key
	SuperPodAffinity = "super-pod-affinity"
	// SoftStrategy soft strategy
	SoftStrategy = "soft"
	// HardStrategy hard strategy
	HardStrategy = "hard"
	// SuperPodRankAnno super pod rank annotation key
	SuperPodRankAnno = "super-pod-rank"
	// Chip2Node16Sp a3x16 super pod schedule policy
	Chip2Node16Sp = "chip2-node16-sp"
	// Chip2Node8Sp a3x8 super pod schedule policy
	Chip2Node8Sp = "chip2-node8-sp"
	// Multilevel multi network level schedule policy
	Multilevel = "multilevel"
)

// GetLogicSuperPodNodes Return the number of computational nodes contained in the logical utils
func getLogicSuperPodNodes(spBlock, chipsPerNode int) int {
	if spBlock < chipsPerNode {
		return 1
	}
	return spBlock / chipsPerNode
}

// GetLogicSuperPodId Return the logical utils ID
func GetLogicSuperPodId(podRank, spBlock, chipsPerNode int) int {
	if spBlock <= 0 || chipsPerNode <= 0 {
		return 0
	}
	return podRank / getLogicSuperPodNodes(spBlock, chipsPerNode)
}

// GetSpBlock get logic superpod sp-block value
func GetSpBlock(job *v1.AscendJob) int {
	if job == nil || job.Annotations == nil {
		return 0
	}

	spBlockStr := job.Annotations[AnnoKeyOfSuperPod]
	spBlock, err := strconv.Atoi(spBlockStr)
	if err != nil {
		spBlock = 0
	}

	if spBlock == 0 && IsMultiLevelJob(job) {
		spBlock = getSpBlockFromAffinityConfig(job)
	}
	return spBlock
}

func getSpBlockFromAffinityConfig(job *v1.AscendJob) int {
	if job == nil || job.Annotations == nil {
		return 0
	}
	affinityBlocks := getAffinityBlocks(job)
	if affinityBlocks == nil {
		return 0
	}
	level1NodeNum, ok := affinityBlocks[Level1BlockKey]
	if !ok {
		return 0
	}
	devicesPerPod := 0
	for _, spec := range job.Spec.ReplicaSpecs {
		if spec == nil {
			continue
		}
		devicesPerPod = getDevicesPerPod(spec.Template.Spec.Containers)
		if devicesPerPod != 0 {
			break
		}
	}
	return level1NodeNum * devicesPerPod
}

func getAffinityBlocks(job *v1.AscendJob) map[string]int {
	affinityStr, ok := job.Annotations[api.AffinityConfigAnnoKey]
	if !ok {
		hwlog.RunLog.Error("AffinityConfigAnnoKey not exist")
		return nil
	}
	splits := strings.Split(affinityStr, ",")
	affinityBlocks := make(map[string]int, len(splits))
	for _, split := range splits {
		configPair := strings.Split(split, "=")
		if len(configPair) != KvPairNum {
			hwlog.RunLog.Error("config pair is not correct")
			return nil
		}
		num, err := strconv.Atoi(configPair[1])
		if err != nil {
			hwlog.RunLog.Error("value of pair is not integer")
			return nil
		}
		affinityBlocks[configPair[0]] = num
	}
	return affinityBlocks
}

func getAllDevicesForJob(job *v1.AscendJob) int {
	if job == nil || job.Spec.ReplicaSpecs == nil {
		return 0
	}

	totalDevices := 0
	for _, spec := range job.Spec.ReplicaSpecs {
		if spec == nil || spec.Replicas == nil {
			continue
		}
		replicas := int(*spec.Replicas)

		devicesPerPod := getDevicesPerPod(spec.Template.Spec.Containers)
		totalDevices += replicas * devicesPerPod
	}
	return totalDevices
}

func getDevicesPerPod(containers []corev1.Container) int {
	devicesPerPod := 0
	for _, container := range containers {
		if quantity, ok := container.Resources.Requests[api.HuaweiAscend910]; ok {
			devices := quantity.Value()
			devicesPerPod = int(devices)
			break
		}
	}
	return devicesPerPod
}

// GetSpBlockNum get spblock num for job
func GetSpBlockNum(job *v1.AscendJob) int {
	if job == nil || job.Annotations == nil {
		return 0
	}

	allDevices := getAllDevicesForJob(job)
	spBlock := GetSpBlock(job)
	if spBlock == 0 {
		return 0
	}
	return allDevices / spBlock
}
