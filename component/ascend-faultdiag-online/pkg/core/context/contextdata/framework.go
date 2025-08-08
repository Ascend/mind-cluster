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
Package contextdata 全局上下文信息
*/
package contextdata

import (
	"log"

	"ascend-faultdiag-online/pkg/core/config"
	"ascend-faultdiag-online/pkg/core/funchandler"
	"ascend-faultdiag-online/pkg/core/model"
)

// Framework 架构信息
type Framework struct {
	Config      *config.FaultDiagConfig         // 插件配置
	FuncHandler map[string]*funchandler.Handler // func handler map for algo_src
	ReqQue      chan *model.RequestContext      // 请求队列
	IsRunning   bool                            // 循环服务是否运行
	StopChan    chan struct{}                   // 停止信号
	Logger      *log.Logger                     // 日志记录器
}
