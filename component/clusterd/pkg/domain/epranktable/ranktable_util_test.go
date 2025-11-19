// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

//go:build !race

// Package epranktable tests for ranktable_util.go
package epranktable

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/domain/pod"
)

const (
	jobId       = "job123"
	namespace   = "default"
	appType     = "appType"
	serverId    = "192.168.1.1"
	serverIp    = "10.0.0.1"
	groupId0    = "0"
	groupId1    = "1"
	serverCount = "1"
	deviceId    = "0"
	rankId      = "0"
	len1        = 1
	rankTableCm = `{"status":"completed","server_list":[{"server_id":"serverId",
    "device":[{"device_id":"deviceId","device_ip":"deviceIp"}]}]}`
	superDeviceId    = "8888"
	superPodId       = "1"
	superPodServerId = "192.168.1.2"
)

var (
	message = &GenerateGlobalRankTableMessage{
		JobId:     jobId,
		Namespace: namespace,
	}
	a2RankTableList = []*A2RankTable{
		{
			ServerList: []*Server{
				{
					ServerID:    serverId,
					ContainerIP: serverIp,
					DeviceList: []*Device{
						{
							DeviceID:      deviceId,
							RankID:        rankId,
							SuperDeviceID: superDeviceId,
						},
					},
				},
			},
			SuperPodInfoList: []*SuperPodInfo{
				{
					SuperPodId: superPodId,
					SuperPodServerList: []*SuperPodServer{
						{
							SuperPodServerId: superPodServerId,
						},
					},
				},
			},
		},
	}

	a2RankTableList2 = []*A2RankTable{
		{
			ServerList: []*Server{
				{
					ServerID:    serverId,
					ContainerIP: serverIp,
					DeviceList: []*Device{
						{
							DeviceID:      deviceId,
							RankID:        rankId,
							SuperDeviceID: superDeviceId,
						},
					},
				},
			},
		},
	}
)

var a3RankTableString = `
{
    "status": "completed",         
    "version": "1.2",              
    "server_count":"4",            
    "server_list": [
        {
            "server_id": "node_0",     
            "host_ip":"172.16.0.100",  
            "device": [
                {"device_id": "0","super_device_id":"0","device_ip": "192.168.1.6","device_port":"16666","backup_device_ip":"192.168.1.7","backup_device_port":"16667","host_port":"16665","rank_id": "0"}, 
                {"device_id": "1","super_device_id":"1","device_ip": "192.168.1.7","device_port":"16666","backup_device_ip":"192.168.1.6","backup_device_port":"16667","host_port":"16666","rank_id": "1"},
                {"device_id": "2","super_device_id":"2","device_ip": "192.168.1.8","device_port":"16668","backup_device_ip":"192.168.1.9","backup_device_port":"16670","host_port":"16667","rank_id": "2"},
                {"device_id": "3","super_device_id":"3","device_ip": "192.168.1.9","device_port":"16669","backup_device_ip":"192.168.1.8","backup_device_port":"16667","host_port":"16668","rank_id": "3"}]
        },
        {
            "server_id": "node_1",
            "host_ip":"172.16.0.101",
            "device": [
                {"device_id": "0","super_device_id":"4","device_ip": "192.168.2.6","device_port":"16666","backup_device_ip":"192.168.2.7","backup_device_port":"16667","host_port":"16665","rank_id": "4"},
                {"device_id": "1","super_device_id":"5","device_ip": "192.168.2.7","device_port":"16666","backup_device_ip":"192.168.2.6","backup_device_port":"16667","host_port":"16666","rank_id": "5"},
                {"device_id": "2","super_device_id":"6","device_ip": "192.168.2.8","device_port":"16668","backup_device_ip":"192.168.2.9","backup_device_port":"16670","host_port":"16667","rank_id": "6"},
                {"device_id": "3","super_device_id":"7","device_ip": "192.168.2.9","device_port":"16669","backup_device_ip":"192.168.2.8","backup_device_port":"16667","host_port":"16668","rank_id": "7"}]
        },
        {
            "server_id": "node_2",
            "host_ip":"172.16.0.102",
            "device": [
                {"device_id":"0","super_device_id":"0","device_ip":"192.168.3.6","device_port":"16666","backup_device_ip":"192.168.3.7","backup_device_port":"16667","host_port":"16665","rank_id":"8"},
                {"device_id":"1","super_device_id":"1","device_ip":"192.168.3.7","device_port":"16666","backup_device_ip":"192.168.3.6","backup_device_port":"16667","host_port":"16666","rank_id":"9"},
                {"device_id":"2","super_device_id":"2","device_ip":"192.168.3.8","device_port":"16668","backup_device_ip":"192.168.3.9","backup_device_port":"16670","host_port":"16667","rank_id":"10"},
                {"device_id":"3","super_device_id":"3","device_ip":"192.168.3.9","device_port":"16669","backup_device_ip":"192.168.3.8","backup_device_port":"16667","host_port":"16668","rank_id":"11"}]
        },
        {
            "server_id": "node_3",
            "host_ip":"172.16.0.103",
            "device": [
                {"device_id":"0","super_device_id":"4","device_ip":"192.168.4.6","device_port":"16666","backup_device_ip":"192.168.4.7","backup_device_port":"16667","host_port":"16665","rank_id":"12"},
                {"device_id":"1","super_device_id":"5","device_ip":"192.168.4.7","device_port":"16666","backup_device_ip":"192.168.4.6","backup_device_port":"16667","host_port":"16666","rank_id":"13"},
                {"device_id":"2","super_device_id":"6","device_ip":"192.168.4.8","device_port":"16668","backup_device_ip":"192.168.4.9","backup_device_port":"16670","host_port":"16667","rank_id":"14"},
                {"device_id":"3","super_device_id":"7","device_ip":"192.168.4.9","device_port":"16669","backup_device_ip":"192.168.4.8","backup_device_port":"16667","host_port":"16668","rank_id":"15"}]
        }
    ],
    "super_pod_list": [
        {
            "super_pod_id": "0",          
            "server_list": [              
                {"server_id": "node_0"},  
                {"server_id": "node_1"}]
        },
        {
            "super_pod_id": "1",
            "server_list": [
                {"server_id":"node_2"},
                {"server_id":"node_3"}]
        }
    ]
}
`

