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

// Package cluster a series of function to watche configmap and do some operation
package cluster

import (
	"context"
	"fmt"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/utils/k8s"
)

// ConfigMapWatcher is a struct to watch the job summary ConfigMap
type ConfigMapWatcher struct {
	watchers   map[string]context.CancelFunc // key: namespace/cmName
	watchersMu sync.Mutex
}

var (
	cmWatcher *ConfigMapWatcher
	once      sync.Once
)

// GetConfigMapWatcher returns a singleton instance of ConfigMapWatcher
func GetConfigMapWatcher() *ConfigMapWatcher {
	once.Do(func() {
		cmWatcher = &ConfigMapWatcher{
			watchers: make(map[string]context.CancelFunc),
		}
	})
	return cmWatcher
}

func (cw *ConfigMapWatcher) keyGenerator(namespace, cmName string) string {
	return fmt.Sprintf("%s/%s", namespace, cmName)
}

// WatchConfigMap starts watching the ConfigMap by given namespace and cmName.
func (cw *ConfigMapWatcher) WatchConfigMap(namespace, cmName string) {
	var key = cw.keyGenerator(namespace, cmName)
	cw.watchersMu.Lock()
	defer cw.watchersMu.Unlock()

	if _, exists := cw.watchers[key]; exists {
		hwlog.RunLog.Infof("[FD-OL SLOWNODE]ConfigMap watcher for %s already exists, skipping", key)
		return // already watching this ConfigMap
	}

	ctx, cancel := context.WithCancel(context.Background())
	cw.watchers[key] = cancel
	go cw.runInformer(ctx, namespace, cmName)
}

func (cw *ConfigMapWatcher) runInformer(ctx context.Context, namespace, cmName string) {
	k8sClient, err := k8s.GetClient()
	if err != nil {
		hwlog.RunLog.Errorf("[FD-OL SLOWNODE]started configMap informer failed: got k8s client failed: %v", err)
		return
	}
	factory := informers.NewFilteredSharedInformerFactory(k8sClient.ClientSet, 0, namespace,
		func(options *metav1.ListOptions) {
			options.FieldSelector = fields.OneTermEqualSelector("metadata.name", cmName).String()
		},
	)
	informer := factory.Core().V1().ConfigMaps().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			processJobSummaryData(obj, watch.Added)
		},
		UpdateFunc: func(oldObj, newObj any) {
			processJobSummaryData(newObj, watch.Modified)
		},
		DeleteFunc: func(obj any) {
			processJobSummaryData(obj, watch.Deleted)
		},
	})

	stopChan := make(chan struct{})
	go func() {
		<-ctx.Done()
		close(stopChan)
		var key = cw.keyGenerator(namespace, cmName)
		delete(cw.watchers, key)
		hwlog.RunLog.Infof("[FD-OL SLOWNODE]ConfigMap watcher for %s/%s stopped", namespace, cmName)
	}()

	hwlog.RunLog.Infof("[FD-OL SLOWNODE]started to watch ConfigMap %s/%s", namespace, cmName)
	informer.Run(stopChan)
}

// StopWatching stops watching the ConfigMap by given namespace and cmName.
func (cw *ConfigMapWatcher) StopWatching(namespace, cmName string) {
	key := cw.keyGenerator(namespace, cmName)
	cw.watchersMu.Lock()
	defer cw.watchersMu.Unlock()

	if cancel, exists := cw.watchers[key]; exists {
		cancel() // stop the watcher
		delete(cw.watchers, key)
		hwlog.RunLog.Infof("[FD-OL SLOWNODE]Stopped watching ConfigMap %s/%s", namespace, cmName)
	} else {
		hwlog.RunLog.Warnf("[FD-OL SLOWNODE]No watcher found for ConfigMap %s/%s", namespace, cmName)
	}
}
