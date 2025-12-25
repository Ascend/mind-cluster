/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

package v1

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/kubeflow/common/pkg/controller.v1/control"
	"github.com/kubeflow/common/pkg/util"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	kubeclientset "k8s.io/client-go/kubernetes"
	schedulinglisters "k8s.io/client-go/listers/scheduling/v1beta1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"volcano.sh/apis/pkg/client/clientset/versioned"

	"ascend-common/common-utils/hwlog"
	"rl-operator/pkg/common"
)

type JobReconciler struct {
	client.Client
	ApiReader     client.Reader
	Scheme        *runtime.Scheme
	Mutex         sync.RWMutex
	Versions      map[types.UID]int32
	BackoffLimits map[types.UID]int32
	ConfigInfo    JobConfigInfo
	Recorder      record.EventRecorder
	// PodControl is reused to add or delete pods.
	PodControl control.PodControlInterface
	// VolcanoControl is used to add or delete resources of volcano.
	VolcanoControl VolcanoControlInterface
	// KubeClientSet is a standard kubernetes clientset.
	KubeClientSet kubeclientset.Interface
	// PriorityClassLister can list/get priorityClasses from the shared informer's store.
	PriorityClassLister schedulinglisters.PriorityClassLister
}

type JobConfigInfo interface {
	GetAPIGroupVersionKind() schema.GroupVersionKind
	GetAPIGroupVersion() schema.GroupVersion
	GetRunPolicy(object client.Object) (*commonv1.RunPolicy, error)
	GetStatus(object client.Object) (*commonv1.JobStatus, error)
}

// NewJobReconciler new base reconciler for all Resource in rl-operator
func NewJobReconciler(mgr manager.Manager, controllerName string) *JobReconciler {
	cfg := mgr.GetConfig()
	recorder := mgr.GetEventRecorderFor(controllerName)
	kubeClientSet := kubernetes.NewForConfigOrDie(cfg)
	volcanoClientSet := versioned.NewForConfigOrDie(cfg)
	sharedInformers := informers.NewSharedInformerFactory(kubeClientSet, 0)
	priorityClassInformer := sharedInformers.Scheduling().V1beta1().PriorityClasses()
	baseReconciler := &JobReconciler{
		Client:              mgr.GetClient(),
		Scheme:              mgr.GetScheme(),
		ApiReader:           mgr.GetAPIReader(),
		Versions:            make(map[types.UID]int32),
		BackoffLimits:       make(map[types.UID]int32),
		Recorder:            recorder,
		KubeClientSet:       kubeClientSet,
		PodControl:          control.RealPodControl{KubeClient: kubeClientSet, Recorder: recorder},
		VolcanoControl:      &RealVolcanoControl{VolcanoClient: volcanoClientSet, Recorder: recorder},
		PriorityClassLister: priorityClassInformer.Lister(),
	}
	baseReconciler.ConfigInfo = baseReconciler

	return baseReconciler
}

func (r *JobReconciler) onOwnerCreateFunc() func(event.CreateEvent) bool {
	return func(e event.CreateEvent) bool {
		r.Mutex.Lock()
		defer r.Mutex.Unlock()
		r.Versions[e.Object.GetUID()] = common.DefaultPodVersion
		r.BackoffLimits[e.Object.GetUID()] = common.UnsetBackoffLimits
		runPolicy, err := r.ConfigInfo.GetRunPolicy(e.Object)
		if err != nil {
			hwlog.RunLog.Errorf("Failed to get RunPolicy of %v: %v", reflect.TypeOf(e.Object), err)
			return false
		}
		if runPolicy.BackoffLimit != nil {
			r.BackoffLimits[e.Object.GetUID()] = *runPolicy.BackoffLimit
		}
		msg := fmt.Sprintf("Job %s is create.", e.Object.GetName())
		hwlog.RunLog.Info(msg)
		jobStatus, err := r.ConfigInfo.GetStatus(e.Object)
		if err != nil {
			hwlog.RunLog.Errorf("Failed to get Job Status of %s: %v", reflect.TypeOf(e.Object).Name(), err)
			return false
		}

		err = util.UpdateJobConditions(jobStatus, commonv1.JobCreated, util.JobCreatedReason, msg)
		if err != nil {
			log.Log.Error(err, "append job condition error")
			return false
		}
		return true
	}
}

