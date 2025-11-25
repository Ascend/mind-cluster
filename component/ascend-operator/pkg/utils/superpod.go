/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
*/

// Package utils is common utils
package utils

import (
	"strconv"

	"ascend-operator/pkg/api/v1"
)

const (
	// AnnoKeyOfSuperPod annotation key of utils
	AnnoKeyOfSuperPod = "sp-block"
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
	// Chip2Node16Sp a3 super pod schedule policy
	Chip2Node16Sp = "chip2-node16-sp"
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
	return spBlock
}
