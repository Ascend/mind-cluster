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

// Package slownodejob provides a sync map for storing slow node and the JobContext
package slownodejob

import (
	"errors"
	"fmt"
	"sync"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/core/model/enum"
	"ascend-faultdiag-online/pkg/model/slownode"
	"ascend-faultdiag-online/pkg/utils/grpc"
	"ascend-faultdiag-online/pkg/utils/k8s"
)

type cluster struct {
	isStartedHeavyProfiling bool
	mu                      sync.Mutex
	// AlgoRes is the slow node algo result
	AlgoRes []*slownode.ClusterAlgoResult
	// TrainingJobStatus is the status of the training job
	TrainingJobStatus string
	// ReportedNodeIps stores all the node ip reported the data profiling
	ReportedNodeIps sync.Map
	// MergeParallelGroupInfoSignal is used to merge the parallel group info
	MergeParallelGroupInfoSignal chan struct{}
	// IsDegradation whether cluster is in degradation state or not
	IsDegradation bool
	// NodeReportSignal node report signal
	NodeReportSignal chan struct{}
	// rescheduleCount the reschedule count of training job
	rescheduleCount int
	// whether need report slow node rank ids or not
	needReport bool
	// slowRankIds the slow rank ids to be reported
	slowRankIds []int
}

// AddAlgoRecord add the slow node algo result in JobContext
func (c *cluster) AddAlgoRecord(result *slownode.ClusterAlgoResult) {
	if c == nil || result == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.AlgoRes) > recordsCapacity {
		start := len(c.AlgoRes) - recordsCapacity + 1
		c.AlgoRes = c.AlgoRes[start:]
	}
	c.AlgoRes = append(c.AlgoRes, result)
}

// AddRecords add the slow node algo result in JobContext
func (c *cluster) UpdateTrainingJobStatus(status string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.TrainingJobStatus = status
}

// AddReportedNodeIp adds the node IP to the parallel group info result
func (c *cluster) AddReportedNodeIp(nodeIp string) {
	if c == nil {
		return
	}
	c.ReportedNodeIps.Store(nodeIp, struct{}{})
}

// GetReportedNodeIps returns the parallel group info result
func (c *cluster) GetReportedNodeIps() []string {
	if c == nil {
		return nil
	}
	var nodeIps = make([]string, 0)
	c.ReportedNodeIps.Range(func(key, _ any) bool {
		nodeIp, ok := key.(string)
		if !ok {
			return true
		}
		nodeIps = append(nodeIps, nodeIp)
		return true
	})
	return nodeIps
}

// TriggerMerge send a signal to merge the parallel group info
func (c *cluster) TriggerMerge() {
	if c == nil {
		return
	}
	select {
	case c.MergeParallelGroupInfoSignal <- struct{}{}:
	default:
		hwlog.RunLog.Warnf("merge parallel group info signal is already sent")
	}
}

