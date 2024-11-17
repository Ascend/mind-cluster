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

/*
Package plugin is using for HuaWei Ascend pin affinity schedule frame.
*/

package plugin

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

func (sHandle *ScheduleHandler) InitTorNodeInfo(ssn *framework.Session) {
	if sHandle == nil || ssn == nil {
		klog.V(util.LogErrorLev).Infof("InitTorNodeInfo failed, err: %s", util.ArgumentError)
		return
	}
	sHandle.Tors = nil
	cm, err := util.GetConfigMapWithRetry(ssn.KubeClient(), util.DevInfoNameSpace, TorNodeCMName)
	if err != nil {
		if !errors.IsNotFound(err) {
			klog.V(util.LogWarningLev).Infof("Get Tor-Node configmap failed, err: %s", util.SafePrint(err))
		}
		return
	}

	torList := &TorList{
		serverIps: map[string]string{},
	}

	if err = torList.ParseFromString(cm.Data[TorInfoCMKey]); err != nil {
		klog.V(util.LogErrorLev).Infof("Unmarshal FaultNodes from cache failed %s", util.SafePrint(err))
		return
	}

	torList.SyncBySsnNodes(sHandle.Nodes)
	torList.SyncBySsnJobs(sHandle.Jobs)
	torList.initSeverMaps()
	// refresh every ssn
	sHandle.Tors = torList
}

// TorList tor info about nodes
type TorList struct {
	Version   string `json:"version"`
	TorCount  int    `json:"tor_count"`
	Tors      []*Tor `json:"server_list"`
	serverIps map[string]string
}

type Tor struct {
	FreeServerCount int       `json:"-"`
	Id              int       `json:"tor_id"`
	IP              string    `json:"tor_ip"`
	Servers         []*Server `json:"server"`
	Jobs            map[api.JobID]SchedulerJob
}

// TorListInfo information for the current plugin
type TorListInfo struct {
	Status      string       `json:"status"`
	Version     string       `json:"version"`
	ServerCount int          `json:"server_count"`
	TorCount    int          `json:"tor_count"`
	ServerList  []ServerList `json:"server_list"`
}

type ServerList struct {
	Id      int                      `json:"tor_id"`
	Servers []map[string]interface{} `json:"server"`
}

type Slice struct {
	Idle  int
	Id    int
	Nodes map[string]*Server
}

type Server struct {
	IsUsedByMulJob bool   `json:"-"`
	NodeRank       string `json:"-"`
	IP             string `json:"server_ip"`
	Count          int    `json:"npu_count"`
	SliceId        int    `json:"slice_id"`
	Jobs           map[api.JobID]SchedulerJob
	CurrentJob     api.JobID
	NPUNode
}

type Servers struct {
	Version     string `json:"version"`
	ServerCount int    `json:"server_count"`
	TorCount    int    `json:"tor_count"`
	ServerList  []*Tor `json:"server_list"`
}

func (tl *TorList) ParseFromString(info string) error {
	if tl == nil {
		klog.V(util.LogInfoLev).Infof("ParseFromString failed, err: %s", util.ArgumentError)
		return fmt.Errorf(util.ArgumentError)
	}
	return json.Unmarshal([]byte(info), tl)
}

func (tl *TorList) initSeverMaps() {
	if tl == nil || len(tl.Tors) == 0 {
		klog.V(util.LogInfoLev).Infof("initSeverMaps failed, err: %s", util.ArgumentError)
		return
	}
	for _, tor := range tl.Tors {
		for _, sr := range tor.Servers {
			tl.serverIps[sr.Name] = tor.IP
		}
	}
}

func (tl *TorList) SyncBySsnNodes(nodes map[string]NPUNode) {
	if tl == nil {
		klog.V(util.LogInfoLev).Infof("SyncBySsnNodes failed, err: %s", util.ArgumentError)
		return
	}
	for _, node := range nodes {
		if _, tNode := tl.GetTorAndServerByNodeIP(node.Address); tNode != nil {
			tNode.NPUNode = node
		}
	}
}

func (tl *TorList) SyncBySsnJobs(jobs map[api.JobID]SchedulerJob) {
	if tl == nil {
		klog.V(util.LogInfoLev).Infof("SyncBySsnJobs failed, err: %s", util.ArgumentError)
		return
	}
	for _, job := range jobs {
		tl.SyncByJob(job)
	}
}

func (tl *TorList) SyncByJob(job SchedulerJob) {
	if tl == nil {
		klog.V(util.LogInfoLev).Infof("SyncByJob failed, err: %s", util.ArgumentError)
		return
	}
	for _, task := range job.Tasks {
		if task.NodeName == "" {
			continue
		}

		tor, node := tl.GetTorAndServerByNodeName(task.NodeName)
		if node != nil {
			if node.Jobs == nil {
				node.Jobs = map[api.JobID]SchedulerJob{}
			}
			node.Jobs[job.Name] = job
			if tor.Jobs == nil {
				tor.Jobs = map[api.JobID]SchedulerJob{}
			}
			tor.Jobs[job.Name] = job
		}

	}
}

func (tl *TorList) GetTorAndServerByNodeIP(ip string) (*Tor, *Server) {
	if tl == nil {
		klog.V(util.LogInfoLev).Infof("GetTorAndServerByNodeIP failed: %s", util.ArgumentError)
		return nil, nil
	}
	for _, tor := range tl.Tors {
		for _, tNode := range tor.Servers {
			if tNode.IP == ip {
				return tor, tNode
			}
		}
	}
	return nil, nil
}

func (tl *TorList) GetTorAndServerByNodeName(name string) (*Tor, *Server) {
	if tl == nil {
		klog.V(util.LogInfoLev).Infof("GetTorAndServerByNodeName failed: %s", util.ArgumentError)
		return nil, nil
	}
	for _, tor := range tl.Tors {
		for _, tNode := range tor.Servers {
			if tNode.Name == name {
				return tor, tNode
			}
		}
	}
	return nil, nil
}

func (t *Tor) GetNodeByNodeName(name string) *Server {
	if t == nil {
		klog.V(util.LogInfoLev).Infof("GetNodeByNodeName failed: %s", util.ArgumentError)
		return nil
	}
	for _, tNode := range t.Servers {
		if tNode.Name == name {
			return tNode
		}
	}
	return nil
}

func (t *Tor) GetNodeByNodeIP(ip string) *Server {
	if t == nil {
		klog.V(util.LogInfoLev).Infof("GetNodeByNodeIP failed: %s", util.ArgumentError)
		return nil
	}
	for _, tNode := range t.Servers {
		if tNode.IP == ip {
			return tNode
		}
	}
	return nil
}

func (t *Tor) HasAcrossJob() bool {
	if t == nil {
		klog.V(util.LogInfoLev).Infof("HasAcrossJob failed: %s", util.ArgumentError)
		return false
	}
	for _, tNode := range t.Servers {
		if tNode.IsUsedByMulJob {
			return true
		}
	}
	for _, job := range t.Jobs {
		if job.Status != util.PodGroupRunning {
			continue
		}
		for _, task := range job.Tasks {
			if task.NodeName == "" {
				continue
			}
			if t.GetNodeByNodeName(task.NodeName) == nil {
				return true
			}
		}
	}
	return false
}

// getNetSliceId get net slice num by first server's SliceId
func getNetSliceId(servers []*Server) int {
	for _, server := range servers {
		if server == nil {
			continue
		}
		return server.SliceId
	}
	return -1
}
