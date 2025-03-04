package statistics

import (
	"context"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/statistics"
)

const (
	jobNotifyChanLen = 5
)

// JobCollectorMgr used to manage job statistic data.
type JobCollectorMgr struct {
	JobNotifyChan chan constant.JobNotifyMsg
}

var (
	// GlobalJobCollectMgr is a global instance of StatisticInfo used for statistic data.
	GlobalJobCollectMgr *JobCollectorMgr
)

func init() {
	GlobalJobCollectMgr = &JobCollectorMgr{
		JobNotifyChan: make(chan constant.JobNotifyMsg, jobNotifyChanLen),
	}
}

// JobCollector get the updated status of the job and update the statistics
func (j *JobCollectorMgr) JobCollector(ctx context.Context) {

	// load configMap into Cache if configMap is existed
	statistics.JobStcMgrInst.LoadConfigMapToCache(statistics.JobStcNamespace, statistics.JobStcCMName)

	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Info("faultProcessCenter stop work")
			return
		case notifyMsg := <-j.JobNotifyChan:
			switch notifyMsg.Operator {
			case constant.JobInfoUpdate:
				statistics.JobStcMgrInst.UpdateJobStatistic(notifyMsg.JobKey)
			case constant.JobInfoAdd:
				statistics.JobStcMgrInst.AddJobStatistic(notifyMsg.JobKey)
			case constant.JobInfoPreDelete:
				statistics.JobStcMgrInst.PreDeleteJobStatistic(notifyMsg.JobKey)
			case constant.JobInfoDelete:
				statistics.JobStcMgrInst.DeleteJobStatistic(notifyMsg.JobKey)
			default:
				hwlog.RunLog.Warnf("this logic branch is unreachable, "+
					"there must have been some issues with the code."+
					"JobKey: %s, Operator: %s", notifyMsg.JobKey, notifyMsg.Operator)
				return
			}
		}
	}
}
