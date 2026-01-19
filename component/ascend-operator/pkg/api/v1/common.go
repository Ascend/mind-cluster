/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

package v1

// SuccessPolicy is the success policy.
type SuccessPolicy string

const (
	// SuccessPolicyDefault is the default policy of success
	SuccessPolicyDefault SuccessPolicy = ""
	// SuccessPolicyAllWorkers is the 'ALLWorkers' policy of success
	SuccessPolicyAllWorkers SuccessPolicy = "AllWorkers"
)
