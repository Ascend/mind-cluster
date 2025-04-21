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
	"ascend-operator/pkg/utils"
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

// New create ranktable generator
func New(job *v1.AscendJob) *RankTable {
	r := &RankTable{
		SuperPodList: make([]*SuperPod, 0),
	}

	r.spBlock = utils.GetSpBlock(job)
	r.BaseGenerator = common.NewBaseGenerator(job, rankTableVersion, r)
	return r
}

// GatherServerList gather server list
func (r *RankTable) GatherServerList() {
	if r.BaseGenerator.GetIsSoftStrategy() {
		r.GatherServerListForSoftStrategy()
		return
	}
	r.GatherServerListForHardStrategy()
}

// GatherServerListForHardStrategy gather server list for hard strategy
func (r *RankTable) GatherServerListForHardStrategy() {
	r.BaseGenerator.GatherServerList()
	r.SuperPodList = make([]*SuperPod, 0)
	for id, server := range r.ServerList {
		vid := utils.GetLogicSuperPodId(id, r.spBlock, len(r.ServerList[0].DeviceList))
		if len(r.SuperPodList) == vid {
			r.SuperPodList = append(r.SuperPodList, &SuperPod{
				SuperPodID: strconv.Itoa(vid),
				ServerList: make([]*Server, 0),
			})
		}
		r.SuperPodList[vid].ServerList = append(r.SuperPodList[vid].ServerList, &Server{ServerID: server.ServerID})
	}
}

// GatherServerListForSoftStrategy gather server list for soft strategy
func (r *RankTable) GatherServerListForSoftStrategy() {
	r.BaseGenerator.GatherServerList()
	tmpSuperPods := make(map[string]*SuperPod)
	for _, server := range r.ServerList {
		if tmpSuperPods[server.SuperPodRank] == nil {
			tmpSuperPods[server.SuperPodRank] = &SuperPod{}
		}
		tmpSuperPods[server.SuperPodRank].SuperPodID = server.SuperPodRank
		tmpSuperPods[server.SuperPodRank].ServerList = append(tmpSuperPods[server.SuperPodRank].ServerList,
			&Server{ServerID: server.ServerID})
	}
	r.SuperPodList = make([]*SuperPod, len(tmpSuperPods))
	for _, superPod := range tmpSuperPods {
		id, err := strconv.Atoi(superPod.SuperPodID)
		if err != nil || id >= len(r.SuperPodList) {
			continue
		}
		r.SuperPodList[id] = superPod
	}
}
