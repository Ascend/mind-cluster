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
Package configmap is using for watching configmap changes.
*/

package configmap

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"ascend-common/common-utils/hwlog"
	"nodeD/pkg/kubeclient"
)

const (
	maxRequeueTimes = 10
)

// Option is a function that configures a configMapWatcher
type Option func(*configMapWatcher)

type handler func(cm *corev1.ConfigMap)

// NamedHandler is a handler with a name
type NamedHandler struct {
	Name   string
	Handle handler
}

// WithLabelSector sets the label selector for the configmap watcher
func WithLabelSector(labelSelector string) Option {
	return func(cw *configMapWatcher) {
		cw.labelSelector = labelSelector
	}
}

// WithNamespace sets the namespace for the configmap watcher
func WithNamespace(namespace string) Option {
	return func(cw *configMapWatcher) {
		cw.namespace = namespace
	}
}

// WithNamedHandlers sets the named handlers for the configmap watcher
func WithNamedHandlers(handlers ...NamedHandler) Option {
	return func(cw *configMapWatcher) {
		for _, h := range handlers {
			cw.handlers[h.Name] = h
		}
	}
}

// NewWatcher creates a new configmap watcher
func NewWatcher(client *kubeclient.ClientK8s, options ...Option) *configMapWatcher {
	cw := &configMapWatcher{
		client:   client,
		handlers: make(map[string]NamedHandler),
		queue:    workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
	}
	for _, opt := range options {
		opt(cw)
	}

	return cw
}

type configMapWatcher struct {
	client          *kubeclient.ClientK8s
	labelSelector   string
	namespace       string
	handlers        map[string]NamedHandler
	informerFactory informers.SharedInformerFactory
	indexer         cache.Indexer
	queue           workqueue.RateLimitingInterface
}

// Init initialize the configmap watcher
func (cw *configMapWatcher) Init() {
	informerFactory := informers.NewSharedInformerFactoryWithOptions(cw.client.ClientSet, 0,
		informers.WithNamespace(cw.namespace), informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = cw.labelSelector
		}))

	informerFactory.Core().V1().ConfigMaps().Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			cm, ok := obj.(*corev1.ConfigMap)
			if !ok {
				return false
			}
			_, ok = cw.handlers[cm.Name]
			return ok
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err != nil {
					return
				}
				hwlog.RunLog.Infof("add configmap %s", key)
				cw.queue.Add(key)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(newObj)
				if err != nil {
					return
				}
				hwlog.RunLog.Infof("update configmap %s", key)
				cw.queue.Add(key)
			},
		},
	})
	cw.informerFactory = informerFactory
	cw.indexer = informerFactory.Core().V1().ConfigMaps().Informer().GetIndexer()
}

// Watch is a function that watches configmap changes
func (cw *configMapWatcher) Watch(stopCh <-chan struct{}) {
	go cw.informerFactory.Start(stopCh)
	cache.WaitForCacheSync(stopCh, cw.informerFactory.Core().V1().ConfigMaps().Informer().HasSynced)

	for cw.processNextItem() {
	}
}

func (cw *configMapWatcher) processNextItem() bool {
	obj, shutdown := cw.queue.Get()
	if shutdown {
		return false
	}
	defer cw.queue.Done(obj)

	namespacedName, ok := obj.(string)
	if !ok {
		cw.queue.Forget(obj)
		return true
	}
	hwlog.RunLog.Infof("process configmap %s", namespacedName)
	cmObj, exists, err := cw.indexer.GetByKey(namespacedName)
	if err != nil {
		cw.queue.Forget(obj)
		return true
	}
	if !exists {
		if cw.queue.NumRequeues(obj) < maxRequeueTimes {
			cw.queue.AddRateLimited(obj)
			return true
		}
		cw.queue.Forget(obj)
		return true
	}
	defer cw.queue.Forget(obj)
	cm, ok := cmObj.(*corev1.ConfigMap)
	if !ok {
		return true
	}
	if h, ok := cw.handlers[cm.Name]; ok {
		h.Handle(cm)
	}
	return true
}
