/*
Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package v1 is using for v1 Ranktable.
*/
package v1

import (
	"ascend-operator/pkg/api/v1"
	"ascend-operator/pkg/ranktable/common"
)

const (
	rankTableVersion = "1.0"
)

// RankTable ranktable of v1
type RankTable struct {
	*common.BaseGenerator
}

// New create ranktable generator
func New(job *v1.AscendJob) *RankTable {
	r := &RankTable{}
	r.BaseGenerator = common.NewBaseGenerator(job, rankTableVersion, r)
	return r
}
