// Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
// package epranktable for
package epranktable

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"k8s.io/api/core/v1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/domain/pod"
	"clusterd/pkg/interface/kube"
)

const (
	// standaloneDeployServerKey is the label key for server num of standalone deploy mode
	standaloneDeployServerKey = "grt-server/deploy-server"
	// distributedDeployServerKey is the label key for server num of distributed deploy mode
	distributedDeployServerKey = "grt-group/deploy-server"
)

// RankTableStatus is rank table status
type RankTableStatus string

// Server to hccl
type Server struct {
	DeviceList   []*Device `json:"device"`    // device list in each server
	ServerID     string    `json:"server_id"` // server id, represented by ip address
	ContainerIP  string    `json:"container_ip,omitempty"`
	HardwareType string    `json:"hardware_type,omitempty"`
}

// Device in hccl.json
type Device struct {
	DeviceID        string `json:"device_id"` // hccl deviceId
	DeviceIP        string `json:"device_ip"` // hccl deviceIp
	SuperDeviceID   string `json:"super_device_id,omitempty"`
	DeviceLogicalId string `json:"device_logical_id,omitempty"`
	RankID          string `json:"rank_id"`
}

// A2RankTable is rank table for a2
type A2RankTable struct {
	deployServer     string
	Status           RankTableStatus `json:"status"`
	ServerList       []*Server       `json:"server_list" json:"server_list,omitempty"`       // hccl_json server list
	ServerCount      string          `json:"server_count" json:"server_count,omitempty"`     // hccl_json server count
	Version          string          `json:"version" json:"version,omitempty"`               // hccl_json version
	SuperPodInfoList []*SuperPodInfo `json:"super_pod_list" json:"super_pod_list,omitempty"` // hccl_json super pod list
}

// SuperPodInfo in superpod hccl.json
type SuperPodInfo struct {
	SuperPodId         string            `json:"super_pod_id" json:"super_pod_id,omitempty"` // hccl_json super pod id
	SuperPodServerList []*SuperPodServer `json:"server_list" json:"server_list,omitempty"`   // hccl_json super pod server list
}

// SuperPodServer in superpod hccl.json
type SuperPodServer struct {
	SuperPodServerId string `json:"server_id" json:"server_id,omitempty"` // hccl_json super pod server id
}

// PdDeployModeRankTable is global rank table for single node or cross node pd deploy mode
type PdDeployModeRankTable struct {
	Version string          `json:"version" json:"version,omitempty"` // hccl_json version
	Status  RankTableStatus `json:"status"`
	// ServerGroupList hccl_json server group list
	ServerGroupList []*ServerGroup `json:"server_group_list" json:"server_group_list,omitempty"`
}

// PdDeployModeServer is server for pd deploy mode
type PdDeployModeServer struct {
	DeviceList   []*Device `json:"device,omitempty"` // device list in each server
	DeployServer string    `json:"deploy_server,omitempty"`
	ServerID     string    `json:"server_id"`               // server id, represented by ip address
	ContainerIP  string    `json:"server_ip,omitempty"`     // pod ip
	HardwareType string    `json:"hardware_type,omitempty"` // hardware type
}

// ServerGroup is server group
type ServerGroup struct {
	GroupId      string                `json:"group_id"`
	DeployServer string                `json:"deploy_server,omitempty"`
	ServerCount  string                `json:"server_count"`
	ServerList   []*PdDeployModeServer `json:"server_list"`
	SuperPodList []*SuperPodInfo       `json:"super_pod_list,omitempty"`
}

// parseMindIeRankTableCM parse mindie rank table configmap
func parseMindIeRankTableCM(obj interface{}) (*A2RankTable, error) {
	ranktableInfoCm, ok := obj.(*v1.ConfigMap)
	if !ok {
		return nil, fmt.Errorf("not configmap")
	}
	ranktableInfo := ranktableInfoCm.Data[job.HcclJson]

	var a2RankTable A2RankTable
	err := json.Unmarshal([]byte(ranktableInfo), &a2RankTable)
	if err != nil {
		return nil, err
	}

	grtServer, hasGrtServer := ranktableInfoCm.Labels[standaloneDeployServerKey]
	grtGroup, hasGrtGroup := ranktableInfoCm.Labels[distributedDeployServerKey]

	switch {
	case hasGrtServer && hasGrtGroup:
		return nil, fmt.Errorf("%s and %s cannot exist at the same time",
			standaloneDeployServerKey, distributedDeployServerKey)
	case hasGrtServer:
		a2RankTable.deployServer = grtServer
	case hasGrtGroup:
		a2RankTable.deployServer = grtGroup
	default:
		return nil, fmt.Errorf("configmap(%s) no %s or %s label", ranktableInfoCm.Name,
			standaloneDeployServerKey, distributedDeployServerKey)
	}

	return &a2RankTable, nil
}

