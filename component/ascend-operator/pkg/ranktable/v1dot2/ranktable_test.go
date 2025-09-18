/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package v1dot2 is using for v1.2 Ranktable.
*/
package v1dot2

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/smartystreets/goconvey/convey"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "ascend-operator/pkg/api/v1"
	"ascend-operator/pkg/ranktable/common"
	_ "ascend-operator/pkg/testtool"
	"ascend-operator/pkg/utils"
)

func TestGatherServerList(t *testing.T) {
	convey.Convey("TestGatherServerList", t, func() {
		convey.Convey("01-servers will be set to serverList and SuperPodList", func() {
			job := &v1.Job{}
			job.Annotations = map[string]string{utils.AnnoKeyOfSuperPod: "2"}
			gen := New(job)
			gen.ServerList = []*common.Server{
				{
					DeviceList: []*common.Device{
						{RankID: "0"},
						{RankID: "1"},
					},
					ServerID: "127.0.0.1",
				},
				{
					DeviceList: []*common.Device{
						{RankID: "2"},
						{RankID: "3"},
					},
					ServerID: "127.0.0.2",
				},
			}
			patch := gomonkey.ApplyMethod(new(common.BaseGenerator), "GatherServerList",
				func(*common.BaseGenerator) {})
			defer patch.Reset()
			gen.GatherServerList()
			expected := 2
			convey.So(len(gen.SuperPodList), convey.ShouldEqual, expected)
			convey.So(gen.SuperPodList[0].ServerList[0].ServerID, convey.ShouldEqual, "127.0.0.1")
		})
	})
}

func TestGatherServerListForSoftStrategy(t *testing.T) {
	patch := gomonkey.ApplyMethod(new(common.BaseGenerator), "GatherServerList",
		func(*common.BaseGenerator) {})
	defer patch.Reset()
	convey.Convey("TestGatherServerListForSoftStrategy", t, func() {
		job := &v1.Job{
			Spec: v1.JobSpec{ReplicaSpecs: map[commonv1.ReplicaType]*commonv1.ReplicaSpec{
				v1.ReplicaTypeWorker: {Template: corev1.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{utils.SuperPodAffinity: utils.SoftStrategy}}}}}}}
		gen := New(job)
		convey.Convey("01-empty server list should return empty super pod list", func() {
			gen.ServerList = []*common.Server{}
			gen.GatherServerListForSoftStrategy()
			convey.So(len(gen.SuperPodList), convey.ShouldEqual, 0)
		})
		convey.Convey("02-single server should create one super pod", func() {
			gen.ServerList = []*common.Server{
				{
					ServerID:     "server-1",
					SuperPodRank: "0",
					DeviceList:   []*common.Device{{RankID: "0"}},
				},
			}
			gen.GatherServerListForSoftStrategy()
			convey.So(len(gen.SuperPodList), convey.ShouldEqual, 1)
			convey.So(gen.SuperPodList[0].SuperPodID, convey.ShouldEqual, "0")
			convey.So(len(gen.SuperPodList[0].ServerList), convey.ShouldEqual, 1)
		})
		convey.Convey("03-multiple servers with same super pod rank", func() {
			gen.ServerList = []*common.Server{
				{
					ServerID:     "server-1",
					SuperPodRank: "0",
					DeviceList:   []*common.Device{{RankID: "0"}},
				},
				{
					ServerID:     "server-2",
					SuperPodRank: "0",
					DeviceList:   []*common.Device{{RankID: "1"}},
				},
			}
			gen.GatherServerListForSoftStrategy()
			convey.So(len(gen.SuperPodList), convey.ShouldEqual, 1)
			convey.So(len(gen.SuperPodList[0].ServerList), convey.ShouldEqual, 2)
		})
	})
}