func TestParseMindIeRankTableCM(t *testing.T) {
	convey.Convey("test parseMindIeRankTableCM", t, func() {
		convey.Convey("01-when obj is not configmap should return error", func() {
			result, err := parseMindIeRankTableCM("invalid")
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(result, convey.ShouldBeNil)
		})
		cm := &v1.ConfigMap{
			Data: map[string]string{
				job.HcclJson: rankTableCm,
			},
		}
		convey.Convey("02-when configmap has labels of grt-server/deploy-server and grt-group/deploy-server should"+
			" return error", func() {
			cm.Labels = map[string]string{
				standaloneDeployServerKey:  "1",
				distributedDeployServerKey: "2",
			}
			result, err := parseMindIeRankTableCM(cm)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(result, convey.ShouldBeNil)
		})
		convey.Convey("02-when configmap has label of grt-server/deploy-server should return correct result",
			func() {
				cm.Labels = map[string]string{standaloneDeployServerKey: "1"}
				rankTable, err := parseMindIeRankTableCM(cm)
				convey.So(err, convey.ShouldBeNil)
				convey.So(rankTable.Status, convey.ShouldEqual, constant.StatusRankTableCompleted)
				convey.So(rankTable.deployServer, convey.ShouldEqual, "1")
			})
		convey.Convey("02-when configmap has label of grt-group/deploy-server should return correct result",
			func() {
				cm.Labels = map[string]string{distributedDeployServerKey: "2"}
				rankTable, err := parseMindIeRankTableCM(cm)
				convey.So(err, convey.ShouldBeNil)
				convey.So(rankTable.Status, convey.ShouldEqual, constant.StatusRankTableCompleted)
				convey.So(rankTable.deployServer, convey.ShouldEqual, "2")
			})
		convey.Convey("02-when configmap does not have superdevice into, should return correct result",
			func() {
				cm.Labels = map[string]string{distributedDeployServerKey: "2"}
				rankTable, err := parseMindIeRankTableCM(cm)
				convey.So(err, convey.ShouldBeNil)
				convey.So(rankTable.SuperPodInfoList, convey.ShouldEqual, nil)
				convey.So(rankTable.ServerList[0].DeviceList[0].SuperDeviceID, convey.ShouldEqual, "")
			})
	})
}

