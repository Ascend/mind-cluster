package common

import (
	"k8s.io/apimachinery/pkg/util/sets"
	"strconv"
)

const decimal = 10

// DeepEqualFaultDevInfo compare two FaultDevInfo
func DeepEqualFaultDevInfo(this, other *FaultDevInfo) bool {
	if this == nil && other == nil {
		return true
	}
	if this == nil || other == nil {
		return false
	}
	if this.NodeStatus != other.NodeStatus {
		return false
	}
	return faultDevListEqual(this.FaultDevList, other.FaultDevList)
}

type faultDevWithCodeSet struct {
	*FaultDev
	codeSet sets.String
}

func faultDevListEqual(thisList, otherList []*FaultDev) bool {
	if len(thisList) != len(otherList) {
		return false
	}
	thisMap := faultDevListToMap(thisList)
	otherMap := faultDevListToMap(otherList)
	if len(thisMap) != len(otherMap) {
		return false
	}
	for k, v1 := range thisMap {
		v2, ok := otherMap[k]
		if !ok {
			return false
		}
		if v1.FaultLevel != v2.FaultLevel {
			return false
		}
		if !v1.codeSet.Equal(v2.codeSet) {
			return false
		}
	}
	return true
}

func faultDevListToMap(list []*FaultDev) map[string]*faultDevWithCodeSet {
	m := make(map[string]*faultDevWithCodeSet, len(list))
	for _, dev := range list {
		m[dev.DeviceType+"/"+strconv.FormatInt(dev.DeviceId, decimal)] = &faultDevWithCodeSet{
			FaultDev: dev,
			codeSet:  sets.NewString(dev.FaultCode...),
		}
	}
	return m
}
