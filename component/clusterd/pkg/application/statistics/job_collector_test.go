package statistics

import (
	"context"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/api/core/v1"

	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/statistics"
	"clusterd/pkg/interface/kube"
)

func TestJobStatisticCollector(t *testing.T) {
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Mock context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Mock configmap loading
	patches.ApplyFunc(kube.GetConfigMap, func(name, namespace string) (*v1.ConfigMap, error) {
		return &v1.ConfigMap{
			Data: map[string]string{
				statistics.JobDataCmKey:   `[{"Name":"test-job"}]`,
				statistics.TotalJobsCmKey: "1",
			},
		}, nil
	})

	// Start collector
	go GlobalJobCollectMgr.JobCollector(ctx)

	// Send test messages
	testCases := []struct {
		msg constant.JobNotifyMsg
	}{
		{constant.JobNotifyMsg{Operator: constant.JobInfoAdd, JobKey: "job1"}},
		{constant.JobNotifyMsg{Operator: constant.JobInfoUpdate, JobKey: "job2"}},
		{constant.JobNotifyMsg{Operator: constant.JobInfoPreDelete, JobKey: "job3"}},
		{constant.JobNotifyMsg{Operator: constant.JobInfoDelete, JobKey: "job4"}},
	}
	for _, tc := range testCases {
		GlobalJobCollectMgr.JobNotifyChan <- tc.msg
		time.Sleep(time.Second)
	}
}