func getPdDeployModeServers(nameSpace, jobId, appType string) ([]*PdDeployModeServer, error) {
	serverJobKey, err := job.GetInstanceJobKey(jobId, nameSpace, appType)
	if err != nil {
		return nil, err
	}
	podMap := pod.GetPodByJobId(serverJobKey)

	if len(podMap) == 0 {
		return nil, fmt.Errorf("%s server pod num is 0", appType)
	}

	var servers []*PdDeployModeServer
	for _, item := range podMap {
		if item.Status.HostIP == "" || item.Status.PodIP == "" {
			return nil, fmt.Errorf("%s server pod is not scheduled", appType)
		}
		servers = append(servers, &PdDeployModeServer{
			ServerID:    item.Status.HostIP,
			ContainerIP: item.Status.PodIP,
		})
	}
	return servers, nil
}

// GetA2RankTableList get completed a2 rank table list
func GetA2RankTableList(message *GenerateGlobalRankTableMessage) ([]*A2RankTable, error) {
	jobId := message.JobId
	hwlog.RunLog.Debugf("will get a2 rank table list, jobId is : %s", jobId)
	nameSpace := message.Namespace
	ranktableCmList, err := GetAllEpRankTableCm(jobId, nameSpace)
	if err != nil {
		return nil, err
	}

	// sub ranktable List
	var a2RankTableList []*A2RankTable
	for _, ranktableCmItem := range *ranktableCmList {
		a2RankTable, err := parseMindIeRankTableCM(&ranktableCmItem)
		if err != nil {
			return nil, err
		}
		if a2RankTable.Status == constant.StatusRankTableCompleted {
			a2RankTableList = append(a2RankTableList, a2RankTable)
		}
	}
	if len(a2RankTableList) == 0 {
		return nil, fmt.Errorf("no completed a2 rank table")
	}
	return a2RankTableList, nil
}

// GenerateServerGroup0Or1 generate server group 0 or 1
func GenerateServerGroup0Or1(message *GenerateGlobalRankTableMessage, appType string) (*ServerGroup, error) {
	jobId := message.JobId
	nameSpace := message.Namespace
	servers, err := getPdDeployModeServers(nameSpace, jobId, appType)
	if err != nil {
		hwlog.RunLog.Errorf("get %s server id and ip failed, err: %v", appType, err)
		return nil, err
	}
	groupId := constant.GroupId0
	if appType == constant.ControllerAppType {
		groupId = constant.GroupId1
	}
	serverGroup := &ServerGroup{
		GroupId:     groupId,
		ServerCount: strconv.Itoa(len(servers)),
		ServerList:  servers,
	}
	return serverGroup, nil
}

// GenerateServerGroup2 generate server group 2
func GenerateServerGroup2(a2RankTableList []*A2RankTable) *ServerGroup {
	pdDeployModeServerList := make([]*PdDeployModeServer, len(a2RankTableList))
	for i, a2RankTable := range a2RankTableList {
		server := a2RankTable.ServerList[0]
		pdDeployModeServer := &PdDeployModeServer{
			DeployServer: a2RankTable.deployServer,
			ServerID:     server.ServerID,
			ContainerIP:  server.ContainerIP,
			DeviceList:   server.DeviceList,
			HardwareType: server.HardwareType,
		}
		// set DeviceLogicalId
		for logicalId, device := range pdDeployModeServer.DeviceList {
			device.DeviceLogicalId = fmt.Sprintf("%d", logicalId)
		}
		pdDeployModeServerList[i] = pdDeployModeServer
	}
	serverGroup2 := &ServerGroup{
		GroupId:     constant.GroupId2,
		ServerCount: fmt.Sprintf("%d", len(a2RankTableList)),
		ServerList:  pdDeployModeServerList,
	}
	return serverGroup2
}

