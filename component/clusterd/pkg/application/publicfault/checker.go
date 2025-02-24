// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package publicfault checker for public fault
package publicfault

import (
	"errors"
	"regexp"
	"strconv"
	"time"

	"ascend-common/api"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/util"
	"clusterd/pkg/domain/publicfault"
	"clusterd/pkg/domain/statistics"
)

const (
	validVersion   = "1.0"
	idStr          = "id"
	timeStr        = "time"
	descriptionStr = "description"
	nodeNameStr    = "nodeName"
	year2025       = 2025
)

var regexps = map[string]*regexp.Regexp{
	idStr:          regexp.MustCompile(`^[a-zA-Z0-9-_.]{8,128}$`),
	timeStr:        regexp.MustCompile(`^\d{10}$`),
	descriptionStr: regexp.MustCompile(`[\S ]{0,512}$`),
	nodeNameStr:    regexp.MustCompile(`^[a-z0-9]([a-z0-9-.]{0,251}[a-z0-9])?$`),
}

// NewPubFaultInfoChecker new pubFaultInfoChecker
func NewPubFaultInfoChecker(pubFaultInfo *api.PubFaultInfo) *pubFaultInfoChecker {
	return &pubFaultInfoChecker{pubFaultInfo: pubFaultInfo}
}

type pubFaultInfoChecker struct {
	pubFaultInfo *api.PubFaultInfo
}

// Check is used to check public fault parameters
func (c *pubFaultInfoChecker) Check() error {
	if c.pubFaultInfo == nil {
		return errors.New("public fault info is nil")
	}
	var checkFuncs = []func() error{
		c.checkId,
		c.checkTimeStamp,
		c.checkVersion,
		c.checkResource,
		c.checkFaults,
	}
	for _, checkFun := range checkFuncs {
		if err := checkFun(); err != nil {
			return err
		}
	}
	return nil
}

func (c *pubFaultInfoChecker) checkId() error {
	if !regexps[idStr].MatchString(c.pubFaultInfo.Id) {
		return errors.New("invalid id")
	}
	return nil
}

