/*
Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package v1dot2 is using for v1.2 Ranktable.
*/
package v1dot2

import (
	"strconv"

	"ascend-operator/pkg/api/v1"
	"ascend-operator/pkg/ranktable/common"
)

const (
	rankTableVersion = "1.2"
)

// RankTable ranktable of v1.2
type RankTable struct {
	*common.BaseGenerator
	SuperPodList []*SuperPod `json:"super_pod_list"`
	spBlock      int
}

// SuperPod superpod of v1.2 ranktable
type SuperPod struct {
	SuperPodID string    `json:"super_pod_id"`
	ServerList []*Server `json:"server_list"`
}

// Server server of v1.2 ranktable
type Server struct {
	ServerID string `json:"server_id"`
}

const (
	// AnnoKeyOfSuperPod annotation key of superpod
	AnnoKeyOfSuperPod = "sp-block"
)

// New create ranktable generator
func New(job *v1.AscendJob) *RankTable {
	r := &RankTable{
		SuperPodList: make([]*SuperPod, 0),
	}
	spBlockStr := job.Annotations[AnnoKeyOfSuperPod]
	spBlock, err := strconv.Atoi(spBlockStr)
	if err != nil {
		spBlock = 0
	}
	r.spBlock = spBlock
	r.BaseGenerator = common.NewBaseGenerator(job, rankTableVersion, r)
	return r
}

// GatherServerList gather server list
func (r *RankTable) GatherServerList() {
	r.BaseGenerator.GatherServerList()
	r.SuperPodList = make([]*SuperPod, 0)
	superPodNum := r.spBlock / len(r.ServerList[0].DeviceList)
	for id, server := range r.ServerList {
		vid := id / superPodNum
		if len(r.SuperPodList) == vid {
			r.SuperPodList = append(r.SuperPodList, &SuperPod{
				SuperPodID: strconv.Itoa(vid),
				ServerList: make([]*Server, 0),
			})
		}
		r.SuperPodList[vid].ServerList = append(r.SuperPodList[vid].ServerList, &Server{ServerID: server.ServerID})
	}
}
