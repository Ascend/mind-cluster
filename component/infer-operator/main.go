/*
Copyright(C) 2026-2026. Huawei Technologies Co.,Ltd. All rights reserved.

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
	"os"
	"syscall"

	"k8s.io/api/admission/v1beta1"
	"k8s.io/api/apiserverinternal/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	volcanov1beta1 "volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	"ascend-common/common-utils/hwlog"
	v1 "infer-operator/pkg/api/v1"
	"infer-operator/pkg/common"
	util "infer-operator/pkg/common/client-go"
	"infer-operator/pkg/common/utils"
	"infer-operator/pkg/configManager"
	clusterctrlv1 "infer-operator/pkg/controller/v1"
	"infer-operator/pkg/controller/workload"
)

var (
	runtimeScheme = runtime.NewScheme()
	hwLogConfig   = &hwlog.LogConfig{}

	version bool
	// BuildVersion is the version of build package
	BuildVersion string
	// BuildName is the name of build package
	BuildName string
)

func init() {
	utilruntime.Must(apiextv1.AddToScheme(runtimeScheme))
	utilruntime.Must(v1alpha1.AddToScheme(runtimeScheme))
	utilruntime.Must(v1beta1.AddToScheme(runtimeScheme))
	utilruntime.Must(appsv1.AddToScheme(runtimeScheme))
	utilruntime.Must(corev1.AddToScheme(runtimeScheme))
	utilruntime.Must(scheme.AddToScheme(runtimeScheme))

	// Add Volcano PodGroup scheme
	utilruntime.Must(volcanov1beta1.AddToScheme(runtimeScheme))

	// Add CRD scheme
	utilruntime.Must(v1.AddToScheme(runtimeScheme))
}

func createCacheOptions() (cache.Options, error) {
	keyExistsRequirement, err := labels.NewRequirement(common.OperatorNameKey, selection.Exists, nil)
	if err != nil {
		return cache.Options{}, err
	}
	keyExistsSelector := labels.NewSelector().Add(*keyExistsRequirement)

	return cache.Options{
		Scheme: runtimeScheme,
		SelectorsByObject: map[client.Object]cache.ObjectSelector{
			&appsv1.StatefulSet{}: {
				Label: keyExistsSelector,
			},
			&appsv1.Deployment{}: {
				Label: keyExistsSelector,
			},
			&corev1.Service{}: {
				Label: keyExistsSelector,
			},
			&corev1.Pod{}: {
				Label: keyExistsSelector,
			},
			&corev1.ConfigMap{}: {
				Label: keyExistsSelector,
			},
			&volcanov1beta1.PodGroup{}: {
				Label: keyExistsSelector,
			},
		},
	}, nil
}

// parseFlags parses command line flags
func parseFlags() {
	flag.BoolVar(&version, "version", false, "Query the verison of the program")
	flag.IntVar(&hwLogConfig.LogLevel, "logLevel", 0,
		"Log level, -1-debug, 0-info, 1-warning, 2-error, 3-critical(default 0)")
	flag.IntVar(&hwLogConfig.MaxAge, "maxAge", hwlog.DefaultMinSaveAge,
		"Maximum number of days for backup log files")
	flag.BoolVar(&hwLogConfig.IsCompress, "isCompress", false,
		"Whether backup files need to be compressed (default false)")
	flag.StringVar(&hwLogConfig.LogFileName, "logFile",
		"/var/log/mindx-dl/infer-operator/infer-operator.log", "Log file path")
	flag.IntVar(&hwLogConfig.MaxBackups, "maxBackups", hwlog.DefaultMaxBackups,
		"Maximum number of backup log files")
	flag.Parse()
}

func main() {
	parseFlags()
	if version {
		fmt.Printf("%s version: %s\n", BuildName, BuildVersion)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err := hwlog.InitRunLogger(hwLogConfig, ctx); err != nil {
		fmt.Printf("unable to init run logger: %v\n", err)
		os.Exit(1)
	}
	go signalCatch(cancel)
	configMgr := configManager.NewConfigManager()
	configMgr.Start()
	cacheOption, err := createCacheOptions()
	if err != nil {
		hwlog.RunLog.Errorf("unable to create cache options: %v", err)
		os.Exit(1)
	}
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:   runtimeScheme,
		NewCache: cacheBuilder(cacheOption),
	})
	if err != nil {
		hwlog.RunLog.Errorf("unable to start manager: %v", err)
		os.Exit(1)
	}
	if err := util.InitInformers(ctx, mgr); err != nil {
		hwlog.RunLog.Errorf("unable to init informers: %v", err)
	}
	inferServiceSetReconciler := clusterctrlv1.NewInferServiceSetReconciler(mgr)
	if err := inferServiceSetReconciler.SetupWithManager(mgr); err != nil {
		hwlog.RunLog.Errorf("unable to setup infer service set reconciler: %v", err)
		os.Exit(1)
	}
	instanceSetReconciler := clusterctrlv1.NewInstanceSetReconciler(mgr, registerWorkLoadHandlersFunc())
	if err := instanceSetReconciler.SetupWithManager(ctx, mgr); err != nil {
		hwlog.RunLog.Errorf("unable to setup instance set reconciler: %v", err)
		os.Exit(1)
	}
	inferServiceReconciler := clusterctrlv1.NewInferServiceReconciler(mgr)
	if err := inferServiceReconciler.SetupWithManager(mgr); err != nil {
		hwlog.RunLog.Errorf("unable to setup infer service reconciler: %v", err)
		os.Exit(1)
	}
	hwlog.RunLog.Info("starting infer-controller manager")
	if err := mgr.Start(ctx); err != nil {
		hwlog.RunLog.Errorf("failed to start infer-controller manager: %v", err)
		os.Exit(1)
	}
}

func cacheBuilder(opts cache.Options) func(config *rest.Config, opts cache.Options) (cache.Cache, error) {
	return func(config *rest.Config, _ cache.Options) (cache.Cache, error) {
		return cache.New(config, opts)
	}
}

func registerWorkLoadHandlersFunc() clusterctrlv1.WorkloadRegister {
	return func(mgr ctrl.Manager, reconciler *workload.WorkLoadReconciler) {
		deploymentGVK := appsv1.SchemeGroupVersion.WithKind("Deployment")
		deploymentHandler := workload.NewDeploymentHandler(mgr.GetClient())
		reconciler.Register(deploymentGVK, deploymentHandler)

		statefulSetGVK := appsv1.SchemeGroupVersion.WithKind("StatefulSet")
		statefulSetHandler := workload.NewStatefulSetHandler(mgr.GetClient())
		reconciler.Register(statefulSetGVK, statefulSetHandler)
	}
}

func signalCatch(cancel context.CancelFunc) {
	osSignalChan := utils.NewSignalWatcher(syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	if osSignalChan == nil {
		hwlog.RunLog.Error("create stop signal channel failed")
		return
	}
	select {
	case sig, sigEnd := <-osSignalChan:
		if !sigEnd {
			hwlog.RunLog.Info("catch system stop signal channel is closed")
			return
		}
		hwlog.RunLog.Infof("receive system signal: %s, infer-operator shutting down", sig.String())
		cancel()
	}
}
