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

// Package utils is to provide go runtime utils
package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/utils"
	"clusterd/pkg/interface/grpc/recover"
	"taskd/common/constant"
)

const maxReadBytes = 1024 * 1024

// InitHwLog init hwlog
func InitHwLog(logFileName string, ctx context.Context) error {
	var logFile string
	logFilePath := os.Getenv(constant.LogFilePathEnv)
	if logFilePath == "" {
		logFile = constant.DefaultLogFilePath + logFileName
	} else {
		logFile = filepath.Join(logFilePath, logFileName)
	}
	hwLogConfig := hwlog.LogConfig{
		LogFileName:   logFile,
		LogLevel:      constant.DefaultLogLevel,
		MaxBackups:    constant.DefaultMaxBackups,
		MaxAge:        constant.DefaultMaxAge,
		MaxLineLength: constant.DefaultMaxLineLength,
		// do not print to screen to avoid influence training log
		OnlyToFile: true,
	}
	if err := hwlog.InitRunLogger(&hwLogConfig, ctx); err != nil {
		fmt.Printf("hwlog init failed, error is %v\n", err)
		return err
	}
	return nil
}

func marshalData(data interface{}) []byte {
	dataBuffer, err := json.Marshal(data)
	if err != nil {
		hwlog.RunLog.Errorf("marshal data err: %v", err)
		return nil
	}
	return dataBuffer
}

// ObjToString obj to string
func ObjToString(data interface{}) string {
	var dataBuffer []byte
	if dataBuffer = marshalData(data); len(dataBuffer) == 0 {
		return ""
	}
	return string(dataBuffer)
}

// StringToObj string to obj
func StringToObj[T any](str string) (T, error) {
	var result T
	err := json.Unmarshal([]byte(str), &result)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal %s for type %s err: %v", str, reflect.TypeOf(result).Name(), err)
	}
	return result, err
}

// GetProfilingSwitch get profile switch status from file, if any fault happened return all switch off
func GetProfilingSwitch(filePath string) (constant.ProfilingSwitch, error) {
	data, err := utils.ReadLimitBytes(filePath, maxReadBytes)
	if err != nil {
		// if reading failed close all
		err := fmt.Errorf("failed to read file %s, err%v", filePath, err)
		return constant.ProfilingSwitch{
			CommunicationOperator: constant.SwitchOFF,
			Step:                  constant.SwitchOFF,
			SaveCheckpoint:        constant.SwitchOFF,
			FP:                    constant.SwitchOFF,
			DataLoader:            constant.SwitchOFF,
		}, err
	}

	var profiling constant.ProfilingSwitch

	err = json.Unmarshal(data, &profiling)
	if err != nil {
		err := fmt.Errorf("failed to parse profiling switch %#v: %v", profiling, err)
		return constant.ProfilingSwitch{
			CommunicationOperator: constant.SwitchOFF,
			Step:                  constant.SwitchOFF,
			SaveCheckpoint:        constant.SwitchOFF,
			FP:                    constant.SwitchOFF,
			DataLoader:            constant.SwitchOFF,
		}, err
	}
	return profiling, nil
}

// PfSwitchToPfDomainSwitch convert ProfilingSwitch to ProfilingDomainCmd
func PfSwitchToPfDomainSwitch(profilingSwitch constant.ProfilingSwitch) constant.ProfilingDomainCmd {
	profilingDomainCmd := constant.ProfilingDomainCmd{
		DefaultDomainAble: true,
		CommDomainAble:    false,
	}
	if profilingSwitch.Step == constant.SwitchOFF && profilingSwitch.SaveCheckpoint == constant.SwitchOFF &&
		profilingSwitch.FP == constant.SwitchOFF && profilingSwitch.DataLoader == constant.SwitchOFF &&
		profilingSwitch.CommunicationOperator == constant.SwitchOFF {
		profilingDomainCmd.DefaultDomainAble = false
	}
	if profilingSwitch.CommunicationOperator == constant.SwitchON {
		profilingDomainCmd.CommDomainAble = true
	}
	return profilingDomainCmd
}

// ProfilingResultToBizCode convert ProfilingResult to code
func ProfilingResultToBizCode(result constant.ProfilingResult) int32 {
	var code int32 = constant.ProfilingAllCloseCode
	switch result.DefaultDomain {
	case constant.ProfilingOnStatus:
		code += constant.ProfilingDefaultOpenInc
	case constant.ProfilingExpStatus:
		code += constant.ProfilingDefaultExpInc
	}

	switch result.CommDomain {
	case constant.ProfilingOnStatus:
		code += constant.ProfilingCommOpenInc
	case constant.ProfilingExpStatus:
		code += constant.ProfilingCommExpInc
	}
	return code
}

