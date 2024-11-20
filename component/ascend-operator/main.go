/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.

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

package main

import (
	"context"
	"flag"
	"fmt"
	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	mindxdlv1 "ascend-operator/pkg/api/v1"
	"ascend-operator/pkg/controllers/v1"
)

const (
	defaultLogFileName = "/var/log/mindx-dl/ascend-operator/ascend-operator.log"
	defaultQPS         = 50.0
	defaultBurst       = 100
	maxQPS             = 10000.0
	maxBurst           = 10000
)

var (
	runtimeScheme        = runtime.NewScheme()
	hwLogConfig          = &hwlog.LogConfig{LogFileName: defaultLogFileName}
	version              bool
	enableGangScheduling bool
	// BuildVersion is the version of build package
	BuildVersion string
	// QPS to use while talking with kubernetes api-server
	QPS float64
	// Burst to use while talking with kubernetes api-server
	Burst int
)

func init() {
	utilruntime.Must(scheme.AddToScheme(runtimeScheme))
	utilruntime.Must(v1beta1.AddToScheme(runtimeScheme))
	utilruntime.Must(mindxdlv1.AddToScheme(runtimeScheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	flag.IntVar(&hwLogConfig.LogLevel, "logLevel", 0,
		"Log level, -1-debug, 0-info, 1-warning, 2-error, 3-critical(default 0)")
	flag.IntVar(&hwLogConfig.MaxAge, "maxAge", hwlog.DefaultMinSaveAge,
		"Maximum number of days for backup log files")
	flag.BoolVar(&hwLogConfig.IsCompress, "isCompress", false,
		"Whether backup files need to be compressed (default false)")
	flag.StringVar(&hwLogConfig.LogFileName, "logFile", defaultLogFileName, "Log file path")
	flag.IntVar(&hwLogConfig.MaxBackups, "maxBackups", hwlog.DefaultMaxBackups,
		"Maximum number of backup log files")
	flag.BoolVar(&enableGangScheduling, "enableGangScheduling", true,
		"Set true to enable gang scheduling")
	flag.Float64Var(&QPS, "kubeApiQps", defaultQPS, "QPS to use while talking with kubernetes api-server")
	flag.IntVar(&Burst, "kubeApiBurst", defaultBurst, "Burst to use while talking with kubernetes api-server")
	flag.BoolVar(&version, "version", false,
		"Query the verison of the program")

	flag.Parse()

	if version {
		fmt.Printf("ascend-operator version: %s\n", BuildVersion)
		return
	}

	if err := hwlog.InitRunLogger(hwLogConfig, context.Background()); err != nil {
		fmt.Printf("hwlog init failed, error is %v\n", err)
		return
	}

	hwlog.RunLog.Infof("ascend-operator starting and the version is %s", BuildVersion)
	mgr, err := ctrl.NewManager(initKubeConfig(), ctrl.Options{
		Scheme:             runtimeScheme,
		MetricsBindAddress: "0",
	})

	if err != nil {
		hwlog.RunLog.Errorf("unable to start manager: %s", err)
		return
	}

	if err = v1.NewReconciler(mgr, enableGangScheduling).SetupWithManager(mgr); err != nil {
		hwlog.RunLog.Errorf("unable to create ascend-controller err: %s", err)
		return
	}

	hwlog.RunLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		hwlog.RunLog.Errorf("problem running manager, err: %s", err)
		return
	}
}

func initKubeConfig() *rest.Config {
	kubeConfig := ctrl.GetConfigOrDie()

	if QPS <= 0 || QPS > maxQPS {
		hwlog.RunLog.Warnf("kubeApiQps is invalid, require (0, %f) use default value %f", maxQPS, defaultQPS)
		QPS = defaultQPS
	}
	if Burst <= 0 || Burst > maxBurst {
		hwlog.RunLog.Warnf("kubeApiBurst is invalid, require (0, %d) use default value %d", maxBurst, defaultBurst)
		Burst = defaultBurst
	}

	kubeConfig.QPS = float32(QPS)
	kubeConfig.Burst = Burst
	return kubeConfig
}