func (r *JobReconciler) onOwnerDeleteFunc() func(deleteEvent event.DeleteEvent) bool {
	return func(e event.DeleteEvent) bool {
		msg := fmt.Sprintf("%v %s is deleted.", reflect.TypeOf(e.Object), e.Object.GetName())
		hwlog.RunLog.Info(msg)
		r.Mutex.Lock()
		delete(r.Versions, e.Object.GetUID())
		delete(r.BackoffLimits, e.Object.GetUID())
		r.Mutex.Unlock()
		return true
	}
}

// onPodDeleteFunc does some necessary processing logic when a pod is deleted.
func (r *JobReconciler) onPodDeleteFunc() func(event.DeleteEvent) bool {
	return func(e event.DeleteEvent) bool {
		controllerRef := metav1.GetControllerOf(e.Object)
		replicaType, ok := e.Object.GetLabels()[commonv1.ReplicaTypeLabel]
		if !ok || len(replicaType) == 0 {
			return false
		}
		version, ok := e.Object.GetLabels()[common.PodVersionLabel]
		if !ok || len(version) == 0 {
			return false
		}
		versionNumber, err := strconv.Atoi(version)
		if err != nil {
			hwlog.RunLog.Errorf("failed to convert string to int, err: %v", err)
			return false
		}
		hwlog.RunLog.Infof("deleted pod <%s> version is: %v", e.Object.GetName(), version)
		if controllerRef == nil {
			return true
		}
		r.Mutex.RLock()
		currentVersion, ok := r.Versions[controllerRef.UID]
		r.Mutex.RUnlock()
		if ok && int32(versionNumber) == currentVersion {
			r.Mutex.Lock()
			r.Versions[controllerRef.UID]++
			r.Mutex.Unlock()
		}
		return true
	}
}

func (r *JobReconciler) IsReachBackoffLimit(job metav1.Object) bool {
	r.Mutex.RLock()
	version, ok := r.Versions[job.GetUID()]
	backoffLimit, backoffLimitOk := r.BackoffLimits[job.GetUID()]
	r.Mutex.RUnlock()
	if !ok || (backoffLimitOk && backoffLimit > 0 && version >= backoffLimit) {
		return true
	}
	return false
}

func (r *JobReconciler) FetchResource(
	ctx context.Context,
	namespacedName types.NamespacedName,
	resource client.Object) error {
	if err := r.Get(ctx, namespacedName, resource); err != nil {
		if k8serr.IsNotFound(err) {
			hwlog.RunLog.Debugf("unable to fetch %s<%s>, err: %s",
				reflect.TypeOf(resource).Name(), namespacedName, err)
		} else {
			hwlog.RunLog.Errorf("unable to fetch %s<%s>, err: %s",
				reflect.TypeOf(resource).Name(), namespacedName, err)
		}
		return client.IgnoreNotFound(err)
	}
	return nil
}

func (r *JobReconciler) UpdateResourceStatus(resource client.Object) error {
	return r.Status().Update(context.Background(), resource)
}

// FetchChildrenForResource fetch child resources via label selector
func (r *JobReconciler) FetchChildrenForResource(resource metav1.Object,
	children client.ObjectList, identification string) error {
	labels := r.GenLabels(resource, identification)
	// Create matching selector.
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: labels,
	})
	if err != nil {
		return fmt.Errorf("couldn't convert Job selector: %v", err)
	}

	return r.List(context.TODO(), children, client.MatchingLabelsSelector{Selector: selector},
		client.InNamespace(resource.GetNamespace()))
}

func (r *JobReconciler) CreateChildForResource(resource client.Object,
	child client.Object, identification string) error {
	selectorLabels := r.GenLabels(resource, identification)
	childLabels := child.GetLabels()
	if childLabels == nil {
		childLabels = map[string]string{}
	}
	// 合并labels
	for key, value := range selectorLabels {
		childLabels[key] = value
	}
	child.SetLabels(childLabels)
	childOwnerReferences := child.GetOwnerReferences()
	childOwnerReferences = append(childOwnerReferences, *r.GenOwnerReference(resource))
	child.SetOwnerReferences(childOwnerReferences)
	return r.Create(context.TODO(), child)
}

