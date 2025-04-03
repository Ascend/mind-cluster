// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

//go:build !race

// Package epranktable tests for ranktable_util.go
package epranktable

import (
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
	len1        = 1
	rankTableCm = `{"status":"completed","server_list":[{"server_id":"serverId",
    "device":[{"device_id":"deviceId","device_ip":"deviceIp"}]}]}`
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
							DeviceID: "0",
							RankID:   "0",
						},
					},
				},
			},
		},
	}
)

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
	})
}

func TestGetServerIdAndIp(t *testing.T) {
	convey.Convey("test GetServerIdAndIp", t, func() {
		convey.Convey("when GetInstanceJobKey fails", func() {
			_, _, err := GetServerIdAndIp(namespace, jobId, appType)
			convey.So(err, convey.ShouldNotBeNil)
		})
		mockGetInstanceJobKey := gomonkey.ApplyFunc(job.GetInstanceJobKey, func(_, _, _ string) (string, error) {
			return "serverJobKey", nil
		})
		convey.Convey("when GetPodByJobId fails", func() {
			defer mockGetInstanceJobKey.Reset()
			_, _, err := GetServerIdAndIp(namespace, jobId, appType)
			convey.So(err, convey.ShouldResemble, fmt.Errorf(appType+" server pod num is not 1"))
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
			_, _, err := GetServerIdAndIp(namespace, jobId, appType)
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
			id, ip, err := GetServerIdAndIp(namespace, jobId, appType)
			convey.So(err, convey.ShouldBeNil)
			convey.So(id, convey.ShouldEqual, serverId)
			convey.So(ip, convey.ShouldEqual, serverIp)
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
			mockGetServerIdAndIp := gomonkey.ApplyFunc(GetServerIdAndIp,
				func(_, _, _ string) (string, string, error) {
					return "", "", fmt.Errorf("error")
				})
			defer mockGetServerIdAndIp.Reset()
			_, err := GenerateServerGroup0Or1(message, appType)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("when group is 0 and 1 ", func() {
			mockGetServerIdAndIp := gomonkey.ApplyFunc(GetServerIdAndIp,
				func(_, _, _ string) (string, string, error) {
					return serverId, serverIp, nil
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

func TestGenerateServerGroup2(t *testing.T) {
	convey.Convey("test GenerateServerGroup2", t, func() {
		result := GenerateServerGroup2(a2RankTableList)
		convey.So(result.GroupId, convey.ShouldEqual, constant.GroupId2)
		convey.So(result.ServerCount, convey.ShouldEqual, serverCount)
		convey.So(len(result.ServerList), convey.ShouldEqual, len1)
		convey.So(result.ServerList[0].ServerID, convey.ShouldEqual, serverId)
	})
}

func TestGenerateServerGroupList(t *testing.T) {
	convey.Convey("test GenerateServerGroupList", t, func() {
		groupList := GenerateServerGroupList(a2RankTableList)
		convey.So(len(groupList), convey.ShouldEqual, len1)
		convey.So(groupList[0].ServerList[0].ServerID, convey.ShouldEqual, serverId)
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
		convey.Convey("when pdDeploymentMode is invalid", func() {
			_, err := getGlobalRankTableInfo(a2RankTableList, serverGroup0, serverGroup1, "invalid")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("when pdDeploymentMode is SingleNodePdDeployMode", func() {
			info, err := getGlobalRankTableInfo(a2RankTableList, serverGroup0, serverGroup1, constant.SingleNodePdDeployMode)
			convey.So(err, convey.ShouldBeNil)
			convey.So(info, convey.ShouldNotBeEmpty)
		})
		convey.Convey("when pdDeploymentMode is CrossNodePdDeployMode", func() {
			info, err := getGlobalRankTableInfo(a2RankTableList, serverGroup0, serverGroup1, constant.CrossNodePdDeployMode)
			convey.So(err, convey.ShouldBeNil)
			convey.So(info, convey.ShouldNotBeEmpty)
		})
	})
}
