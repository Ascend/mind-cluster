/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

package main

import (
	"context"
	"flag"
	"fmt"

	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
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
)

var (
	scheme               = runtime.NewScheme()
	hwLogConfig          = &hwlog.LogConfig{LogFileName: defaultLogFileName}
	version              bool
	enableGangScheduling bool
	BuildVersion         string
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1beta1.AddToScheme(scheme))
	utilruntime.Must(mindxdlv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
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
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
	})

	if err != nil {
		hwlog.RunLog.Errorf("unable to start manager: %s", err)
		return
	}

	if err = controllers.NewReconciler(mgr, enableGangScheduling).SetupWithManager(mgr); err != nil {
		hwlog.RunLog.Errorf("unable to create ascend-controller err: %s", err)
		return
	}

	hwlog.RunLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		hwlog.RunLog.Errorf("problem running manager, err: %s", err)
		return
	}
}
