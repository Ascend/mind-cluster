/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

package v1

// SuccessPolicy is the success policy.
type SuccessPolicy string

const (
	SuccessPolicyDefault    SuccessPolicy = ""
	SuccessPolicyAllWorkers SuccessPolicy = "AllWorkers"
)