func TestGetServerIdAndIp(t *testing.T) {
	convey.Convey("test GetServerIdAndIp", t, func() {
		convey.Convey("when GetInstanceJobKey fails", func() {
			_, err := getPdDeployModeServers(namespace, jobId, appType)
			convey.So(err, convey.ShouldNotBeNil)
		})
		mockGetInstanceJobKey := gomonkey.ApplyFunc(job.GetInstanceJobKey, func(_, _, _ string) (string, error) {
			return "serverJobKey", nil
		})
		convey.Convey("when GetPodByJobId fails", func() {
			defer mockGetInstanceJobKey.Reset()
			_, err := getPdDeployModeServers(namespace, jobId, appType)
			convey.So(err, convey.ShouldResemble, fmt.Errorf(appType+" server pod num is 0"))
		})
		convey.Convey("when pod is not scheduled", func() {
			defer mockGetInstanceJobKey.Reset()
			fakePod := v1.Pod{
				Spec:   v1.PodSpec{},
				Status: v1.PodStatus{},
			}
			podMap := map[string]v1.Pod{"pod1": fakePod}
			mockGetPodByJobId := gomonkey.ApplyFunc(pod.GetPodByJobId, func(string) map[string]v1.Pod {
				return podMap
			})
			defer mockGetPodByJobId.Reset()
			_, err := getPdDeployModeServers(namespace, jobId, appType)
			convey.So(err, convey.ShouldResemble, fmt.Errorf(appType+" server pod is not scheduled"))
		})
		convey.Convey("when pod is scheduled", func() {
			defer mockGetInstanceJobKey.Reset()
			mockGetPodByJobId := gomonkey.ApplyFunc(pod.GetPodByJobId, func(_ string) map[string]v1.Pod {
				return map[string]v1.Pod{
					"pod1": {
						Status: v1.PodStatus{
							HostIP: serverId,
							PodIP:  serverIp,
						},
						Spec: v1.PodSpec{
							NodeName: "node1",
						},
					},
				}
			})
			defer mockGetPodByJobId.Reset()
			servers, err := getPdDeployModeServers(namespace, jobId, appType)
			convey.So(err, convey.ShouldBeNil)
			convey.So(servers[0].ServerID, convey.ShouldEqual, serverId)
			convey.So(servers[0].ContainerIP, convey.ShouldEqual, serverIp)
		})
	})
}

func TestGetA2RankTableList(t *testing.T) {
	convey.Convey("test GetA2RankTableList", t, func() {
		convey.Convey("when GetAllEpRankTableCm fails", func() {
			mockGetAllEpRankTableCm := gomonkey.ApplyFunc(GetAllEpRankTableCm,
				func(_, _ string) (*[]v1.ConfigMap, error) {
					return nil, fmt.Errorf("error")
				})
			defer mockGetAllEpRankTableCm.Reset()
			_, err := GetA2RankTableList(message)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("when GetA2RankTableList succeeds", func() {
			cmList := []v1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							standaloneDeployServerKey: "1",
						},
					},
					Data: map[string]string{
						job.HcclJson: rankTableCm,
					},
				},
			}
			mockGetAllEpRankTableCm := gomonkey.ApplyFunc(GetAllEpRankTableCm,
				func(_, _ string) (*[]v1.ConfigMap, error) {
					return &cmList, nil
				})
			defer mockGetAllEpRankTableCm.Reset()
			rankTableList, err := GetA2RankTableList(message)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(rankTableList), convey.ShouldEqual, len1)
		})
	})
}

