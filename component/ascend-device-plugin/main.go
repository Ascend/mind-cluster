/* Copyright(C) 2022-2023. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package main implements initialization of the startup parameters of the device plugin.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/next/devicefactory"
	"Ascend-device-plugin/pkg/topology"
	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
)

const (
	// socket name
	defaultLogPath = "/var/log/mindx-dl/devicePlugin/devicePlugin.log"

	// defaultListWatchPeriod is the default listening device state's period
	defaultListWatchPeriod = 5

	// maxListWatchPeriod is the max listening device state's period
	maxListWatchPeriod = 1800
	// minListWatchPeriod is the min listening device state's period
	minListWatchPeriod = 3
	maxLogLineLength   = 1024

	// defaultLinkdownTimeout is the default linkdown timeout duration
	defaultLinkdownTimeout = 30
	// maxLinkdownTimeout is the max linkdown timeout duration
	maxLinkdownTimeout = 30
	// minLinkdownTimeout is the min linkdown timeout duration
	minLinkdownTimeout = 1
)

var (
	fdFlag          = flag.Bool("fdFlag", false, "Whether to use fd system to manage device (default false)")
	useAscendDocker = flag.Bool(api.UseAscendDocker, true, "Whether to use npu docker. "+
		"This parameter will be deprecated in future versions")
	volcanoType = flag.Bool("volcanoType", false,
		"Specifies whether to use volcano for scheduling ")
	version     = flag.Bool("version", false, "Output version information")
	edgeLogFile = flag.String("edgeLogFile", "/var/alog/AtlasEdge_log/devicePlugin.log",
		"Log file path in edge scene")
	listWatchPeriod = flag.Int("listWatchPeriod", defaultListWatchPeriod,
		"Listen and watch device state's period, unit second, range [3, 1800]")
	autoStowing = flag.Bool("autoStowing", true, "Whether to automatically stow the fixed device")
	logLevel    = flag.Int("logLevel", 0,
		"Log level, -1-debug, 0-info, 1-warning, 2-error, 3-critical(default 0)")
	logMaxAge = flag.Int("maxAge", common.MaxAge,
		"Maximum number of days for backup run log files, range [7, 700] days")
	logFile = flag.String("logFile", defaultLogPath,
		"The log file path, if the file size exceeds 20MB, will be rotate")
	logMaxBackups = flag.Int("maxBackups", common.MaxBackups,
		"Maximum number of backup log files, range is (0, 30]")
	presetVirtualDevice = flag.Bool("presetVirtualDevice", true, "Open the static of "+
		"computing power splitting function, only support "+api.Ascend910+" and "+api.Ascend310P)
	use310PMixedInsert = flag.Bool(api.Use310PMixedInsert, false, "Whether to use mixed insert "+
		api.Ascend310PMix+" card mode")
	hotReset = flag.Int("hotReset", -1, "set hot reset mode: -1-close, 0-infer, "+
		"1-train-online, 2-train-offline")
	shareDevCount = flag.Uint("shareDevCount", 1, "share device function, enable the func by setting "+
		"a value greater than 1, range is [1, 100], only support inference product")
	linkdownTimeout = flag.Int64("linkdownTimeout", defaultLinkdownTimeout, "linkdown timeout duration, "+
		", range [1, 30]")
	dealWatchHandler = flag.Bool("dealWatchHandler", false,
		"update pod cache when receiving pod informer watch errors")
	checkCachedPods = flag.Bool("checkCachedPods", true,
		"check pods in cache periodically, default true")
	enableSlowNode = flag.Bool("enableSlowNode", false,
		"switch of set slow node notice environment,default false")
	thirdPartyScanDelay = flag.Int("thirdPartyScanDelay", common.DefaultScanDelay,
		"delay time(second) before scanning devices reset by third party")
	deviceResetTimeout = flag.Int(api.DeviceResetTimeout, api.DefaultDeviceResetTimeout,
		"when device-plugin starts, if the number of chips is insufficient, the maximum duration to wait for "+
			"the driver to report all chips, unit second, range [10, 600]")
)

var (
	// BuildName show app name
	BuildName string
	// BuildVersion show app version
	BuildVersion string
	// BuildScene show app staring scene
	BuildScene string
)

func initLogModule(ctx context.Context) error {
	var loggerPath string
	loggerPath = *logFile
	if *fdFlag {
		loggerPath = *edgeLogFile
	}
	if !common.CheckFileUserSameWithProcess(loggerPath) {
		return fmt.Errorf("check log file failed")
	}
	hwLogConfig := hwlog.LogConfig{
		LogFileName:   loggerPath,
		LogLevel:      *logLevel,
		MaxBackups:    *logMaxBackups,
		MaxAge:        *logMaxAge,
		MaxLineLength: maxLogLineLength,
	}
	if err := hwlog.InitRunLogger(&hwLogConfig, ctx); err != nil {
		fmt.Printf("log init failed, error is %v\n", err)
		return err
	}
	return nil
}

func checkParam() bool {
	checks := []func() bool{
		checkListWatchPeriod,
		checkPresetAndVolcanoRelation,
		checkUse310PMixedInsertWithVolcano,
		checkUse310PMixedInsertWithShareDevCount,
		checkPresetWithShareDevCount,
		checkVolcanoWithShareDevCount,
		checkHotResetMode,
		checkBuildScene,
		checkLinkdownTimeout,
		checkThirdPartyScanDelay,
		checkDeviceResetTimeout,
		checkShareDevCount,
	}
	for _, check := range checks {
		if !check() {
			return false
		}
	}
	return true
}

func checkListWatchPeriod() bool {
	if *listWatchPeriod < minListWatchPeriod || *listWatchPeriod > maxListWatchPeriod {
		hwlog.RunLog.Errorf("list and watch period %d out of range", *listWatchPeriod)
		return false
	}
	return true
}

func checkPresetAndVolcanoRelation() bool {
	if !(*presetVirtualDevice) && !(*volcanoType) {
		hwlog.RunLog.Error("presetVirtualDevice is false, volcanoType should be true")
		return false
	}
	return true
}

func checkUse310PMixedInsertWithVolcano() bool {
	if *use310PMixedInsert && *volcanoType {
		hwlog.RunLog.Errorf("%s is true, volcanoType should be false", api.Use310PMixedInsert)
		return false
	}
	return true
}

func checkUse310PMixedInsertWithShareDevCount() bool {
	if *use310PMixedInsert && *shareDevCount > 1 {
		hwlog.RunLog.Errorf("%s is true, shareDevCount should be 1", api.Use310PMixedInsert)
		return false
	}
	return true
}

func checkPresetWithShareDevCount() bool {
	if !(*presetVirtualDevice) && *shareDevCount > 1 {
		hwlog.RunLog.Error("presetVirtualDevice is false, shareDevCount should be 1")
		return false
	}
	return true
}

func checkVolcanoWithShareDevCount() bool {
	if *volcanoType && *shareDevCount > 1 {
		hwlog.RunLog.Error("volcanoType is true, shareDevCount should be 1")
		return false
	}
	return true
}

func checkHotResetMode() bool {
	switch *hotReset {
	case common.HotResetClose, common.HotResetInfer, common.HotResetTrainOnLine, common.HotResetTrainOffLine:
		return true
	default:
		hwlog.RunLog.Error("hot reset mode param invalid")
		return false
	}
}

func checkBuildScene() bool {
	if BuildScene != common.EdgeScene && BuildScene != common.CenterScene {
		hwlog.RunLog.Error("unSupport build scene, only support edge and center")
		return false
	}
	return true
}

func checkLinkdownTimeout() bool {
	if (*linkdownTimeout) < minLinkdownTimeout || (*linkdownTimeout) > maxLinkdownTimeout {
		hwlog.RunLog.Warn("linkdown timeout duration out of range")
		return false
	}
	return true
}

func checkThirdPartyScanDelay() bool {
	if *thirdPartyScanDelay < 0 {
		hwlog.RunLog.Errorf("reset scan delay %v is invalid", *thirdPartyScanDelay)
		return false
	}
	return true
}

func checkDeviceResetTimeout() bool {
	if *deviceResetTimeout < api.MinDeviceResetTimeout || *deviceResetTimeout > api.MaxDeviceResetTimeout {
		hwlog.RunLog.Errorf("deviceResetTimeout %d out of range [%d,%d]", *deviceResetTimeout,
			api.MinDeviceResetTimeout, api.MaxDeviceResetTimeout)
		return false
	}
	return true
}

func checkShareDevCount() bool {
	if *shareDevCount < 1 || *shareDevCount > common.MaxShareDevCount {
		hwlog.RunLog.Error("share device function params invalid")
		return false
	}
	return true
}

func main() {
	flag.Parse()
	if *version {
		fmt.Printf("%s version: %s\n", BuildName, BuildVersion)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err := initLogModule(ctx); err != nil {
		return
	}
	if !checkParam() {
		return
	}
	hwlog.RunLog.Infof("device plugin starting and the version is %s", BuildVersion)
	hwlog.RunLog.Infof("device plugin starting scene is %s", BuildScene)
	setParameters()
	hdm, err := devicefactory.InitFunction()
	if err != nil {
		return
	}
	setUseAscendDocker()
	go hdm.ListenDevice(ctx)
	go hdm.ListenDpu(ctx)
	// start goroutine to dump topo of rack A5 for ras
	go topology.RasTopoWriteTask(ctx, hdm)
	hwlog.RunLog.Infof("device plugin started.")
	hdm.SignCatch(cancel)
}

func setParameters() {
	common.ParamOption = common.Option{
		GetFdFlag:           *fdFlag,
		UseAscendDocker:     *useAscendDocker,
		UseVolcanoType:      *volcanoType,
		AutoStowingDevs:     *autoStowing,
		ListAndWatchPeriod:  *listWatchPeriod,
		PresetVDevice:       *presetVirtualDevice,
		Use310PMixedInsert:  *use310PMixedInsert,
		HotReset:            *hotReset,
		BuildScene:          BuildScene,
		ShareCount:          *shareDevCount,
		LinkdownTimeout:     *linkdownTimeout,
		DealWatchHandler:    *dealWatchHandler,
		CheckCachedPods:     *checkCachedPods,
		EnableSlowNode:      *enableSlowNode,
		ThirdPartyScanDelay: *thirdPartyScanDelay,
		DeviceResetTimeout:  *deviceResetTimeout,
	}
}

func setUseAscendDocker() {
	*useAscendDocker = true
	ascendDocker := os.Getenv(api.AscendDockerRuntimeEnv)
	if ascendDocker != "True" {
		*useAscendDocker = false
		hwlog.RunLog.Debugf("get docker runtime from env is: %#v", ascendDocker)
	}
	if common.ParamOption.Use310PMixedInsert {
		*useAscendDocker = false
		hwlog.RunLog.Debugf("mixed insert mode do not use npu docker")
	}
	if len(common.ParamOption.ProductTypes) == 1 && common.ParamOption.ProductTypes[0] == common.Atlas200ISoc {
		*useAscendDocker = false
	}

	common.ParamOption.UseAscendDocker = *useAscendDocker
	hwlog.RunLog.Infof("device-plugin set npu docker as: %v", *useAscendDocker)
}
