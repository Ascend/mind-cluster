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

// Package processmanager for plugin function
package processmanager

import (
	"errors"
	"time"

	"ascend-common/common-utils/hwlog"
	"nodeD/pkg/common"
	"nodeD/pkg/control/faultcontrol"
	"nodeD/pkg/kubeclient"
	"nodeD/pkg/monitoring/config"
	"nodeD/pkg/monitoring/ipmimonitor"
	"nodeD/pkg/reporter/cmreporter"
)

var (
	precessPluginMap map[string]Plugin
)

const (
	processNum = 3
	retryTime  = 3
)

// Plugin monitor、reporter、control plugin
type Plugin struct {
	// only one monitor for start event
	monitor   common.PluginMonitor
	reporters []common.PluginReporter
	controls  []common.PluginControl
}

// InitPlugin init process plugin
func InitPlugin() error {
	if kubeclient.GetK8sClient() == nil {
		return errors.New("k8s client is nil")
	}
	ipmiEventMonitor := ipmimonitor.NewIpmiEventMonitor()
	configmapEventMonitor := config.NewFaultConfigurator(kubeclient.GetK8sClient())
	nodeController := faultcontrol.NewNodeController()
	configMapReporter := cmreporter.NewConfigMapReporter(kubeclient.GetK8sClient())

	precessPluginMap = make(map[string]Plugin, processNum)
	ipmiPlugin := Plugin{
		monitor:   ipmiEventMonitor,
		controls:  []common.PluginControl{nodeController},
		reporters: []common.PluginReporter{configMapReporter},
	}
	precessPluginMap[common.IpmiProcess] = ipmiPlugin

	configPlugin := Plugin{
		monitor:  configmapEventMonitor,
		controls: []common.PluginControl{nodeController},
	}
	precessPluginMap[common.ConfigProcess] = configPlugin
	if err := startAllMonitor(); err != nil {
		return err
	}
	return nil
}

// GetMonitorPlugins get monitor plugins with process type
func GetMonitorPlugins(processType string) common.PluginMonitor {
	if pluginMonitor, ok := precessPluginMap[processType]; !ok {
		return nil
	} else {
		return pluginMonitor.monitor
	}
}

// GetControlPlugins get control plugins with process type
func GetControlPlugins(processType string) []common.PluginControl {
	if pluginControl, ok := precessPluginMap[processType]; !ok {
		return []common.PluginControl{}
	} else {
		return pluginControl.controls
	}
}

// GetReporterPlugins get Reporter plugins with process type
func GetReporterPlugins(processType string) []common.PluginReporter {
	if pluginReporter, ok := precessPluginMap[processType]; !ok {
		return []common.PluginReporter{}
	} else {
		return pluginReporter.reporters
	}
}

// GetAllProcessType get all process type
func GetAllProcessType() []string {
	return []string{common.IpmiProcess, common.ConfigProcess}
}

// GetAllLoopProcessType get all loop process type
func GetAllLoopProcessType() []string {
	return []string{common.IpmiProcess}
}

func startAllMonitor() error {
	errNum := 0
	for _, precessPlugin := range precessPluginMap {
		for i := 0; i < retryTime; i++ {
			if err := precessPlugin.monitor.Init(); err == nil {
				hwlog.RunLog.Infof("init monitor[%s] success", precessPlugin.monitor.Name())
				go precessPlugin.monitor.Monitoring()
				break
			} else if i+1 < retryTime {
				hwlog.RunLog.Errorf("init monitor[%s] failed, error: %v, retry count: %d",
					precessPlugin.monitor.Name(), err, i+1)
				time.Sleep(time.Second)
			} else {
				errNum++
			}
		}
	}
	if errNum == len(precessPluginMap) {
		return errors.New("all monitor init failed")
	}
	return nil
}
