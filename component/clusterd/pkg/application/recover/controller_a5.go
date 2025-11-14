// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package recover a series of service function
package recover

import (
	"slices"
	"strconv"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/interface/grpc/recover"
	"clusterd/pkg/interface/kube"
)

const (
	tpBlockStr16 = "16"
	tpBlockStr32 = "32"
	tpBlockStr64 = "64"
)

func (ctl *EventController) getAllRankByTPBlock() ([]*pb.FaultRank, []string, error) {
	var allFaults []*pb.FaultRank
	var allFaultRanks []string
	pg, err := kube.RetryGetPodGroup(ctl.jobInfo.PgName, ctl.jobInfo.Namespace, constant.GetPodGroupTimes)
	if err != nil {
		hwlog.RunLog.Errorf("kube get pod group err, err=%v", err)
		return allFaults, allFaultRanks, err
	}
	tpBlock, ok := pg.Annotations[constant.TpBlock]
	if ok && slices.Contains([]string{tpBlockStr16, tpBlockStr32, tpBlockStr64}, tpBlock) {
		if tpBlockInt, err2 := strconv.Atoi(tpBlock); err2 == nil {
			allFaults, allFaultRanks = ctl.normalFaultAssociateSameTpRank(tpBlockInt)
		}
	} else {
		allFaults, allFaultRanks = ctl.normalFaultAssociateSameNodeRank()
	}
	return allFaults, allFaultRanks, nil
}

func (ctl *EventController) normalFaultAssociateSameTpRank(devicePerTp int) ([]*pb.FaultRank, []string) {
	var faultRankIds []string
	for _, fault := range ctl.cacheNormalFault {
		faultRankIds = append(faultRankIds, fault.RankId)
	}
	allFaultRankIds := common.GetFaultRankIdsInSameTp(faultRankIds, devicePerTp)
	removeSameRankIds := util.RemoveSliceDuplicateElement(allFaultRankIds)
	var res []*pb.FaultRank
	for _, rank := range removeSameRankIds {
		res = append(res, &pb.FaultRank{
			RankId:    rank,
			FaultType: constant.NormalFaultType,
		})
	}
	return res, removeSameRankIds
}
