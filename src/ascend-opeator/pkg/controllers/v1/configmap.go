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
	"k8s.io/apimachinery/pkg/types"
)

func (r *ASJobReconciler) getConfigmap(namespace, name string) (*v1.ConfigMap, error) {
	cm := v1.ConfigMap{}
	if err := r.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: name}, &cm); err != nil {
		return nil, err
	}
	return &cm, nil
}

func (r *ASJobReconciler) getVcRescheduleCM() (*v1.ConfigMap, error) {
	return r.getConfigmap(vcNamespace, vcRescheduleCMName)
}
