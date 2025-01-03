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
	"sort"
	"strconv"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog"
	"volcano.sh/volcano/pkg/scheduler/api"
	"volcano.sh/volcano/pkg/scheduler/framework"

	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/config"
	"volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/util"
)

// InitTorNodeInfo init tor node if basic tor node configmap exits
func (sHandle *ScheduleHandler) InitTorNodeInfo(ssn *framework.Session) {
	if sHandle == nil || ssn == nil {
		klog.V(util.LogErrorLev).Infof("InitTorNodeInfo failed, err: %s", util.ArgumentError)
		return
	}
	sHandle.Tors = nil
	cm, err := util.GetTorNodeWithOneMinuteDelay(ssn.KubeClient(), util.DevInfoNameSpace, TorNodeCMName)
	if err != nil {
		if !errors.IsNotFound(err) {
			klog.V(util.LogWarningLev).Infof("Get Tor-Node configmap failed, err: %s", util.SafePrint(err))
		}
		return
	}

	torList := &TorList{
		torIpMap:   map[string]string{},
		torMaps:    map[string]*Tor{},
		serverMaps: map[string]*Server{},
	}
	if err = json.Unmarshal([]byte(cm.Data[TorInfoCMKey]), torList); err != nil {
		klog.V(util.LogErrorLev).Infof("Unmarshal tor node from cache failed %s", util.SafePrint(err))
		return
	}
	if level, ok := cm.Data[TorLevelCMKey]; ok {
		torList.torLevel = level
		klog.V(util.LogInfoLev).Infof("basic tor level is %s", level)
	}
	if len(torList.Tors) == 0 {
		klog.V(util.LogDebugLev).Infof("basic tor node configmap is nil, stop tor node init")
		return
	}

	if sHandle.NslbAttr.nslbVersion == "" {
		torList.initParamFromConfig(sHandle.FrameAttr.Confs)
		sHandle.NslbAttr.nslbVersion = torList.nslbVersion
		sHandle.NslbAttr.sharedTorNum = torList.sharedTorNum
	} else {
		torList.nslbVersion = sHandle.NslbAttr.nslbVersion
		torList.sharedTorNum = sHandle.NslbAttr.sharedTorNum
	}

	torList.initNodeNameByNodeIp(sHandle.Nodes)
	torList.syncBySsnJobs(sHandle.Jobs)
	torList.initTorMaps()
	if torList.nslbVersion == NSLB2Version {
		torList.initTorShareStatus(sHandle.Jobs)
	}

	// refresh every ssn
	sHandle.Tors = torList
}

// TorList tor info about nodes
type TorList struct {
	sharedTorNum int
	nslbVersion  string
	torLevel     string
	Version      string `json:"version"`
	TorCount     int    `json:"tor_count"`
	Tors         []*Tor `json:"server_list"`
	torMaps      map[string]*Tor
	serverMaps   map[string]*Server
	torIpMap     map[string]string
}