func TestGenerateServerGroup0Or1(t *testing.T) {
	convey.Convey("test GenerateServerGroup0Or1", t, func() {
		convey.Convey("when GetServerIdAndIp fails", func() {
			mockGetServerIdAndIp := gomonkey.ApplyFunc(getPdDeployModeServers,
				func(_, _, _ string) ([]*PdDeployModeServer, error) {
					return nil, fmt.Errorf("error")
				})
			defer mockGetServerIdAndIp.Reset()
			_, err := GenerateServerGroup0Or1(message, appType)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("when group is 0 and 1 ", func() {
			mockGetServerIdAndIp := gomonkey.ApplyFunc(getPdDeployModeServers,
				func(_, _, _ string) ([]*PdDeployModeServer, error) {
					return []*PdDeployModeServer{
						{
							ServerID:    serverId,
							ContainerIP: serverIp,
						},
					}, nil
				})
			defer mockGetServerIdAndIp.Reset()
			group, err := GenerateServerGroup0Or1(message, constant.CoordinatorAppType)
			convey.So(err, convey.ShouldBeNil)
			convey.So(group.GroupId, convey.ShouldEqual, groupId0)
			group, err = GenerateServerGroup0Or1(message, constant.ControllerAppType)
			convey.So(err, convey.ShouldBeNil)
			convey.So(group.GroupId, convey.ShouldEqual, groupId1)
		})
	})
}

func TestGenerateServerGroupList(t *testing.T) {
	convey.Convey("test GenerateServerGroupList", t, func() {
		groupList := GenerateServerGroupList(a2RankTableList)
		convey.So(len(groupList), convey.ShouldEqual, len1)
		convey.So(groupList[0].ServerList[0].ServerID, convey.ShouldEqual, serverId)
		convey.So(groupList[0].ServerList[0].DeviceList[0].SuperDeviceID, convey.ShouldEqual, superDeviceId)
		convey.So(groupList[0].SuperPodList[0].SuperPodId, convey.ShouldEqual, superPodId)
		convey.So(groupList[0].SuperPodList[0].SuperPodServerList[0].SuperPodServerId, convey.ShouldEqual, superPodServerId)
	})
	convey.Convey("test GenerateServerGroupList do not have SuperPodInfoList ", t, func() {
		groupList := GenerateServerGroupList(a2RankTableList2)
		convey.So(groupList[0].ServerList[0].ServerID, convey.ShouldEqual, serverId)
		convey.So(groupList[0].ServerList[0].DeviceList[0].SuperDeviceID, convey.ShouldEqual, superDeviceId)
		convey.So(groupList[0].SuperPodList, convey.ShouldEqual, nil)
	})
}

func TestGetGlobalRankTableInfo(t *testing.T) {
	convey.Convey("test getGlobalRankTableInfo", t, func() {
		serverGroup0 := &ServerGroup{
			GroupId:     groupId0,
			ServerCount: serverCount,
			ServerList: []*PdDeployModeServer{
				{
					ServerID:    serverId,
					ContainerIP: serverIp,
				},
			},
		}
		serverGroup1 := &ServerGroup{
			GroupId:     groupId1,
			ServerCount: serverCount,
			ServerList: []*PdDeployModeServer{
				{
					ServerID:    serverId,
					ContainerIP: serverIp,
				},
			},
		}
		convey.Convey("when pdDeploymentMode is CrossNodePdDeployMode", func() {
			info, err := getGlobalRankTableInfo(a2RankTableList, serverGroup0, serverGroup1)
			convey.So(err, convey.ShouldBeNil)
			convey.So(info, convey.ShouldNotBeEmpty)
		})
	})
}

func TestConvertGrtServerId(t *testing.T) {
	var a2RankTable A2RankTable
	err := json.Unmarshal([]byte(a3RankTableString), &a2RankTable)
	if err != nil {
		t.Errorf("convert a3RankTableString to ranktable error %s", err)
	}

	convertGrtServerId(&a2RankTable)
	serverIds := make(map[string]struct{})
	for _, server := range a2RankTable.ServerList {
		serverIds[server.ServerID] = struct{}{}
		if server.HostIp != server.ServerID {
			t.Errorf("HostIp %s and ServerID %s is not equal", server.HostIp, server.ServerID)
		}
	}
	for _, info := range a2RankTable.SuperPodInfoList {
		for _, server := range info.SuperPodServerList {
			if _, ok := serverIds[server.SuperPodServerId]; !ok {
				t.Errorf("SuperPodServerId %s is not host ip", server.SuperPodServerId)
			}
		}
	}
}