func (r *JobReconciler) DeleteResource(resource client.Object) error {
	return r.Delete(context.TODO(), resource)
}

func (r *JobReconciler) GenLabels(ownerResource metav1.Object, identification string) map[string]string {
	labels := make(map[string]string)
	namespace := ownerResource.GetNamespace()
	name := ownerResource.GetName()
	labelValue := fmt.Sprintf("%s-%s-%s", namespace, name, identification)
	labels[common.OwnerLabel] = labelValue
	return labels
}

func (r *JobReconciler) SetCommonPodLabel(podTemplate *corev1.PodTemplateSpec, job client.Object) {
	podTemplate.Labels[common.PodVersionLabel] = strconv.FormatInt(int64(common.DefaultPodVersion), common.Decimal)
	if version, ok := r.Versions[job.GetUID()]; ok {
		podTemplate.Labels[common.PodVersionLabel] = strconv.FormatInt(int64(version), common.Decimal)
	}
}

func (r *JobReconciler) GenOwnerReference(obj client.Object) *metav1.OwnerReference {
	boolPtr := func(b bool) *bool { return &b }
	return &metav1.OwnerReference{
		APIVersion:         r.ConfigInfo.GetAPIGroupVersion().String(),
		Kind:               r.ConfigInfo.GetAPIGroupVersionKind().Kind,
		Name:               obj.GetName(),
		UID:                obj.GetUID(),
		BlockOwnerDeletion: boolPtr(true),
		Controller:         boolPtr(true),
	}
}

func (r *JobReconciler) GetSvcFromApiserver(svcName, svcNamespace string) (*corev1.Service, error) {
	return r.KubeClientSet.CoreV1().Services(svcNamespace).Get(context.Background(), svcName, metav1.GetOptions{})
}

func (r *JobReconciler) CreateService(namespace string, svc *corev1.Service) (*corev1.Service, error) {
	return r.KubeClientSet.CoreV1().Services(namespace).Create(context.TODO(), svc, metav1.CreateOptions{})
}

func (r *JobReconciler) SetRayEnv(pod *corev1.Pod, cmIdentify string) error {
	cmName := common.GenConfigInfoConfigMapName(cmIdentify)
	cm, err := r.GetConfigMapWithRetry(pod.Namespace, cmName)
	if err != nil {
		hwlog.RunLog.Warnf("Error obtaining ConfigMap %s/%s: %v", pod.Namespace, cmName, err)
		return err
	}
	if cm == nil {
		hwlog.RunLog.Warnf("ConfigMap %s/%s not found", pod.Namespace, cmName)
		return nil
	}
	if needUpdate := common.SetRayEnvToCM(pod, cm); needUpdate {
		err := r.UpdateConfigMap(cm)
		if err != nil {
			hwlog.RunLog.Warnf("update pod %s/%s env error: %s",
				pod.Namespace, pod.Name, err)
		} else {
			hwlog.RunLog.Infof("update pod %s/%s env success", pod.Namespace, pod.Name)
		}
	}
	return nil
}

func (r *JobReconciler) CreateConfigMap(cm *corev1.ConfigMap) error {
	_, err := r.KubeClientSet.CoreV1().ConfigMaps(cm.Namespace).Create(context.TODO(),
		cm, metav1.CreateOptions{})
	return err
}

func (r *JobReconciler) DeleteConfigMap(cm *corev1.ConfigMap) error {
	err := r.KubeClientSet.CoreV1().ConfigMaps(cm.Namespace).Delete(context.TODO(),
		cm.Name, metav1.DeleteOptions{})
	return err
}

func (r *JobReconciler) UpdateConfigMap(cm *corev1.ConfigMap) error {
	hwlog.RunLog.Infof("cmName: %s, cmNamespace: %s", cm.Name, cm.Namespace)
	// To reduce the cm write operations
	if !r.IsConfigMapChanged(cm) {
		hwlog.RunLog.Infof("configMap not changed,no need update")
		return nil
	}
	_, err := r.KubeClientSet.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Update(context.TODO(),
		cm, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("unable to update ConfigMap:%s", err)
	}
	return nil
}

