/*
Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package generator is interface of Ranktable generator.
*/
package generator

import (
	"github.com/looplab/fsm"
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
	SetStatus(utils.RankTableStatus)
	GetStatus() utils.RankTableStatus
	AddPod(*v1.Pod) error
	DeletePod(*v1.Pod) utils.RankTableStatus
	GatherServerList()
	ToString() (string, error)
	GetConfigmapExist() utils.ConfigmapCheck
	SetConfigmapExist(utils.ConfigmapCheck)
	GetFsm(string) *fsm.FSM
}
