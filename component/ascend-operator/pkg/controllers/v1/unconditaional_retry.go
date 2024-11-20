/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile AscendJob.
*/

package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/types"

	mindxdlv1 "ascend-operator/pkg/api/v1"
)

// RemainRetryTimes data in volcano reschedule configmap
type RemainRetryTimes struct {
	UUID  types.UID
	Times int
}

func (r *ASJobReconciler) isUnconditionalRetryJob(job *mindxdlv1.AscendJob) bool {
	if r.Config.EnableGangScheduling {
		times, ok := job.Labels[unconditionalRetryLabelKey]
		if !ok {
			return false
		}
		if t, err := strconv.Atoi(times); err == nil && t > 0 {
			return true
		}
	}
	return false
}

func (r *ASJobReconciler) getJobRemainRetryTimes(job *mindxdlv1.AscendJob) (int, error) {
	vcReCM, err := r.getVcRescheduleCM()
	if err != nil {
		return -1, err
	}

	rrt, ok := vcReCM.Data[cmJobRemainRetryTimes]
	if !ok {
		return -1, fmt.Errorf("volcaco reschedule confimap has no remain-retry-times key")
	}

	rTimes := make(map[types.UID]*RemainRetryTimes)
	if unmarshalErr := json.Unmarshal([]byte(rrt), &rTimes); unmarshalErr != nil {
		return -1, fmt.Errorf("remain times convert from CM error %s", unmarshalErr)
	}

	uid := job.GetNamespace() + "/" + job.GetName() + "-" + string(job.GetUID())
	if rt, ok := rTimes[types.UID(uid)]; ok {
		return rt.Times, nil
	}

	return -1, fmt.Errorf("remain times has no job<%s> data", job.GetUID())
}