// BizCodeToProfilingCmd convert code to ProfilingDomainCmd
func BizCodeToProfilingCmd(code int32) (constant.ProfilingDomainCmd, error) {
	if code == constant.ProfilingAllCloseCmdCode {
		return constant.ProfilingDomainCmd{
			DefaultDomainAble: false,
			CommDomainAble:    false,
		}, nil
	}
	if code == constant.ProfilingDefaultDomainOnCode {
		return constant.ProfilingDomainCmd{
			DefaultDomainAble: true,
			CommDomainAble:    false,
		}, nil
	}
	if code == constant.ProfilingCommDomainOnCode {
		return constant.ProfilingDomainCmd{
			DefaultDomainAble: false,
			CommDomainAble:    true,
		}, nil
	}
	if code == constant.ProfilingAllOnCmdCode {
		return constant.ProfilingDomainCmd{
			DefaultDomainAble: true,
			CommDomainAble:    true,
		}, nil
	}
	return constant.ProfilingDomainCmd{}, fmt.Errorf("cannot convert code %d to ProfilingDomainCmd", code)
}

// GetOnesDigit get code ones digit num
func GetOnesDigit(code int32) int32 {
	return code % constant.Ten
}

// GetTensDigit get code tens digit num
func GetTensDigit(code int32) int32 {
	return code / constant.Ten % constant.Ten
}

// GetThousandsAndHundreds get thousands and hundreds num
func GetThousandsAndHundreds(code int32) int32 {
	return code / constant.Hundred * constant.Hundred
}

// CopyStringMap copy string map
func CopyStringMap(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// SliceContains judge slice contain some item
func SliceContains[T comparable](slice []T, one T) bool {
	for _, item := range slice {
		if item == one {
			return true
		}
	}
	return false
}

// ParseProfilingDomainCmd convert string to ProfilingDomainCmd
func ParseProfilingDomainCmd(defaultDomainCmd string, commDomainCmd string) (constant.ProfilingDomainCmd, error) {
	var switchOff = constant.ProfilingDomainCmd{
		DefaultDomainAble: false,
		CommDomainAble:    false,
	}
	defaultDomainOpen, err := strconv.ParseBool(defaultDomainCmd)
	if err != nil {
		return switchOff, fmt.Errorf("get DefaultDomainCmd %s err: %v", defaultDomainCmd, err)
	}
	commDomainOpen, err := strconv.ParseBool(commDomainCmd)
	if err != nil {
		return switchOff, fmt.Errorf("get commDomainCmd %s err: %v", commDomainCmd, err)
	}
	return constant.ProfilingDomainCmd{
		DefaultDomainAble: defaultDomainOpen,
		CommDomainAble:    commDomainOpen,
	}, nil
}

// ProfilingCmdToBizCode convert code to ProfilingDomainCmd
func ProfilingCmdToBizCode(cmd constant.ProfilingDomainCmd) int32 {
	if cmd.DefaultDomainAble && !cmd.CommDomainAble {
		return constant.ProfilingDefaultDomainOnCode
	}
	if !cmd.DefaultDomainAble && cmd.CommDomainAble {
		return constant.ProfilingCommDomainOnCode
	}
	if cmd.DefaultDomainAble && cmd.CommDomainAble {
		return constant.ProfilingAllOnCmdCode
	}
	return constant.ProfilingAllCloseCmdCode
}

// GetClusterdAddr get ClusterD addr
func GetClusterdAddr() (string, error) {
	proxyIp := os.Getenv(constant.LocalProxyEnableEnv)
	if proxyIp == constant.LocalProxyEnableOn {
		hwlog.RunLog.Infof("use proxy connect clusterd")
		return constant.LocalProxyIP + constant.ClusterdPort, nil
	}
	ipFromEnv := os.Getenv(constant.MindxServerIp)
	if err := utils.IsHostValid(ipFromEnv); err != nil {
		return "", err
	}
	return ipFromEnv + constant.ClusterdPort, nil
}

// GetFaultRanksMapByList get fault rank map by list
func GetFaultRanksMapByList(faultRanks []*pb.FaultRank) map[int]int {
	ranksMap := make(map[int]int)
	for _, faultRank := range faultRanks {
		rankIdInt, err := strconv.Atoi(faultRank.RankId)
		if err != nil {
			hwlog.RunLog.Warnf("convert rankId %s to int failed", faultRank.RankId)
			continue
		}
		typeInt, err := strconv.Atoi(faultRank.FaultType)
		if err != nil {
			hwlog.RunLog.Warnf("convert type %s to int failed", faultRank.FaultType)
			continue
		}
		ranksMap[rankIdInt] = typeInt
	}
	return ranksMap
}
