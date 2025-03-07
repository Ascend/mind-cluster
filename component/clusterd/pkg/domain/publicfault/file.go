// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault file utils for public fault
package publicfault

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"clusterd/pkg/common/util"
)

var (
	// PubFaultResource allow public fault resource
	PubFaultResource []string
	// PubFaultCodeCfg public fault code configuration
	PubFaultCodeCfg pubFaultCodeCache
)

type pubFaultCfgFromFile struct {
	FaultCode pubFaultCode `json:"publicFaultCode"`
	Resource  []string     `json:"publicFaultResource"`
}

type pubFaultCode struct {
	NotHandleFaultCodes []string
	SubHealthFaultCodes []string
	SeparateNPUCodes    []string
}

type pubFaultCodeCache struct {
	NotHandleFaultCodes map[string]struct{}
	SubHealthFaultCodes map[string]struct{}
	SeparateNPUCodes    map[string]struct{}
}

func init() {
	PubFaultCodeCfg = pubFaultCodeCache{
		NotHandleFaultCodes: make(map[string]struct{}),
		SubHealthFaultCodes: make(map[string]struct{}),
		SeparateNPUCodes:    make(map[string]struct{}),
	}
}

// LoadPubFaultCfgFromFile load fault resource and fault code fault level from file
func LoadPubFaultCfgFromFile(filePath string) error {
	fileData, err := utils.LoadFile(filePath)
	if err != nil {
		hwlog.RunLog.Errorf("load fault config from <%s> failed, error: %v", filePath, err)
		return fmt.Errorf("load fault config from <%s> failed", filePath)
	}
	var pubFaultCfgFile pubFaultCfgFromFile
	if err = json.Unmarshal(fileData, &pubFaultCfgFile); err != nil {
		hwlog.RunLog.Errorf("unmarshal from <%s> failed, error: %v", filepath.Base(filePath), err)
		return fmt.Errorf("unmarshal from <%s> failed", filepath.Base(filePath))
	}

	PubFaultResource = util.RemoveDuplicates(pubFaultCfgFile.Resource)

	resetPubFaultCodeCache()
	// if one fault code corresponds to multiple fault levels, the most severe one will be dealt with
	for _, code := range util.RemoveDuplicates(pubFaultCfgFile.FaultCode.SeparateNPUCodes) {
		PubFaultCodeCfg.SeparateNPUCodes[code] = struct{}{}
	}
	for _, code := range util.RemoveDuplicates(pubFaultCfgFile.FaultCode.SubHealthFaultCodes) {
		PubFaultCodeCfg.SubHealthFaultCodes[code] = struct{}{}
	}
	for _, code := range util.RemoveDuplicates(pubFaultCfgFile.FaultCode.NotHandleFaultCodes) {
		PubFaultCodeCfg.NotHandleFaultCodes[code] = struct{}{}
	}
	hwlog.RunLog.Infof("load fault config from <%s> success", filepath.Base(filePath))
	return nil
}

func resetPubFaultCodeCache() {
	PubFaultCodeCfg.SeparateNPUCodes = make(map[string]struct{})
	PubFaultCodeCfg.SubHealthFaultCodes = make(map[string]struct{})
	PubFaultCodeCfg.NotHandleFaultCodes = make(map[string]struct{})
}