// Tor tor info include server
type Tor struct {
	FreeServerCount int
	IsHealthy       int
	IsSharedTor     int
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

// ServerList server interface
type ServerList struct {
	Id      int                      `json:"tor_id"`
	Servers []map[string]interface{} `json:"server"`
}

// Slice include server
type Slice struct {
	Idle  int
	Id    int
	Nodes map[string]*Server
}

// Server server info
type Server struct {
	IsUsedByMulJob bool   `json:"-"`
	NodeRank       string `json:"-"`
	IP             string `json:"server_ip"`
	Count          int    `json:"npu_count"`
	SliceId        int    `json:"slice_id"`
	Jobs           map[api.JobID]SchedulerJob
	CurrentJob     *api.JobID
	Name           string
}

// Servers include basic tor
type Servers struct {
	Version     string      `json:"version"`
	ServerCount int         `json:"server_count"`
	TorCount    int         `json:"tor_count"`
	ServerList  []*basicTor `json:"server_list"`
}

type basicTor struct {
	IsHealthy   int
	IsSharedTor int
	Id          int            `json:"tor_id"`
	IP          string         `json:"tor_ip"`
	Servers     []*basicServer `json:"server"`
}

type basicServer struct {
	torIp          string
	IsUsedByMulJob bool   `json:"-"`
	NodeRank       string `json:"-"`
	IP             string `json:"server_ip"`
	Count          int    `json:"npu_count"`
	SliceId        int    `json:"slice_id"`
}

// TorShare tor share info
type TorShare struct {
	IsHealthy   int
	IsSharedTor int
	NodeJobs    []NodeJobInfo `json:"nodes"`
}

// NodeJobInfo node job info
type NodeJobInfo struct {
	NodeIp   string
	NodeName string
	JobName  []string
}

func (tl *TorList) initTorMaps() {
	for _, tor := range tl.Tors {
		tl.torMaps[tor.IP] = tor
		for _, server := range tor.Servers {
			tl.serverMaps[server.Name] = server
			tl.torIpMap[server.Name] = tor.IP
		}
	}
}

func (tl *TorList) initNodeNameByNodeIp(nodes map[string]NPUNode) {
	ipNodeMap := make(map[string]NPUNode, len(nodes))
	for _, node := range nodes {
		ipNodeMap[node.Address] = node
	}

	for _, tor := range tl.Tors {
		for _, tNode := range tor.Servers {
			if node, ok := ipNodeMap[tNode.IP]; ok {
				tNode.Name = node.Name
				continue
			}
			klog.V(util.LogDebugLev).Infof("tor node configmap info error, the ip : %s missing", tNode.IP)
		}
	}
}

func (tl *TorList) syncBySsnJobs(jobs map[api.JobID]SchedulerJob) {
	for _, job := range jobs {
		tl.syncByJob(job)
	}
}

func (tl *TorList) syncByJob(job SchedulerJob) {
	for _, task := range job.Tasks {
		if task.NodeName == "" {
			continue
		}

		tor, server := tl.getTorAndServerByNodeName(task.NodeName)
		if server == nil {
			continue
		}
		if server.Jobs == nil {
			server.Jobs = map[api.JobID]SchedulerJob{}
		}
		server.Jobs[job.Name] = job
		if tor.Jobs == nil {
			tor.Jobs = map[api.JobID]SchedulerJob{}
		}
		tor.Jobs[job.Name] = job
	}
}

func (tl *TorList) getTorAndServerByNodeName(nodeName string) (*Tor, *Server) {
	for _, tor := range tl.Tors {
		for _, tNode := range tor.Servers {
			if tNode.Name == nodeName {
				return tor, tNode
			}
		}
	}
	return nil, nil
}

func (tl *TorList) initParamFromConfig(Confs []config.Configuration) {
	tl.sharedTorNum = 0
	tl.nslbVersion = defaultNSLBVersion
	if len(Confs) == 0 {
		err := fmt.Errorf(util.ArgumentError)
		klog.V(util.LogWarningLev).Infof("getSharedTorNum %s. use default config", err)
		return
	}
	configuration, err := util.GetConfigFromSchedulerConfigMap(util.CMInitParamKey, Confs)
	if err != nil {
		klog.V(util.LogWarningLev).Infof("getSharedTorNum %s. use default config", err)
		return
	}
	str := configuration.Arguments[keyOfSharedTorNum]
	sharedTorNum, err := strconv.Atoi(str)
	if err != nil {
		klog.V(util.LogWarningLev).Infof("getSharedTorNum %s.", err)
		return
	}
	if sharedTorNum != shareTorNum1 && sharedTorNum != shareTorNum2 {
		klog.V(util.LogWarningLev).Infof("sharedTorNum is illegal. use default config")
		return
	}
	nslbVersion := configuration.Arguments[keyOfNSLBVersion]
	if nslbVersion != defaultNSLBVersion && nslbVersion != NSLB2Version {
		klog.V(util.LogWarningLev).Infof("nslbVersion is illegal. use default config")
		return
	}
	tl.nslbVersion = nslbVersion
	tl.sharedTorNum = sharedTorNum
	klog.V(util.LogWarningLev).Infof("nslbVersion and sharedTorNum init success.can not change the parameters and" +
		" it will not be changed during normal operation of the volcano")
}

func (tl *TorList) initTorShareStatus(jobs map[api.JobID]SchedulerJob) {
	for _, job := range jobs {
		if job.Status != util.PodGroupRunning && job.Status != util.PodGroupUnknown {
			continue
		}
		for _, task := range job.Tasks {
			if tor, ok := tl.torMaps[tl.torIpMap[task.NodeName]]; ok {
				tor.setTorIsSharedTor(task.Annotation[isSharedTor])
				tor.setTorIsHealthy(task.Annotation[isHealthy])
			}
		}
	}
}

func (t *Tor) setTorIsSharedTor(isShared string) {
	if t.IsSharedTor != freeTor {
		return
	}
	isSharedT, err := strconv.Atoi(isShared)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("setTorIsSharedTor err %s", err)
	}
	t.IsSharedTor = isSharedT
}

func (t *Tor) setTorIsHealthy(isHealthy string) {
	if t.IsHealthy == unhealthyTor {
		return
	}
	isHealthyT, err := strconv.Atoi(isHealthy)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("setTorIsHealthy err %s", err)
	}
	t.IsHealthy = isHealthyT
}

// GetNodeByNodeName get node by node name
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

