/*
Copyright(C)2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

/*
Package context is used to manage the global state and resources of the plugin.
*/
package context

import (
	"log"
	"os"

	"ascend-faultdiag-online/pkg/config"
	"ascend-faultdiag-online/pkg/context/metrics"
	"ascend-faultdiag-online/pkg/context/so_handle"
	"ascend-faultdiag-online/pkg/model/cluster"
	"ascend-faultdiag-online/pkg/model/node"
	"ascend-faultdiag-online/pkg/service/request"
)

type FaultDiagContext struct {
	Config        *config.FaultDiagConfig         // 插件配置
	SoHandlerMap  map[string]*so_handle.SoHandler // .so 文件处理器map
	ReqQue        chan *request.Context           // 请求队列
	IsRunning     bool                            // 循环服务是否运行
	StopChan      chan struct{}                   // 停止信号
	Logger        *log.Logger                     // 日志记录器
	MetricCtx     *metrics.Context                // 指标诊断上下文
	NodeStatus    *node.Status                    // 节点状态， node时使用
	ClusterStatus *cluster.Status                 // 集群状态， cluster时使用
}

func NewFaultDiagContext(config *config.FaultDiagConfig) (*FaultDiagContext, error) {
	soHandlerMap, err := so_handle.GenerateSoHandlerMap(config.SoDir)
	if err != nil {
		return nil, err
	}
	logger := log.New(os.Stdout, "[FaultDiag Online] ", log.LstdFlags)
	metricCtx := metrics.NewMetricContext()
	return &FaultDiagContext{
		Config:       config,
		SoHandlerMap: soHandlerMap,
		ReqQue:       make(chan *request.Context, config.QueueSize),
		StopChan:     make(chan struct{}),
		Logger:       logger,
		MetricCtx:    metricCtx,
	}, nil
}
