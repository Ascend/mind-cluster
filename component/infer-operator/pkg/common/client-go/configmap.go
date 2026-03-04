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

package util

import (
	"context"
	"os"
	"path/filepath"

	"k8s.io/api/core/v1"
	toolscache "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	"ascend-common/common-utils/hwlog"
	"infer-operator/pkg/common"
)

// inferOperatorCfgEventFuncs stores event callback functions for infer-operator-config configmap
var inferOperatorCfgEventFuncs = []func(interface{}, interface{}, string){}

// getCurrentNamespace gets the namespace of the current pod
func getCurrentNamespace() string {
	// Get namespace from service account file
	if namespace, err := os.ReadFile(filepath.Clean(common.NamespacePath)); err == nil {
		return string(namespace)
	}
	hwlog.RunLog.Warnf("Failed to get current namespace from %s, using default: %s", common.NamespacePath, common.DefaultNamespace)
	return common.DefaultNamespace
}

// AddInferOperatorCfgEventFuncs adds event callback functions for infer-operator-config configmap
func AddInferOperatorCfgEventFuncs(func1 ...func(interface{}, interface{}, string)) {
	inferOperatorCfgEventFuncs = append(inferOperatorCfgEventFuncs, func1...)
}

// InitCMInformer initializes configmap informer
func InitCMInformer(ctx context.Context, informers cache.Cache) error {
	cmGVK := v1.SchemeGroupVersion.WithKind("ConfigMap")
	cmInformer, err := informers.GetInformerForKind(ctx, cmGVK)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get ConfigMap informer: %v", err)
		return err
	}
	addInferOperatorCfgEventHandler(&cmInformer)
	return nil
}

func filterInferOperatorCfg(obj interface{}) bool {
	cm, ok := obj.(*v1.ConfigMap)
	if !ok {
		return false
	}
	return cm.Name == common.InferOperatorCfgName && cm.Namespace == getCurrentNamespace()
}

func addInferOperatorCfgEventHandler(cmInformer *cache.Informer) {
	(*cmInformer).AddEventHandler(toolscache.FilteringResourceEventHandler{
		FilterFunc: filterInferOperatorCfg,
		Handler: toolscache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				inferOperatorCfgHandler(nil, obj, common.AddOperator)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				inferOperatorCfgHandler(oldObj, newObj, common.UpdateOperator)
			},
			DeleteFunc: func(obj interface{}) {
				inferOperatorCfgHandler(nil, obj, common.DeleteOperator)
			},
		},
	})
}

func inferOperatorCfgHandler(oldObj interface{}, newObj interface{}, operator string) {
	for _, eventFunc := range inferOperatorCfgEventFuncs {
		eventFunc(oldObj, newObj, operator)
	}
}
