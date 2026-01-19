// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

//go:build !race

// Package epranktable tests for ranktable_manager_go
package epranktable

import (
	"context"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/job"
)

const (
	testNamespace = "test-namespace1"
	testJobId     = "test-job-id1"
	testRankTable = "test-rank-table1"
)

var (
	deviceList = []*Device{
		{DeviceID: "0", RankID: "0"},
	}
	msg = &GenerateGlobalRankTableMessage{
		JobId:     testJobId,
		Namespace: testNamespace,
	}
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

func TestGeneratePdDeployModeRankTable(t *testing.T) {
	convey.Convey("test GeneratePdDeployModeRankTable", t, func() {
		convey.Convey("when get rank table list fails", func() {
			patches := gomonkey.ApplyFunc(GetA2RankTableList, func(*GenerateGlobalRankTableMessage) ([]*A2RankTable, error) {
				return nil, fmt.Errorf("mock error")
			})
			defer patches.Reset()
			info, retry := GeneratePdDeployModeRankTable(msg)
			convey.So(retry, convey.ShouldBeTrue)
			convey.So(info, convey.ShouldBeEmpty)
		})
		convey.Convey("when successful", func() {
			patchGetRankTable := gomonkey.ApplyFunc(GetA2RankTableList, func(*GenerateGlobalRankTableMessage) ([]*A2RankTable, error) {
				return []*A2RankTable{
					{
						Status: constant.StatusRankTableCompleted,
						ServerList: []*Server{
							{
								ServerID:   serverId,
								DeviceList: deviceList,
							},
						},
						ServerCount: serverCount,
					},
				}, nil
			})
			defer patchGetRankTable.Reset()
			patchGenGroup := gomonkey.ApplyFunc(GenerateServerGroup0Or1, func(*GenerateGlobalRankTableMessage, string) (*ServerGroup, error) {
				return &ServerGroup{
					GroupId:     constant.GroupId0,
					ServerCount: serverCount,
					ServerList: []*PdDeployModeServer{
						{
							ServerID:   serverId,
							DeviceList: deviceList,
						},
					},
				}, nil
			})
			defer patchGenGroup.Reset()
			info, retry := GeneratePdDeployModeRankTable(msg)
			convey.So(retry, convey.ShouldBeFalse)
			convey.So(info, convey.ShouldNotBeEmpty)
		})
	})
}

func TestPushGlobalRankTable(t *testing.T) {
	convey.Convey("test pushGlobalRankTable", t, func() {
		rm := GetEpGlobalRankTableManager()
		message := &GenerateGlobalRankTableMessage{JobId: testJobId, Namespace: testNamespace}
		convey.Convey("when HandlerRankTable is nil", func() {
			rm.HandlerRankTable = nil
			rm.pushGlobalRankTable(message, testRankTable)
			convey.So(rm.rankTableQueue.NumRequeues(message), convey.ShouldEqual, 0)
		})
		convey.Convey("when HandlerRankTable fails", func() {
			rm.HandlerRankTable = func(_, _ string) (bool, error) {
				return false, fmt.Errorf("error")
			}
			rm.pushGlobalRankTable(message, testRankTable)
			convey.So(rm.rankTableQueue.NumRequeues(message), convey.ShouldEqual, 1)
		})
		convey.Convey("when HandlerRankTable succeeds", func() {
			rm.HandlerRankTable = func(_, _ string) (bool, error) {
				return true, nil
			}
			rm.pushGlobalRankTable(message, testRankTable)
			convey.So(rm.rankTableQueue.NumRequeues(message), convey.ShouldEqual, 0)
		})
	})
}

func TestEpRankTableInformerHandler(t *testing.T) {
	convey.Convey("test InformerHandler", t, func() {
		convey.Convey("with ConfigMap", func() {
			cm := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cm",
					Namespace: testNamespace,
					Labels: map[string]string{
						constant.MindIeJobIdLabelKey: testJobId,
					},
				},
			}
			InformerHandler(nil, cm, "")
			msg, _ := epGlobalRankTableManager.rankTableQueue.Get()
			if msg != nil {
				rankTableMsg, ok := msg.(*GenerateGlobalRankTableMessage)
				convey.So(ok, convey.ShouldBeTrue)
				convey.So(rankTableMsg.JobId, convey.ShouldEqual, testJobId)
				convey.So(rankTableMsg.Namespace, convey.ShouldEqual, testNamespace)
			}
		})
		convey.Convey("with Pod", func() {
			pod := &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: testNamespace,
					Labels: map[string]string{
						constant.MindIeJobIdLabelKey: testJobId,
					},
				},
			}
			InformerHandler(nil, pod, "")
			msg, _ := epGlobalRankTableManager.rankTableQueue.Get()
			if msg != nil {
				rankTableMsg, ok := msg.(*GenerateGlobalRankTableMessage)
				convey.So(ok, convey.ShouldBeTrue)
				convey.So(rankTableMsg.JobId, convey.ShouldEqual, testJobId)
				convey.So(rankTableMsg.Namespace, convey.ShouldEqual, testNamespace)
			}
		})
		convey.Convey("with invalid object", func() {
			InformerHandler(nil, "invalid", "")
		})
	})
}

func TestGetMessageInfo(t *testing.T) {
	convey.Convey("test getMessageInfo", t, func() {
		rm := GetEpGlobalRankTableManager()
		convey.Convey("when item is not GenerateGlobalRankTableMessage", func() {
			item := "invalid"
			message, err := rm.getMessageInfo(item)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(message, convey.ShouldBeNil)
		})
		convey.Convey("when message has non-empty namespace", func() {
			item := &GenerateGlobalRankTableMessage{JobId: testJobId, Namespace: testNamespace}
			message, err := rm.getMessageInfo(item)
			convey.So(err, convey.ShouldBeNil)
			convey.So(message.Namespace, convey.ShouldEqual, testNamespace)
		})
		convey.Convey("when GetNamespaceByJobIdAndAppType fails", func() {
			item := &GenerateGlobalRankTableMessage{JobId: testJobId, Namespace: ""}
			patches := gomonkey.ApplyFunc(job.GetNamespaceByJobIdAndAppType, func(string, string) (string, error) {
				return "", fmt.Errorf("mock error")
			})
			defer patches.Reset()
			message, err := rm.getMessageInfo(item)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(message.Namespace, convey.ShouldEqual, "")
		})
	})
}