// IsConfigMapChanged judge the cm wither is same. true is no change.
func (r *JobReconciler) IsConfigMapChanged(cm *corev1.ConfigMap) bool {
	cmData, getErr := r.GetConfigMapWithRetry(cm.Namespace, cm.Name)
	if getErr != nil {
		return true
	}
	if reflect.DeepEqual(cmData, cm) {
		return false
	}

	return true
}

// GetConfigMapWithRetry  Get config map from k8s.
func (r *JobReconciler) GetConfigMapWithRetry(namespace, cmName string) (*corev1.ConfigMap, error) {
	var cm *corev1.ConfigMap
	var err error

	for i := 0; i < common.ConfigMapRetry; i++ {
		// There can be no delay or blocking operations in a session.
		cm, err = r.KubeClientSet.CoreV1().ConfigMaps(namespace).Get(context.TODO(), cmName, metav1.GetOptions{})
		if k8serr.IsNotFound(err) {
			return nil, nil
		}
		if err != nil {
			time.Sleep(common.ConfigMapRetrySleepTime)
			continue
		}
		return cm, nil
	}
	return nil, err
}

func (r *JobReconciler) DeleteJob(job client.Object) error {
	if err := r.Delete(context.Background(), job); err != nil {
		r.Recorder.Eventf(job, corev1.EventTypeWarning,
			common.FailedDeleteJobReason, "Error deleting: %v", err)
		hwlog.RunLog.Errorf("failed to delete job<%s-%s>, err: %s", job.GetNamespace(), job.GetName(), err)
		return err
	}

	r.Recorder.Eventf(job, corev1.EventTypeWarning,
		common.SuccessfulDeleteJobReason, "Deleted job: %v", job.GetName())
	hwlog.RunLog.Infof("job<%s-%s> has been deleted", job.GetNamespace(), job.GetName())
	return nil
}

func (r *JobReconciler) SetRestartPolicy(job runtime.Object, podTemplateSpec *corev1.PodTemplateSpec,
	restartPolicy commonv1.RestartPolicy) {
	// Submit a warning event if the user specifies restart policy for
	// the pod template. We recommend to set it from the replica level.
	if podTemplateSpec.Spec.RestartPolicy != corev1.RestartPolicy("") {
		errMsg := "Restart policy in pod template will be overwritten by restart policy in replica spec"
		hwlog.RunLog.Warnf(errMsg)
		r.Recorder.Event(job, corev1.EventTypeWarning, common.PodTemplateRestartPolicyReason, errMsg)
	}

	// This is necessary since restartPolicyExitCode is not supported in v1.PodTemplateSpec
	if restartPolicy == commonv1.RestartPolicyExitCode {
		podTemplateSpec.Spec.RestartPolicy = corev1.RestartPolicyNever
	} else {
		podTemplateSpec.Spec.RestartPolicy = corev1.RestartPolicy(restartPolicy)
	}
}

func (r *JobReconciler) GetAPIGroupVersionKind() schema.GroupVersionKind {
	hwlog.RunLog.Errorf("Reconciler must implement GetAPIGroupVersionKind()")
	panic("Reconciler must implement GetAPIGroupVersionKind()")
}

func (r *JobReconciler) GetAPIGroupVersion() schema.GroupVersion {
	hwlog.RunLog.Errorf("Reconciler must implement GetAPIGroupVersion()")
	panic("Reconciler must implement GetAPIGroupVersion()")
}

func (r *JobReconciler) GetRunPolicy(_ client.Object) (*commonv1.RunPolicy, error) {
	hwlog.RunLog.Errorf("Reconciler must implement GetRunPolicy()")
	panic("Reconciler must implement GetRunPolicy()")
}

func (r *JobReconciler) GetStatus(_ client.Object) (*commonv1.JobStatus, error) {
	hwlog.RunLog.Errorf("Reconciler must implement GetStatus()")
	panic("Reconciler must implement GetStatus()")
}
