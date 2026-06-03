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

package builtin

import (
	"fmt"

	"Ascend-device-plugin/pkg/kubeclient"
	"Ascend-device-plugin/pkg/plugin"
	"ascend-common/common-utils/hwlog"
	"ascend-common/devmanager"
)

func InitPluginManager(dmgr devmanager.DeviceInterface,
	kubeClient *kubeclient.ClientK8s) (*plugin.PluginManager, error) {
	pm := plugin.NewPluginManager()
	if err := pm.RegisterPlugin(NewOutBandResetPlugin(dmgr)); err != nil {
		return nil, fmt.Errorf("register outbandReset plugin failed: %w", err)
	}
	if err := pm.RegisterPlugin(NewResetRecordPlugin(kubeClient)); err != nil {
		return nil, fmt.Errorf("register resetRecord plugin failed: %w", err)
	}
	if err := pm.Init(); err != nil {
		return nil, fmt.Errorf("init plugin manager failed: %w", err)
	}
	hwlog.RunLog.Infof("plugin manager initialized with %d plugins", len(pm.Plugins))
	return pm, nil
}
