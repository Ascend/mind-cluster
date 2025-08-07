// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.

// Package pingmesh a series of function handle ping mesh configmap create/update/delete
package pingmesh

import (
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/fdapi"
	"clusterd/pkg/common/constant"
)

// ConfigCollector the collector of fault network
func ConfigCollector(_, newInfo constant.ConfigPingMesh, operator string) {
	if operator == constant.AddOperator || operator == constant.UpdateOperator {
		updatePingMeshConfigCM(newInfo)
		return
	}
	hwlog.RunLog.Info("deleting pingmesh config will do nothing to controller")
}

func updatePingMeshConfigCM(newConfigInfo constant.ConfigPingMesh) {
	hwlog.RunLog.Info("ready to update pingmesh config")
	if isNeedToStop(newConfigInfo) {
		ConfigPingMeshInst.UpdateConfigData(newConfigInfo)
		rasNetDetectInst.Update(&constant.NetFaultInfo{NetFault: constant.RasNetDetectOff})
		fdapi.StopController()
		return
	}
	rasNetDetectInst.Update(&constant.NetFaultInfo{NetFault: constant.RasNetDetectOn})
	if err := ConfigPingMeshInst.UpdateConfig(newConfigInfo); err != nil {
		hwlog.RunLog.Errorf("update pingmesh config from cm failed, error :%s", err.Error())
	}
}

func isNeedToStop(newConfigInfo constant.ConfigPingMesh) bool {
	if newConfigInfo == nil || len(newConfigInfo) == 0 {
		return true
	}
	var retFlag = true
	for _, item := range newConfigInfo {
		if item != nil && item.Activate == constant.RasNetDetectOnStr {
			retFlag = false
			break
		}
	}
	if retFlag {
		hwlog.RunLog.Infof("all activate of the super-pod-x is %s, decide to stop", constant.RasNetDetectOffStr)
	}
	return retFlag
}
