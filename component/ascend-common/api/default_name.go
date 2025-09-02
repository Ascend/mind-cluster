// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package api common brand moniker
package api

// common
const (
	// Pod910DeviceAnno annotation value is for generating 910 hccl rank table
	Pod910DeviceAnno = "ascend.kubectl.kubernetes.io/ascend-910-configuration"

	// ResourceNamePrefix pre resource name
	ResourceNamePrefix = "huawei.com/"

	// Ascend910Lowercase for 910 chip Lowercase
	Ascend910Lowercase = "ascend910"

	// PodAnnotationAscendReal pod annotation ascend real
	PodAnnotationAscendReal = "huawei.com/AscendReal"
)

// common 910
const (
	// Ascend910 for 910 chip
	Ascend910 = "Ascend910"
)

// common 910 A2
const (
	// Ascend910B ascend 910B chip
	Ascend910B = "Ascend910B"
)

// common 910 A3
const (
	// Ascend910A3 ascend Ascend910A3 chip
	Ascend910A3 = "Ascend910A3"
)

// common 310
const (
	// Ascend310 ascend 310 chip
	Ascend310 = "Ascend310"
	// Ascend310No 310 chip number
	Ascend310No = "310"
)

// common 310B
const (
	// Ascend310B ascend 310B chip
	Ascend310B = "Ascend310B"
	// Ascend310BNo 310B chip number
	Ascend310BNo = "310B"
)

// common 310P
const (
	// Ascend310P ascend 310P chip
	Ascend310P = "Ascend310P"
	// Ascend310PNo 310P chip number
	Ascend310PNo = "310P"
)

// device plugin
const (
	// Use310PMixedInsert use 310P Mixed insert
	Use310PMixedInsert = "use310PMixedInsert"
	// Ascend310PMix dp use310PMixedInsert parameter usage
	Ascend310PMix = "ascend310P-V, ascend310P-VPro, ascend310P-IPro"
)

// npu exporter
const (
	// DevicePathPattern device path pattern
	DevicePathPattern = `^/dev/davinci\d+$`
	// HccsBWProfilingTimeStr  preset parameter name
	HccsBWProfilingTimeStr = "hccsBWProfilingTime"
	// Hccs log options domain value
	Hccs = "hccs"
	// Prefix pre statistic info
	Prefix = "npu_chip_info_hccs_statistic_info_"
	// BwPrefix pre bandwidth info
	BwPrefix = "npu_chip_info_hccs_bandwidth_info_"
	// AscendDeviceInfo
	AscendDeviceInfo = "ASCEND_VISIBLE_DEVICES"
)
