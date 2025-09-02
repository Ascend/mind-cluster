// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package api common brand moniker
package api

// common
const (
	// Pod910DeviceAnno annotation value is for generating 910 hccl rank table
	Pod910DeviceAnno = "ascend.kubectl.kubernetes.io/ascend-910-configuration"

	// ResourceNamePrefix pre resource name
	ResourceNamePrefix = "huawei.com/"

	// Ascend910 for 910 chip
	Ascend910 = "Ascend910"

	// Ascend910Lowercase for 910 chip Lowercase
	Ascend910Lowercase = "ascend910"

	// PodAnnotationAscendReal pod annotation ascend real
	PodAnnotationAscendReal = "huawei.com/AscendReal"
)

// device plugin
const (
	// Use310PMixedInsert use 310P Mixed insert
	Use310PMixedInsert = "use310PMixedInsert"
	// Ascend310PMix dp use310PMixedInsert parameter usage
	Ascend310PMix = "ascend310P-V, ascend310P-VPro, ascend310P-IPro"
)
