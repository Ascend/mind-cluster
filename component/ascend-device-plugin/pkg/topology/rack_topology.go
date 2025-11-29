/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package topology for generate topology of Rack
package topology

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/server"
	"ascend-common/api/slownet"
	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
)

var topoFilePathMap = map[int8]string{
	common.ProductTypeServer:    common.Server8PTopoPath,
	common.ProductType1D:        common.Pod1DTopoPath,
	common.ProductType2D:        common.Pod2DTopoPath,
	common.ProductType16PServer: common.Server16PTopoPath,
	common.ProductType32PServer: common.Server32PTopoPath,
}

var superPodId int32
var superPodType int8
var rackId int32
var serverIndex int32

// GetTopoFileAndWrite get topology file and write to to shared file
func GetTopoFileAndWrite(topoJsonFile string) {
	path, exist := topoFilePathMap[superPodType]
	if !exist {
		hwlog.RunLog.Errorf("super pod type:<%d> topo path not exist", superPodType)
		return
	}
	if err := ToFile(topoJsonFile, path); err != nil {
		hwlog.RunLog.Errorf("write topology info of RackID=%d SuperPodID=%d failed, err is %s",
			rackId, superPodId, err.Error())
	}
	hwlog.RunLog.Infof("write rack %d topology info to file %s success", rackId, topoJsonFile)

}

// RasTopoWriteTask write topology of rack to ras file path
func RasTopoWriteTask(ctx context.Context, hdm *server.HwDevManager) {
	if common.ParamOption.RealCardType != common.Ascend910A5 {
		hwlog.RunLog.Info("current is not a5, no need start RasTopoWriteTask")
		return
	}
	if hdm == nil {
		hwlog.RunLog.Error("illegal input, hdm is nil")
		return
	}
	_, err := slownet.GetRasNetRootPath()
	if err != nil {
		hwlog.RunLog.Errorf("get ras net root path failed, err: %v", err)
		return
	}
	hdm.ManagerLock.Lock()
	superPodId = hdm.GetSuperPodID()
	superPodType = hdm.GetSuperPodType()
	rackId = hdm.GetRackID()
	serverIndex = hdm.GetDevManager().GetServerIndex()
	hdm.ManagerLock.Unlock()
	for {
		select {
		case _, ok := <-ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel closed")
			}
			hwlog.RunLog.Info("RasTopoWriteTask stop")
			return
		default:
			if filePath, ok := checkConfigReady(fmt.Sprintf("%d", superPodId)); ok {
				hwlog.RunLog.Infof("ready to wirte topology file of rack to %s", filePath)
				GetTopoFileAndWrite(filePath)
			}
			// 30 seconds retry
			time.Sleep(common.TopologyRefreshTime * time.Second)
		}
	}
}

// return true means the netfault config is ok and generate topology file
// else return false do nothing and retry
func checkConfigReady(superPodIdStr string) (string, bool) {
	// check RAS_NET_ROOT_PATH env exist and get super-pod-x.json file path
	filePath, err := slownet.GetSuperPodInfoFilePath(superPodIdStr, publishCmNamePrefix)
	if err != nil {
		hwlog.RunLog.Errorf("get superpod topo file path err: %v", err)
		return "", false
	}
	// get super-pod-x dir and check its exist
	fileParentDir := filepath.Dir(filePath)
	if !utils.IsLexist(fileParentDir) {
		hwlog.RunLog.Infof("superpod topo file parent dir %s is not exist", fileParentDir)
		return "", false
	}
	hwlog.RunLog.Infof("the super-pod-x dir is exist in %s", fileParentDir)
	// ready to generate topology file of rack in ras file path so make sure the dir is exist with 755 mode
	topoFile, err := slownet.GetRackTopologyFilePath(superPodId, rackId,
		serverIndex)
	if err != nil {
		hwlog.RunLog.Errorf("get rack topo file path err: %v", err)
		return "", false
	}
	rackDir := filepath.Dir(topoFile)
	if !utils.IsLexist(rackDir) {
		hwlog.RunLog.Infof("%s topo path is not exist and will create it", rackDir)
		if mkErr := os.MkdirAll(rackDir, rackDirPerm); mkErr != nil {
			hwlog.RunLog.Errorf("create dir failed rack dir %s, err: %v", rackDir, mkErr)
			return "", false
		}
		hwlog.RunLog.Infof("create rack dir %s success", rackDir)
	}
	if chErr := os.Chmod(rackDir, rackDirPerm); chErr != nil {
		hwlog.RunLog.Errorf("change mod failed rack dir %s, err: %v", rackDir, chErr)
		return "", false
	}
	return topoFile, true
}