func (c *pubFaultInfoChecker) checkTimeStamp() error {
	if !regexps[timeStr].MatchString(strconv.Itoa(int(c.pubFaultInfo.TimeStamp))) {
		return errors.New("invalid timestamp")
	}
	minAvailTime := time.Date(year2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	if c.pubFaultInfo.TimeStamp < minAvailTime {
		return errors.New("invalid timestamp, can not before 2025/01/01T00:00:00Z")
	}
	return nil
}

func (c *pubFaultInfoChecker) checkVersion() error {
	if c.pubFaultInfo.Version != validVersion {
		return errors.New("invalid version")
	}
	return nil
}

func (c *pubFaultInfoChecker) checkResource() error {
	if !util.IsSliceContain(c.pubFaultInfo.Resource, publicfault.PubFaultResource) {
		return errors.New("invalid resource")
	}
	return nil
}

func (c *pubFaultInfoChecker) checkFaults() error {
	const (
		minFaultsLen = 1
		maxFaultsLen = 100
	)
	if len(c.pubFaultInfo.Faults) < minFaultsLen || len(c.pubFaultInfo.Faults) > maxFaultsLen {
		return errors.New("invalid faults length")
	}
	for _, fault := range c.pubFaultInfo.Faults {
		var checker = faultChecker{fault: &fault}
		if err := checker.check(); err != nil {
			return err
		}
	}
	return nil
}

type faultChecker struct {
	fault *api.Fault
}

func (c *faultChecker) check() error {
	if c.fault == nil {
		return errors.New("fault is nil")
	}
	var checkFuncs = []func() error{
		c.checkFaultId,
		c.checkFaultType,
		c.checkFaultCode,
		c.checkFaultTime,
		c.checkAssertion,
		c.checkFaultLocation,
		c.checkInfluence,
		c.checkDescription,
	}
	for _, checkFun := range checkFuncs {
		if err := checkFun(); err != nil {
			return err
		}
	}
	return nil
}

func (c *faultChecker) checkFaultId() error {
	if !regexps[idStr].MatchString(c.fault.FaultId) {
		return errors.New("invalid fault id")
	}
	return nil
}

func (c *faultChecker) checkFaultType() error {
	var allowFaultTypes = []string{constant.FaultTypeNPU, constant.FaultTypeNode,
		constant.FaultTypeNetwork, constant.FaultTypeStorage}
	if !util.IsSliceContain(c.fault.FaultType, allowFaultTypes) {
		return errors.New("invalid fault type")
	}
	return nil
}

func (c *faultChecker) checkFaultCode() error {
	const faultCodeLen = 9
	if len(c.fault.FaultCode) != faultCodeLen {
		return errors.New("invalid fault code")
	}
	faultLevel := publicfault.GetFaultLevelByCode(c.fault.FaultCode)
	if faultLevel == "" {
		return errors.New("invalid fault code, not exist in the configuration file")
	}
	return nil
}

func (c *faultChecker) checkFaultTime() error {
	if !regexps[timeStr].MatchString(strconv.Itoa(int(c.fault.FaultTime))) {
		return errors.New("invalid fault time")
	}
	minAvailTime := time.Date(year2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	if c.fault.FaultTime < minAvailTime {
		return errors.New("invalid fault time, can not before 2025/01/01T00:00:00Z")
	}
	return nil
}

func (c *faultChecker) checkAssertion() error {
	var allowAssertions = []string{constant.AssertionOccur, constant.AssertionRecover}
	if !util.IsSliceContain(c.fault.Assertion, allowAssertions) {
		return errors.New("invalid fault assertion")
	}
	return nil
}

func (c *faultChecker) checkFaultLocation() error {
	const (
		maxMapLen   = 10
		maxKeyLen   = 16
		maxValueLen = 128
	)
	if len(c.fault.FaultLocation) > maxMapLen {
		return errors.New("invalid fault location length")
	}
	for key, value := range c.fault.FaultLocation {
		if len(key) > maxKeyLen || len(value) > maxValueLen {
			return errors.New("invalid fault location key or value length")
		}
	}
	return nil
}

func (c *faultChecker) checkInfluence() error {
	const (
		minInfluenceLen = 1
		maxInfluenceLen = 1000
	)
	if len(c.fault.Influence) < minInfluenceLen || len(c.fault.Influence) > maxInfluenceLen {
		return errors.New("invalid influence length")
	}
	for _, influence := range c.fault.Influence {
		var checker = influenceChecker{influence: &influence}
		if err := checker.check(); err != nil {
			return err
		}
	}
	return nil
}

func (c *faultChecker) checkDescription() error {
	if !regexps[descriptionStr].MatchString(c.fault.Description) {
		return errors.New("invalid fault description")
	}
	return nil
}

type influenceChecker struct {
	influence *api.Influence
}

func (c *influenceChecker) check() error {
	if c.influence == nil {
		return errors.New("fault is nil")
	}
	var checkFuncs = []func() error{
		c.checkNodeNameOrSN,
		c.checkDeviceIds,
	}
	for _, checkFun := range checkFuncs {
		if err := checkFun(); err != nil {
			return err
		}
	}
	return nil
}

func (c *influenceChecker) checkNodeNameOrSN() error {
	if c.influence.NodeName != "" {
		if !regexps[nodeNameStr].MatchString(c.influence.NodeName) {
			return errors.New("invalid node name")
		}
		return nil
	}
	_, ok := statistics.GetNodeNameBySN(c.influence.NodeSN)
	if !ok {
		return errors.New("invalid node name and node sn, both them does not exist")
	}
	return nil
}

func (c *influenceChecker) checkDeviceIds() error {
	const (
		minDeviceIdLen = 1
		maxDeviceIdLen = 32
		minDeviceId    = 0
		maxDeviceId    = 31
	)
	if len(c.influence.DeviceIds) < minDeviceIdLen || len(c.influence.DeviceIds) > maxDeviceIdLen {
		return errors.New("invalid device id length")
	}
	if len(c.influence.DeviceIds) != len(util.RemoveDuplicates(c.influence.DeviceIds)) {
		return errors.New("invalid device id, duplicate ids")
	}
	for _, deviceId := range c.influence.DeviceIds {
		if deviceId < minDeviceId || deviceId > maxDeviceId {
			return errors.New("invalid device id")
		}
	}
	return nil
}
