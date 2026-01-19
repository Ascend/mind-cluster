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
	"context"
	"errors"
	"time"

	"ascend-common/common-utils/hwlog"
	"nodeD/pkg/common"
	"nodeD/pkg/control/dpccontrol"
	"nodeD/pkg/control/faultcontrol"
	"nodeD/pkg/kubeclient"
	"nodeD/pkg/monitoring/config"
	"nodeD/pkg/monitoring/dpcmonitor"
	"nodeD/pkg/monitoring/ipmimonitor"
	"nodeD/pkg/reporter/cmreporter"
	"nodeD/pkg/reporter/publicfault"
)

var (
	processPluginMap map[string]Plugin = nil
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
func InitPlugin(ctx context.Context) error {
	if kubeclient.GetK8sClient() == nil {
		return errors.New("k8s client is nil")
	}
	ipmiEventMonitor := ipmimonitor.NewIpmiEventMonitor()
	configmapEventMonitor := config.NewFaultConfigurator(kubeclient.GetK8sClient())
	dpcEventMonitor := dpcmonitor.NewDpcEventMonitor(ctx)

	nodeController := faultcontrol.NewNodeController()
	dpcController := dpccontrol.NewDpcController()

	configMapReporter := cmreporter.NewConfigMapReporter(kubeclient.GetK8sClient())
	pfReporter := publicfault.NewGrpcReporter()

	processPluginMap = make(map[string]Plugin, processNum)
	processPluginMap[common.IpmiProcess] = Plugin{
		monitor:   ipmiEventMonitor,
		controls:  []common.PluginControl{nodeController},
		reporters: []common.PluginReporter{configMapReporter},
	}
	processPluginMap[common.ConfigProcess] = Plugin{
		monitor:  configmapEventMonitor,
		controls: []common.PluginControl{nodeController},
	}
	processPluginMap[common.DpcProcess] = Plugin{
		monitor:   dpcEventMonitor,
		controls:  []common.PluginControl{dpcController},
		reporters: []common.PluginReporter{pfReporter},
	}
	if err := startAllMonitor(); err != nil {
		return err
	}
	return nil
}

// GetMonitorPlugins get monitor plugins with process type
func GetMonitorPlugins(processType string) common.PluginMonitor {
	if pluginMonitor, ok := processPluginMap[processType]; !ok {
		return nil
	} else {
		return pluginMonitor.monitor
	}
}

// GetControlPlugins get control plugins with process type
func GetControlPlugins(processType string) []common.PluginControl {
	if pluginControl, ok := processPluginMap[processType]; !ok {
		return []common.PluginControl{}
	} else {
		return pluginControl.controls
	}
}

// GetReporterPlugins get Reporter plugins with process type
func GetReporterPlugins(processType string) []common.PluginReporter {
	if pluginReporter, ok := processPluginMap[processType]; !ok {
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
	for _, processPlugin := range processPluginMap {
		for i := 0; i < retryTime; i++ {
			if err := processPlugin.monitor.Init(); err == nil {
				hwlog.RunLog.Infof("init monitor[%s] success", processPlugin.monitor.Name())
				go processPlugin.monitor.Monitoring()
				break
			} else if i+1 < retryTime {
				hwlog.RunLog.Errorf("init monitor[%s] failed, error: %v, retry count: %d",
					processPlugin.monitor.Name(), err, i+1)
				time.Sleep(time.Second)
			} else {
				errNum++
			}
		}
	}
	if errNum == len(processPluginMap) {
		return errors.New("all monitor init failed")
	}
	return nil
}