// GenerateServerGroupList generate server group list
func GenerateServerGroupList(a2RankTableList []*A2RankTable) []*ServerGroup {
	serverGroupList := make([]*ServerGroup, len(a2RankTableList))
	for i, a2RankTable := range a2RankTableList {
		pdDeployModeServerList := make([]*PdDeployModeServer, len(a2RankTable.ServerList))
		for j, server := range a2RankTable.ServerList {
			pdDeployModeServer := &PdDeployModeServer{
				ServerID:     server.ServerID,
				ContainerIP:  server.ContainerIP,
				DeviceList:   server.DeviceList,
				HardwareType: server.HardwareType,
			}
			// set DeviceLogicalId
			for logicalId, device := range pdDeployModeServer.DeviceList {
				device.DeviceLogicalId = fmt.Sprintf("%d", logicalId)
			}
			pdDeployModeServerList[j] = pdDeployModeServer
		}
		serverGroup := &ServerGroup{
			GroupId:      fmt.Sprintf("%d", i+constant.GroupIdOffset),
			DeployServer: a2RankTable.deployServer,
			ServerCount:  a2RankTable.ServerCount,
			ServerList:   pdDeployModeServerList,
			SuperPodList: a2RankTable.SuperPodInfoList,
		}
		serverGroupList[i] = serverGroup
	}
	hwlog.RunLog.Debugf("GenerateServerGroupList : %v", serverGroupList)
	return serverGroupList
}

// getGlobalRankTableInfo get string global rank table info
func getGlobalRankTableInfo(a2RankTableList []*A2RankTable, serverGroup0, serverGroup1 *ServerGroup,
	pdDeploymentMode string) (string, error) {
	var pdDeployModeRankTable = &PdDeployModeRankTable{}
	switch pdDeploymentMode {
	case constant.SingleNodePdDeployMode:
		serverGroup2 := GenerateServerGroup2(a2RankTableList)
		singleNodePdDeployModeRankTable := &PdDeployModeRankTable{
			Version: constant.RankTableVersion,
			Status:  constant.StatusRankTableCompleted,
			ServerGroupList: []*ServerGroup{
				serverGroup0,
				serverGroup1,
				serverGroup2,
			},
		}
		pdDeployModeRankTable = singleNodePdDeployModeRankTable
	case constant.CrossNodePdDeployMode:
		serverGroupList := GenerateServerGroupList(a2RankTableList)
		crossNodePdDeployModeRankTable := &PdDeployModeRankTable{
			Version: a2RankTableList[0].Version,
			Status:  constant.StatusRankTableCompleted,
			ServerGroupList: []*ServerGroup{
				serverGroup0,
				serverGroup1,
			},
		}
		crossNodePdDeployModeRankTable.ServerGroupList =
			append(crossNodePdDeployModeRankTable.ServerGroupList, serverGroupList...)
		pdDeployModeRankTable = crossNodePdDeployModeRankTable
	default:
		return "", fmt.Errorf("pd deployment mode is invalid")
	}
	globalRankTableInfo, err := pdDeployModeRankTable.ToString()
	if err != nil {
		return "", err
	}
	return globalRankTableInfo, nil
}

// ToString convert rank table to string
func (pmr *PdDeployModeRankTable) ToString() (string, error) {
	bytes, err := json.Marshal(pmr)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// GetAllEpRankTableCm get all ep ranktable configmap
func GetAllEpRankTableCm(jobId, nameSpace string) (*[]v1.ConfigMap, error) {
	// retrieve the configmap of all names under nameSpace with the prefix "rings config -" and
	// the label containing jobID=jobId
	var resCm []v1.ConfigMap
	// get the list of all configmaps in the informer cache
	itemList := kube.GetCmInformer().GetIndexer().List()
	for _, item := range itemList {
		cm, ok := item.(*v1.ConfigMap)
		if !ok {
			return nil, fmt.Errorf("failed to convert informer cache indexer to configmap")
		}
		if cm.Namespace == nameSpace && jobId == cm.Labels[constant.MindIeJobIdLabelKey] &&
			strings.HasPrefix(cm.Name, constant.MindIeRanktablePrefix) {
			resCm = append(resCm, *cm)
		}
	}
	if len(resCm) == 0 {
		return nil, fmt.Errorf("no rank table")
	}
	return &resCm, nil
}
