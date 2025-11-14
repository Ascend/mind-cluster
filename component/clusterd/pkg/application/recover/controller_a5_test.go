// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package recover a series of controller test function
package recover

import (
	"errors"
	"strconv"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/domain/pod"
	"clusterd/pkg/interface/grpc/recover"
	"clusterd/pkg/interface/kube"
)

const (
	errRank          = 40
	deviceNumPerNode = 8
	tp8              = 8
	tp16             = 16
	tp32             = 32
	tp64             = 64
)

func generateGetAllRankByTPBlockSlice(tpBlock int) ([]*pb.FaultRank, []string) {
	start := errRank / tpBlock * tpBlock
	expectedFaults := make([]*pb.FaultRank, 0)
	expectedRanks := make([]string, 0)
	for i := 0; i < tpBlock; i++ {
		expectedFaults = append(expectedFaults, &pb.FaultRank{
			RankId:    strconv.Itoa(start + i),
			FaultType: constant.NormalFaultType,
		})
		expectedRanks = append(expectedRanks, strconv.Itoa(start+i))
	}
	return expectedFaults, expectedRanks
}

func TestGetAllRankByTPBlock(t *testing.T) {
	convey.Convey("TestGetAllRankByTPBlock", t, func() {
		pg := &v1beta1.PodGroup{}
		pg.Annotations = make(map[string]string)
		patchGetPodGroup := gomonkey.ApplyFunc(kube.RetryGetPodGroup,
			func(name, namespace string, retryTimes int) (*v1beta1.PodGroup, error) {
				return pg, nil
			})
		defer patchGetPodGroup.Reset()

		patchGetPodDevice := gomonkey.ApplyFunc(pod.GetPodDeviceNumByJobId,
			func(jobKey string) int {
				return deviceNumPerNode
			})
		defer patchGetPodDevice.Reset()

		ctl := &EventController{
			jobInfo: common.JobBaseInfo{
				JobId:         "test-job-id",
				JobName:       "test-job",
				Namespace:     "test-namespace",
				RecoverConfig: common.RecoverConfig{PlatFormMode: true, ProcessRecoverEnable: true},
			},
			faultPod: make(map[string]string),
			uuid:     "test-uuid",
			cacheNormalFault: []*pb.FaultRank{
				{
					RankId:    strconv.Itoa(errRank),
					FaultType: constant.NormalFaultType,
				},
			},
		}

		testGetAllRankByTPBlockWithTpBlock(ctl, pg)
		testGetAllRankByTPBlockWithUnknownTpBlock(ctl, pg)
		testGetAllRankByTPBlockWithKubeError(ctl, pg)
		testGetAllRankByTPBlockWithNoTpBlockAnnotation(ctl)
	})
}

func testGetAllRankByTPBlockWithTpBlock(ctl *EventController, pg *v1beta1.PodGroup) {
	convey.Convey("Test getAllRankByTPBlock with TpBlock=8", func() {
		pg.Annotations[constant.TpBlock] = strconv.Itoa(tp8)
		expectedFaults, expectedRanks := generateGetAllRankByTPBlockSlice(tp8)
		faults, ranks, err := ctl.getAllRankByTPBlock()
		convey.So(err, convey.ShouldBeNil)
		convey.So(faults, convey.ShouldResemble, expectedFaults)
		convey.So(ranks, convey.ShouldResemble, expectedRanks)
	})
	convey.Convey("Test getAllRankByTPBlock with TpBlock=16", func() {
		pg.Annotations[constant.TpBlock] = strconv.Itoa(tp16)
		expectedFaults, expectedRanks := generateGetAllRankByTPBlockSlice(tp16)
		faults, ranks, err := ctl.getAllRankByTPBlock()
		convey.So(err, convey.ShouldBeNil)
		convey.So(faults, convey.ShouldResemble, expectedFaults)
		convey.So(ranks, convey.ShouldResemble, expectedRanks)
	})
	convey.Convey("Test getAllRankByTPBlock with TpBlock=32", func() {
		pg.Annotations[constant.TpBlock] = strconv.Itoa(tp32)
		expectedFaults, expectedRanks := generateGetAllRankByTPBlockSlice(tp32)
		faults, ranks, err := ctl.getAllRankByTPBlock()
		convey.So(err, convey.ShouldBeNil)
		convey.So(faults, convey.ShouldResemble, expectedFaults)
		convey.So(ranks, convey.ShouldResemble, expectedRanks)
	})
	convey.Convey("Test getAllRankByTPBlock with TpBlock=64", func() {
		pg.Annotations[constant.TpBlock] = strconv.Itoa(tp64)
		expectedFaults, expectedRanks := generateGetAllRankByTPBlockSlice(tp64)
		faults, ranks, err := ctl.getAllRankByTPBlock()
		convey.So(err, convey.ShouldBeNil)
		convey.So(faults, convey.ShouldResemble, expectedFaults)
		convey.So(ranks, convey.ShouldResemble, expectedRanks)
	})
}
func testGetAllRankByTPBlockWithUnknownTpBlock(ctl *EventController, pg *v1beta1.PodGroup) {
	convey.Convey("Test getAllRankByTPBlock with unknown TpBlock", func() {
		pg.Annotations[constant.TpBlock] = "unknown"
		expectedFaults, expectedRanks := generateGetAllRankByTPBlockSlice(tp8)
		faults, ranks, err := ctl.getAllRankByTPBlock()
		convey.So(err, convey.ShouldBeNil)
		convey.So(faults, convey.ShouldResemble, expectedFaults)
		convey.So(ranks, convey.ShouldResemble, expectedRanks)
	})
}
func testGetAllRankByTPBlockWithKubeError(ctl *EventController, pg *v1beta1.PodGroup) {
	convey.Convey("Test getAllRankByTPBlock with kube error", func() {
		patchKube := gomonkey.ApplyFunc(kube.RetryGetPodGroup,
			func(name, namespace string, retryTimes int) (*v1beta1.PodGroup, error) {
				return pg, errors.New("kube error")
			})
		defer patchKube.Reset()
		faults, ranks, err := ctl.getAllRankByTPBlock()
		convey.So(err, convey.ShouldResemble, errors.New("kube error"))
		convey.So(faults, convey.ShouldBeEmpty)
		convey.So(ranks, convey.ShouldBeEmpty)
	})
}
func testGetAllRankByTPBlockWithNoTpBlockAnnotation(ctl *EventController) {
	convey.Convey("Test getAllRankByTPBlock with no TpBlock annotation", func() {
		expectedFaults, expectedRanks := generateGetAllRankByTPBlockSlice(tp8)
		faults, ranks, err := ctl.getAllRankByTPBlock()
		convey.So(err, convey.ShouldBeNil)
		convey.So(faults, convey.ShouldResemble, expectedFaults)
		convey.So(ranks, convey.ShouldResemble, expectedRanks)
	})
}
