package common

import (
	"fmt"
	"testing"
)

func TestGetAndCleanRemovedReasonEvent(t *testing.T) {
	InsertUpgradeFaultCache(1, 12345, "faultcode", ManuallySeparateNPU)
	cm, _ := upgradeFaultCacheMgr.cache.ConvertCacheToCm(func(i int32) (int32, error) {
		return i, nil
	})
	toString := cm.CmToString("device")
	fmt.Println(toString)
	RemoveManuallySeparateReasonCache([]LogicId{1})
	cm, _ = upgradeFaultCacheMgr.cache.ConvertCacheToCm(func(i int32) (int32, error) {
		return i, nil
	})
	toString = cm.CmToString("device")
	fmt.Println(toString)
}
