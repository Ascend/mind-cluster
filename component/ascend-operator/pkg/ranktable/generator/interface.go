/*
Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package generator is interface of Ranktable generator.
*/
package generator

import (
	"k8s.io/api/core/v1"

	"ascend-operator/pkg/ranktable/utils"
)

// FileManager is used to write ranktable to file and delete.
type FileManager interface {
	WriteToFile() error
	DeleteFile() error
}

// RankTableGenerator is used to generate ranktable.
type RankTableGenerator interface {
	FileManager
	GetNeedGenerate() bool
	SetStatus(utils.RankTableStatus)
	GetStatus() utils.RankTableStatus
	AddPod(*v1.Pod) error
	DeletePod()
	GatherServerList()
	ToString() (string, error)
	GetConfigmapExist() utils.ConfigmapCheck
	SetConfigmapExist(utils.ConfigmapCheck)
	GetTimeStamp() uint64
	SetTimeStamp(uint64)
	GetConfigmapStatus() utils.RankTableStatus
	SetConfigmapStatus(utils.RankTableStatus)
	GetFileStatus() utils.RankTableStatus
	SetFileStatus(utils.RankTableStatus)
	Lock()
	Unlock()
}
