/* Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package kubeclient a series of k8s function
package kubeclient

import (
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	"Ascend-device-plugin/pkg/common"
	"ascend-common/common-utils/hwlog"
)

// InitPodInformer init pod informer
func (ki *ClientK8s) InitPodInformer() {
	factory := informers.NewSharedInformerFactoryWithOptions(ki.Clientset, 0,
		informers.WithTweakListOptions(func(options *v1.ListOptions) {
			options.FieldSelector = "spec.nodeName=" + ki.NodeName
		}))
	podInformer := factory.Core().V1().Pods().Informer()
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			UpdatePodList(nil, obj, EventTypeAdd)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if !reflect.DeepEqual(oldObj, newObj) {
				UpdatePodList(oldObj, newObj, EventTypeUpdate)
			}
		},
		DeleteFunc: func(obj interface{}) {
			UpdatePodList(nil, obj, EventTypeDelete)
		},
	})
	podInformer.AddEventHandler(ki.ResourceEventHandler(PodResource, checkPod))
	podInformer.SetWatchErrorHandler(func(r *cache.Reflector, err error) {
		hwlog.RunLog.Errorf("pod informer watch error: %v", err)
		if common.ParamOption.DealWatchHandler {
			ki.FlushPodCacheNextQuerying()
		}
	})
	factory.Start(make(chan struct{}))

	ki.PodInformer = podInformer
}

func checkPod(obj interface{}) bool {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return false
	}
	_, exist := pod.Annotations[common.HuaweiAscend910]
	return exist
}
