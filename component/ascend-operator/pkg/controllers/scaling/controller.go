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

/*
Package scaling is using for scale AscendJobs.
*/
package scaling

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"
	volschedulinglisters "volcano.sh/apis/pkg/client/listers/scheduling/v1beta1"

	"ascend-common/common-utils/hwlog"
	apiv1 "ascend-operator/pkg/api/v1"
)

// Controller is the controller for scaling.
type Controller struct {
	client   kubernetes.Interface
	pgLister volschedulinglisters.PodGroupLister
}

// New creates a new instance of the Controller with the provided Kubernetes client and PodGroup lister.
func New(client kubernetes.Interface, pgLister volschedulinglisters.PodGroupLister) *Controller {
	return &Controller{
		client:   client,
		pgLister: pgLister,
	}
}

// ValidJob checks if the given job has the required labels for scaling and grouping.
func (c *Controller) ValidJob(job *apiv1.Job) error {
	_, ok := job.Labels[scalingRuleKey]
	if !ok {
		return nil
	}

	_, ok = job.Labels[groupNameKey]
	if !ok {
		return fmt.Errorf("job %s has no group name", job.Name)
	}

	_, ok = job.Labels[jobGroupNameKey]
	if !ok {
		return fmt.Errorf("job %s has no jobID", job.Name)
	}
	return nil
}

func (c *Controller) getRuleRefPodGroups(namespace, rule, jobID string) (map[string]int, error) {
	selector, err := v1.LabelSelectorAsSelector(&v1.LabelSelector{
		MatchLabels: map[string]string{
			scalingRuleKey:  rule,
			jobGroupNameKey: jobID,
		},
	})
	if err != nil {
		return nil, err
	}

	podGroups, err := c.pgLister.PodGroups(namespace).List(selector)
	if err != nil {
		return nil, err
	}
	hwlog.RunLog.Debugf("get pod groups: %d", len(podGroups))

	groupMap := make(map[string]int)
	for _, pg := range podGroups {
		groupName, ok := pg.Labels[groupNameKey]
		if !ok {
			return nil, fmt.Errorf("pod group %s has no group name", pg.Name)
		}

		if pg.Status.Phase == v1beta1.PodGroupRunning {
			hwlog.RunLog.Infof("group %s podGroup %s is running", groupName, pg.Name)
			groupMap[groupName]++
		}
	}
	return groupMap, nil
}

func (c *Controller) getScalingRule(namespace, name string) ([]map[string]*groupInfo, error) {
	cm, err := c.client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	ruleStr, ok := cm.Data[configmapRuleKey]
	if !ok {
		return nil, fmt.Errorf("configmap %s has no %s", name, configmapRuleKey)
	}

	rule := Rule{}

	if err = json.Unmarshal([]byte(ruleStr), &rule); err != nil {
		return nil, fmt.Errorf("unmarshal rule.json failed: %v", err)
	}
	hwlog.RunLog.Debugf("get scaling rule: %v", rule)
	return convertRule(&rule)
}

// CanCreatePod checks if a new pod can be created for the given job.
func (c *Controller) CanCreatePod(job *apiv1.Job) bool {
	if job == nil {
		return false
	}

	ruleName, ok := job.Labels[scalingRuleKey]
	if !ok {
		return true
	}

	currentGroupName := job.Labels[groupNameKey]
	jobID := job.Labels[jobGroupNameKey]

	rule, err := c.getScalingRule(job.Namespace, ruleName)
	if err != nil {
		hwlog.RunLog.Errorf("get scaling rule failed: %v", err)
		return false
	}
	hwlog.RunLog.Debugf("get scaling rule: %v", rule)
	currentGroup, err := c.getRuleRefPodGroups(job.Namespace, ruleName, jobID)
	if err != nil {
		hwlog.RunLog.Errorf("get rule ref pod groups failed: %v", err)
		return false
	}
	hwlog.RunLog.Debugf("get current group: %v", currentGroup)
	destIndex := getDestOfCurrentState(rule, currentGroup)
	if destIndex == invalidIndex {
		hwlog.RunLog.Infof("current state is arrived at the destination")
		return false
	}
	hwlog.RunLog.Infof("current group num: %d, dest: %d", currentGroup[currentGroupName],
		rule[destIndex][currentGroupName].groupNum)
	return rule[destIndex][currentGroupName].groupNum > currentGroup[currentGroupName]
}
