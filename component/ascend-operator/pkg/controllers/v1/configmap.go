/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile AscendJob.
*/

package v1

import (
	"context"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *ASJobReconciler) getConfigmapFromApiserver(namespace, name string) (*v1.ConfigMap, error) {
	return r.KubeClientSet.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func (r *ASJobReconciler) getVcRescheduleCM() (*v1.ConfigMap, error) {
	return r.getConfigmapFromApiserver(vcNamespace, vcRescheduleCMName)
}
