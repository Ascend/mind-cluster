// Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
// package epranktable for
package epranktable

import (
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/api/core/v1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/domain/pod"
	"clusterd/pkg/interface/kube"
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
	Status      RankTableStatus `json:"status"`
	ServerList  []*Server       `json:"server_list" json:"server_list,omitempty"`   // hccl_json server list
	ServerCount string          `json:"server_count" json:"server_count,omitempty"` // hccl_json server count
	Version     string          `json:"version" json:"version,omitempty"`           // hccl_json version
}

// PdDeployModeRankTable is global rank table for single node or cross node pd deploy mode
type PdDeployModeRankTable struct {
	Version         string          `json:"version" json:"version,omitempty"` // hccl_json version
	Status          RankTableStatus `json:"status"`
	ServerGroupList []*ServerGroup  `json:"server_group_list" json:"server_group_list,omitempty"` // hccl_json server group list
}

// PdDeployModeServer is server for pd deploy mode
type PdDeployModeServer struct {
	DeviceList   []*Device `json:"device,omitempty"`        // device list in each server
	ServerID     string    `json:"server_id"`               // server id, represented by ip address
	ContainerIP  string    `json:"server_ip,omitempty"`     // pod ip
	HardwareType string    `json:"hardware_type,omitempty"` // hardware type
}

// ServerGroup is server group
type ServerGroup struct {
	GroupId     string                `json:"group_id"`
	ServerCount string                `json:"server_count"`
	ServerList  []*PdDeployModeServer `json:"server_list"`
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
	return &a2RankTable, nil
}

// GetServerIdAndIp get server id and ip
func GetServerIdAndIp(nameSpace, jobId, appType string) (string, string, error) {
	serverJobKey, err := job.GetInstanceJobKey(jobId, nameSpace, appType)
	if err != nil {
		return "", "", err
	}
	podMap := pod.GetPodByJobId(serverJobKey)
	// later modifications are needed, as the controller may have multiple pods
	if len(podMap) == 0 || len(podMap) > 1 {
		return "", "", fmt.Errorf(appType + " server pod num is not 1")
	}

	var serverPod v1.Pod
	for _, item := range podMap {
		serverPod = item
		break
	}
	// 如果pod还未被调度
	if serverPod.Spec.NodeName == "" {
		return "", "", fmt.Errorf(appType + " server pod is not scheduled")
	}
	serverId := serverPod.Status.HostIP
	serverIp := serverPod.Status.PodIP
	return serverId, serverIp, nil
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
	serverId, serverIp, err := GetServerIdAndIp(nameSpace, jobId, appType)
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
		ServerCount: constant.ServerCountGroupId0Or1,
		ServerList: []*PdDeployModeServer{
			{
				ServerID:    serverId,
				ContainerIP: serverIp,
			},
		},
	}
	return serverGroup, nil
}

// GenerateServerGroup2 generate server group 2
func GenerateServerGroup2(a2RankTableList []*A2RankTable) *ServerGroup {
	pdDeployModeServerList := make([]*PdDeployModeServer, len(a2RankTableList))
	for i, a2RankTable := range a2RankTableList {
		server := a2RankTable.ServerList[0]
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
			GroupId:     fmt.Sprintf("%d", i+constant.GroupIdOffset),
			ServerCount: a2RankTable.ServerCount,
			ServerList:  pdDeployModeServerList,
		}
		serverGroupList[i] = serverGroup
	}
	return serverGroupList
}

// getGlobalRankTableInfo get string global rank table info
func getGlobalRankTableInfo(a2RankTableList []*A2RankTable, serverGroup0, serverGroup1 *ServerGroup,
	pdDeploymentMode string) (string, error) {
	var pdDeployModeRankTable *PdDeployModeRankTable
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
			Version: constant.RankTableVersion,
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
