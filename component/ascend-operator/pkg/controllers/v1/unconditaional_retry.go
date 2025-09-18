/*
Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
*/

/*
Package controllers is using for reconcile Job.
*/

package v1

import (
	"encoding/json"
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/types"

	"ascend-operator/pkg/api/v1"
)

// RemainRetryTimes data in volcano reschedule configmap
type RemainRetryTimes struct {
	UUID  types.UID
	Times int
}

func (r *ASJobReconciler) isUnconditionalRetryJob(job *v1.Job) bool {
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

func (r *ASJobReconciler) getJobRemainRetryTimes(job *v1.Job) (int, error) {
	vcReCM, err := r.getVcRescheduleCM()
	if err != nil {
		return -1, err
	}
	rrt, ok := vcReCM.Data[cmJobRemainRetryTimes]
	if !ok {
		return -1, fmt.Errorf("volcaco reschedule confimap has no remain-retry-times key")
	}
	rTimes, err := unmarshalRemainRetryTimes(rrt)
	if err != nil {
		return -1, err
	}
	uid := job.GetNamespace() + "/" + job.GetName() + "-" + string(job.GetUID())
	if rt, ok := rTimes[types.UID(uid)]; ok {
		return rt.Times, nil
	}

	return -1, fmt.Errorf("remain times has no job<%s> data", job.GetUID())
}

func unmarshalRemainRetryTimes(data string) (map[types.UID]*RemainRetryTimes, error) {
	rTimes := make(map[types.UID]*RemainRetryTimes)
	if unmarshalErr := json.Unmarshal([]byte(data), &rTimes); unmarshalErr != nil {
		return nil, fmt.Errorf("remain times convert from CM error %s", unmarshalErr)
	}
	return rTimes, nil
}