// HasAcrossJob whether has across job
func (t *Tor) HasAcrossJob(isNSLBv2 bool, jobName api.JobID) bool {
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
		if !job.preCheckForTorHasAcrossJob(isNSLBv2, jobName) {
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

// IsUsedByAcrossLargeModelJob whether used by across large model job
func (t *Tor) IsUsedByAcrossLargeModelJob() bool {
	if t == nil {
		klog.V(util.LogInfoLev).Infof("IsUsedByAcrossLargeModelJob failed: %s", util.ArgumentError)
		return false
	}
	return t.HasAcrossJob(true, "")
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

// getTorServer get tors by tor is shared and is healthy
func getTorServer(tors []*Tor, isShare int, isHealthy int, sortType string) []*Tor {
	tmpTors := initTorsByTorAttr(tors, isShare, isHealthy)
	sort.Slice(tmpTors, func(i, j int) bool {
		if sortType == descOrder {
			return tmpTors[i].FreeServerCount > tmpTors[j].FreeServerCount
		}
		return tmpTors[i].FreeServerCount < tmpTors[j].FreeServerCount
	})
	return tmpTors
}

// initTorsByTorAttr init a tors by tor is shared and is healthy
func initTorsByTorAttr(tors []*Tor, isShare, isHealthy int) []*Tor {
	var tmpTors []*Tor
	for _, tor := range tors {
		if (tor.IsSharedTor == isShare || isShare == allTor) && (tor.IsHealthy == isHealthy || isHealthy == allTor) {
			tmpTors = append(tmpTors, tor)
		}
	}
	return tmpTors
}

// GetTorServer get all healthy tor
func GetTorServer(tors []*Tor, sortType string) []*Tor {
	return getTorServer(tors, allTor, healthyTor, sortType)
}

// GetSharedTorServer get healthy shared tors
func GetSharedTorServer(tors []*Tor, sortType string) []*Tor {
	return getTorServer(tors, sharedTor, healthyTor, sortType)
}

// GetMaxSharedTorServerNum get max shared tor num a job can use
func GetMaxSharedTorServerNum(tors []*Tor, sharedTorNum int) int {
	n := len(tors)
	if n == 0 || sharedTorNum <= 0 {
		return 0
	}
	if n == 1 {
		return tors[0].FreeServerCount
	}
	switch sharedTorNum {
	case oneTor:
		return tors[n-oneTor].FreeServerCount
	case twoTor:
		return tors[n-oneTor].FreeServerCount + tors[n-twoTor].FreeServerCount
	default:
		return 0
	}
}

// GetNotShareTorServer get the exclusiveTor tors
func GetNotShareTorServer(tors []*Tor, sortType string) []*Tor {
	return getTorServer(tors, exclusiveTor, healthyTor, sortType)

}

// GetNotShareAndFreeTorServer get the free tors
func GetNotShareAndFreeTorServer(tors []*Tor, sortType string) []*Tor {
	return getTorServer(tors, freeTor, healthyTor, sortType)
}

// GetLargeModelMaxServerNum get the node num that nslb 2.0 job can use at most
func GetLargeModelMaxServerNum(tors []*Tor, sharedTorNum int) int {
	var n int
	for _, tor := range GetNotShareAndFreeTorServer(tors, descOrder) {
		n += tor.FreeServerCount
	}
	return n + GetMaxSharedTorServerNum(GetSharedTorServer(tors, ascOrder), sharedTorNum)
}

// GetUnhealthyTorServer get unhealthy shared tors
func GetUnhealthyTorServer(tors []*Tor, sortType string) []*Tor {
	return getTorServer(tors, sharedTor, unhealthyTor, sortType)
}

// GetHealthyTorUsedByNormalJob get shared tors only used by normal job
func GetHealthyTorUsedByNormalJob(tors []*Tor, sortType string) []*Tor {
	var tmpTors []*Tor
	allShareTor := getTorServer(tors, sharedTor, healthyTor, sortType)
	for _, tor := range allShareTor {
		if !tor.IsUsedByAcrossLargeModelJob() {
			tmpTors = append(tmpTors, tor)
		}
	}
	return tmpTors
}

func initTempTor(tor *Tor, isShared, isHealthy int) *Tor {
	return &Tor{
		IsSharedTor: isShared,
		IsHealthy:   isHealthy,
		Id:          tor.Id,
		IP:          tor.IP,
	}
}

func getOneSharedTorServer(tors []*Tor, serverNum int) *Tor {
	if len(tors) == 0 {
		return nil
	}
	for _, tor := range tors {
		if tor.FreeServerCount < serverNum {
			continue
		}
		return tor
	}
	return nil
}

func getJobFreeServerNum(jobUid api.JobID, tors []*Tor) int {
	var num int
	for _, tor := range tors {
		for _, s := range tor.Servers {
			if s.CurrentJob != nil && *s.CurrentJob == jobUid {
				num++
			}
		}
	}
	return num
}

func copyTorList(t []*Tor) []*Tor {
	tmpTors := make([]*Tor, len(t))
	copy(tmpTors, t)
	return tmpTors
}

func deepCopyTorList(t []*Tor) []*Tor {
	str, err := json.Marshal(t)
	if err != nil {
		klog.V(util.LogErrorLev).Infof("deepCopyTorList Marshal %s", err)
		return nil
	}
	var tmpTor []*Tor
	if unMarErr := json.Unmarshal(str, &tmpTor); unMarErr != nil {
		klog.V(util.LogErrorLev).Infof("deepCopyTorList Unmarshal %s", unMarErr)
		return nil
	}
	return tmpTor
}

func initJobTorInfos() jobTorInfos {
	return jobTorInfos{
		torNums:        map[string]int{},
		usedAllTorNum:  0,
		usedHealthyTor: []*Tor{},
		otherTor:       []*Tor{},
	}
}