// GetRescheduleCount get the reschedule count of the training job
func (c *cluster) GetRescheduleCount() int {
	if c == nil {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.rescheduleCount
}

// SetRescheduleCount set the reschedule count of the training job
func (c *cluster) SetRescheduleCount(count int) {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.rescheduleCount = count
	defer c.mu.Unlock()
}

// NeedReport returns whether need report slow node rank ids or not
func (c *cluster) NeedReport() bool {
	if c == nil {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.needReport
}

// SetNeedReport sets whether need report slow node rank ids or not
func (c *cluster) SetNeedReport(need bool) {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.needReport = need
	defer c.mu.Unlock()
}

// GetSlowRankIds returns the slow rank ids to be reported
func (c *cluster) GetSlowRankIds() []int {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.slowRankIds
}

// SetSlowRankIds sets the slow rank ids to be reported
func (c *cluster) SetSlowRankIds(rankIds []int) {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.slowRankIds = rankIds
	c.mu.Unlock()
}

// ClearSlowRankIds clears the slow rank ids
func (c *cluster) ClearSlowRankIds() {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.slowRankIds = []int{}
	c.mu.Unlock()
}

type node struct {
	// RealRankId realRankIds parsed in data parse
	RealRankIds []string
}

// JobContext is a mixed struct with job and cluster/node info
type JobContext struct {
	// mutex
	mu sync.Mutex

	// Job is the object data from configmap
	Job       *slownode.Job
	step      Step
	isRunning bool

	// Deployment is the deployment type, cluster or node
	Deployment enum.DeployMode
	// StopChan is a stop signal for all goroutine
	StopChan chan struct{}
	// AllCMNames is a sync map including all the cm names have been created
	AllCMNames sync.Map
	cluster
	node
}

// NewSlowNode returns a new SlowNode object
func NewJobContext(job *slownode.Job, deployment enum.DeployMode) *JobContext {
	if job == nil {
		hwlog.RunLog.Error("[FD-OL SLOWNODE]create slow node JobContext failed: job is nil")
		return nil
	}
	ctx := &JobContext{
		Job:        job,
		Deployment: deployment,
		isRunning:  false,
		step:       InitialStep,
		cluster: cluster{
			AlgoRes:     make([]*slownode.ClusterAlgoResult, 0),
			slowRankIds: make([]int, 0),
		},
	}
	return ctx
}

// Start the job
func (ctx *JobContext) Start() {
	if ctx == nil {
		return
	}
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.StopChan = make(chan struct{}, channelCapacity)
	ctx.isRunning = true
	ctx.step = InitialStep
	ctx.ReportedNodeIps = sync.Map{}
	ctx.MergeParallelGroupInfoSignal = make(chan struct{}, channelCapacity)
	ctx.NodeReportSignal = make(chan struct{}, channelCapacity)
	ctx.IsDegradation = false
	ctx.isStartedHeavyProfiling = false
	ctx.cluster.ClearSlowRankIds()
	ctx.cluster.SetNeedReport(false)
}

// Stop the job
func (ctx *JobContext) Stop() {
	if ctx == nil {
		return
	}
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if ctx.isRunning {
		close(ctx.StopChan)
		ctx.isRunning = false
		ctx.cluster.ReportedNodeIps = sync.Map{}
		ctx.MergeParallelGroupInfoSignal = make(chan struct{}, channelCapacity)
		ctx.NodeReportSignal = make(chan struct{}, channelCapacity)
		ctx.IsDegradation = false
		ctx.isStartedHeavyProfiling = false
		ctx.cluster.ClearSlowRankIds()
		ctx.cluster.SetNeedReport(false)
	}
}

// IsRunning check if the job is running
func (ctx *JobContext) IsRunning() bool {
	if ctx == nil {
		return false
	}
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	return ctx.isRunning
}

// AddStep add the step
func (ctx *JobContext) AddStep() {
	if ctx == nil {
		return
	}
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.step++
}

// Step get the step
func (ctx *JobContext) Step() Step {
	if ctx == nil {
		return InitialStep
	}
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	return ctx.step
}

// IsStartedHeavyProfiling returns a bool whether the heavy profiling starts or not
func (ctx *JobContext) IsStartedHeavyProfiling() bool {
	if ctx == nil {
		return false
	}
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	return ctx.isStartedHeavyProfiling
}

func (ctx *JobContext) LogPrefix() string {
	if ctx == nil || ctx.Job == nil {
		return "[FD-OL SLOWNODE]job(nil)"
	}
	return fmt.Sprintf("[FD-OL SLOWNODE]job(name=%s, namespace=%v, jobId=%s)",
		ctx.Job.JobName, ctx.Job.Namespace, ctx.Job.JobId)
}

// StartAllProfiling start all the profiling
func (ctx *JobContext) StartAllProfiling() error {
	if ctx == nil || ctx.Job == nil {
		return errors.New("ctx is nil or ctx.Job is nil")
	}
	grpcClient, err := grpc.GetClient()
	if err != nil {
		return err
	}
	if err := grpcClient.StartAllProfiling(ctx.Job.JobName, ctx.Job.Namespace); err != nil {
		return err
	}
	hwlog.RunLog.Infof("%s started all profiling successfully", ctx.LogPrefix())
	// step from 0 to 1
	ctx.AddStep()
	return nil
}

// StopAllProfiling stop all the profiling
func (ctx *JobContext) StopAllProfiling() {
	if ctx == nil || ctx.Job == nil {
		return
	}
	grpcClient, err := grpc.GetClient()
	if err != nil {
		hwlog.RunLog.Errorf("%s got grpc client failed: %v", ctx.LogPrefix(), err)
		return
	}
	if err := grpcClient.StopAllProfiling(ctx.Job.JobName, ctx.Job.Namespace); err != nil {
		hwlog.RunLog.Errorf("%s stopped all profiling failed: %s", ctx.LogPrefix(), err.Error())
		return
	}
	hwlog.RunLog.Infof("%s stopped all profiling successfully", ctx.LogPrefix())
}

// StartHeavyProfiling start the heavy profiling
func (ctx *JobContext) StartHeavyProfiling() {
	if ctx == nil || ctx.Job == nil {
		return
	}
	grpcClient, err := grpc.GetClient()
	if err != nil {
		hwlog.RunLog.Errorf("%s got grpc client failed: %v", ctx.LogPrefix(), err)
		return
	}
	if err := grpcClient.StartHeavyProfiling(ctx.Job.JobName, ctx.Job.Namespace); err != nil {
		hwlog.RunLog.Errorf("%s) started heavy profiling failed: %s", ctx.LogPrefix(), err.Error())
		return
	}
	hwlog.RunLog.Infof("%s started heavy profiling successfully", ctx.LogPrefix())
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.isStartedHeavyProfiling = true
}

// StopHeavyProfiling stop the heavy profiling
func (ctx *JobContext) StopHeavyProfiling() {
	if ctx == nil || ctx.Job == nil {
		return
	}
	grpcClient, err := grpc.GetClient()
	if err != nil {
		hwlog.RunLog.Errorf("%s got grpc client failed: %v", ctx.LogPrefix(), err)
		return
	}
	if err := grpcClient.StopHeavyProfiling(ctx.Job.JobName, ctx.Job.Namespace); err != nil {
		hwlog.RunLog.Errorf("%s stopped heavy profiling failed: %s", ctx.LogPrefix(), err.Error())
		return
	}
	hwlog.RunLog.Infof("%s stopped heavy profiling successfully", ctx.LogPrefix())
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.isStartedHeavyProfiling = false
	ctx.AlgoRes = make([]*slownode.ClusterAlgoResult, 0)
}

// Update the slow node to the current job
func (ctx *JobContext) Update(job *slownode.Job) {
	if ctx == nil || ctx.Job == nil || job == nil {
		return
	}
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.Job.SlowNode = job.SlowNode
	ctx.Job.Servers = job.Servers
	ctx.Job.RankIds = job.RankIds
}

// RemoveAllCM remove all the config map stored in ctx
func (ctx *JobContext) RemoveAllCM() {
	if ctx == nil || ctx.Job == nil {
		return
	}
	k8sClient, err := k8s.GetClient()
	if err != nil {
		hwlog.RunLog.Errorf("%s got k8s client failed: %v", ctx.LogPrefix(), err)
		return
	}
	ctx.AllCMNames.Range(func(key, value any) bool {
		cmName, ok := key.(string)
		if !ok {
			hwlog.RunLog.Errorf(
				"%s deleted cm: %s failed: key is not a string type", ctx.LogPrefix(), cmName)
		}
		if err := k8sClient.DeleteConfigMap(cmName, ctx.Job.Namespace); err != nil {
			hwlog.RunLog.Errorf("%s deleted cm: %s failed: %s", ctx.LogPrefix(), cmName, err)
		} else {
			hwlog.RunLog.Infof("%s deleted cm: %s successfully", ctx.LogPrefix(), cmName)
		}
		return true
	})
}

// AllNodesReported checks if all the nodes have reported the data profiling
func (ctx *JobContext) AllNodesReported() bool {
	if ctx == nil || ctx.Job == nil {
		return false
	}
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	for _, server := range ctx.Job.Servers {
		if _, ok := ctx.ReportedNodeIps.Load(server.Ip); !ok {
			return false
		}
	}
	return true
}
