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

// Package proxy for taskd proxy backend
package proxy

import (
	"context"
	"errors"
	"fmt"

	"ascend-common/common-utils/hwlog"
	"taskd/common/constant"
	"taskd/common/utils"
	"taskd/toolkit_backend/net"
	"taskd/toolkit_backend/net/common"
)

const (
	localHost = "127.0.0.1"
)

type proxyClient struct {
	networkInstence *net.NetInstance
	proxyInfo       *proxyInfo
	proxyLogger     *hwlog.CustomLogger
}

type proxyInfo struct {
	proxyConfig *common.TaskNetConfig
}

var proxyInstance *proxyClient

func newProxyInstance(proxyConfig *common.TaskNetConfig, logger *hwlog.CustomLogger) error {
	var err error
	if proxyInstance != nil {
		return nil
	}
	proxyInstance = &proxyClient{
		networkInstence: nil,
		proxyInfo:       &proxyInfo{},
		proxyLogger:     logger,
	}
	err = proxyInstance.initNetwork(proxyConfig)
	return err
}

func (p *proxyClient) validNetConfig(proxyConfig *common.TaskNetConfig) error {
	if proxyConfig.Pos.ProcessRank != "-1" {
		p.proxyLogger.Errorf("proxyClient validNetConfig failed, ProcessRank is not -1")
		return errors.New("proxyClient validNetConfig failed, ProcessRank is not -1")
	}

	return nil
}

func (p *proxyClient) initNetwork(proxyConfig *common.TaskNetConfig) error {
	var err error
	err = p.validNetConfig(proxyConfig)
	if err != nil {
		return err
	}
	p.proxyInfo.proxyConfig = proxyConfig
	p.networkInstence, err = net.InitNetwork(proxyConfig, p.proxyLogger)
	if err == nil {
		p.proxyLogger.Info("proxyClient InitNetwork succeed.")
		return nil
	}
	p.proxyLogger.Errorf("proxyClient InitNetwork failed:%v.", err)
	return err
}

func (p *proxyClient) destroyNet() {
	p.networkInstence.Destroy()
}

// InitProxy init proxy grpc
func InitProxy(proxyConfig *common.TaskNetConfig) error {
	if proxyConfig == nil {
		return errors.New("proxyConfig is nil")
	}
	logName := fmt.Sprintf(constant.ProxyLogPathPattern, proxyConfig.Pos.ServerRank)
	hwLogConfig, err := utils.GetLoggerConfigWithFileName(logName)
	if err != nil {
		fmt.Printf("hwlog init failed, error is %v\n", err)
		return err
	}
	var logger *hwlog.CustomLogger
	logger, err = hwlog.NewCustomLogger(hwLogConfig, context.Background())
	if err != nil {
		fmt.Printf("hwlog init failed, error is %v\n", err)
		return err
	}
	err = newProxyInstance(proxyConfig, logger)
	if err != nil {
		logger.Errorf("InitProxy failed:%s.", err)
		return err
	}
	logger.Info("InitProxy success.")
	return nil
}

// DestroyProxy destroy proxy
func DestroyProxy() {
	if proxyInstance == nil {
		return
	}
	proxyInstance.destroyNet()
	proxyInstance = nil
}
