/* Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package conf global config watcher
package conf

import (
	"context"
	"time"

	"gopkg.in/yaml.v2"
	"k8s.io/api/core/v1"

	"ascend-common/api"
	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/conf"
	"clusterd/pkg/domain/manualfault"
	"clusterd/pkg/interface/kube"
)

// WatchGlobalConfig watch global config from config cm
func WatchGlobalConfig(ctx context.Context) {
	const interval = 300 * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case _, ok := <-ctx.Done():
			if !ok {
				hwlog.RunLog.Info("stop signal chanel closed")
			}
			hwlog.RunLog.Info("watch global config from cm stop")
			return
		case <-ticker.C:
			TryLoadGlobalConfig()
		}
	}
}

func loadGlobalConfig(cm *v1.ConfigMap) {
	if cm.Data == nil {
		hwlog.RunLog.Errorf("cm <%s/%s> data is nil", api.ClusterNS, constant.ConfigCmName)
		return
	}
	data, ok := cm.Data[constant.ManuallySeparateNPUConfigKey]
	if !ok {
		hwlog.RunLog.Errorf("key %s is not found in cm <%s/%s>", constant.ManuallySeparateNPUConfigKey,
			api.ClusterNS, constant.ConfigCmName)
		return
	}
	var policy conf.ManuallySeparatePolicy
	if err := yaml.Unmarshal([]byte(data), &policy); err != nil {
		hwlog.RunLog.Errorf("unmarshal manually separate policy config failed from cm <%s/%s>, error: %v",
			api.ClusterNS, constant.ConfigCmName, err)
		return
	}
	if err := conf.Check(policy); err != nil {
		hwlog.RunLog.Errorf("check manually separate policy config failed, error: %v", err)
		return
	}
	conf.SetManualSeparatePolicy(policy)
	if !conf.GetManualEnabled() {
		manualfault.InitJobFaultManager(constant.DefaultSlidingWindow)
		manualfault.InitCounter()
		manualfault.InitFaultCmInfo()
	}
	hwlog.RunLog.Info("load manually separate policy config success")
}

// TryLoadGlobalConfig try load global config from cm
func TryLoadGlobalConfig() {
	const retryTime = 3
	for i := 0; i < retryTime; i++ {
		cm, err := kube.GetConfigMap(constant.ConfigCmName, api.ClusterNS)
		if err != nil {
			hwlog.RunLog.Errorf("get cm <%s/%s> info failed, error: %v", api.ClusterNS, constant.ConfigCmName, err)
			time.Sleep(1 * time.Second)
			continue
		}
		loadGlobalConfig(cm)
		break
	}
}
