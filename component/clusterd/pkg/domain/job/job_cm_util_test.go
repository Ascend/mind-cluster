// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

//go:build !race

// Package job a series of job test function
package job

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/interface/kube"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

func TestInitCM(t *testing.T) {
	convey.Convey("test initCM", t, func() {
		jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
		convey.Convey("when CreateOrUpdateConfigMap failed, return false", func() {
			mockCreateOrUpdateConfigMap := gomonkey.ApplyFunc(kube.CreateOrUpdateConfigMap,
				func(cmName, nameSpace string, data, label map[string]string) error {
					return errors.New("test error")
				})
			defer mockCreateOrUpdateConfigMap.Reset()
			result := initCM(jobInfo)
			convey.So(result, convey.ShouldEqual, false)
		})
		convey.Convey("when CreateOrUpdateConfigMap success, return true", func() {
			mockCreateOrUpdateConfigMap := gomonkey.ApplyFunc(kube.CreateOrUpdateConfigMap,
				func(cmName, nameSpace string, data, label map[string]string) error {
					return nil
				})
			defer mockCreateOrUpdateConfigMap.Reset()
			result := initCM(jobInfo)
			convey.So(result, convey.ShouldEqual, true)
		})
	})
}

func TestUpdateCM(t *testing.T) {
	convey.Convey("test updateCM", t, func() {
		jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
		convey.Convey("when framework is empty, index is 0, UpdateOrCreateConfigMap failed."+
			" return false", func() {
			mockUpdateOrCreateConfigMap := gomonkey.ApplyFunc(kube.UpdateOrCreateConfigMap,
				func(cmName, nameSpace string, data, label map[string]string) error {
					return errors.New("test error")
				})
			defer mockUpdateOrCreateConfigMap.Reset()
			result := updateCM(jobInfo, 0, "")
			convey.So(result, convey.ShouldEqual, false)
		})
		convey.Convey("when framework is pytorch, index is 1, UpdateOrCreateConfigMap success."+
			" return true", func() {
			mockUpdateOrCreateConfigMap := gomonkey.ApplyFunc(kube.UpdateOrCreateConfigMap,
				func(cmName, nameSpace string, data, label map[string]string) error {
					return nil
				})
			defer mockUpdateOrCreateConfigMap.Reset()
			jobInfo.Framework = ptFramework
			result := updateCM(jobInfo, 1, "")
			convey.So(result, convey.ShouldEqual, true)
		})
	})
}

func TestPreDeleteCM(t *testing.T) {
	convey.Convey("test preDeleteCM", t, func() {
		jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
		convey.Convey("when framework is empty, totalCmNum is 0, CreateOrUpdateConfigMap success."+
			" return true", func() {
			mockCreateOrUpdateConfigMap := gomonkey.ApplyFunc(kube.CreateOrUpdateConfigMap,
				func(cmName, nameSpace string, data, label map[string]string) error {
					return nil
				})
			defer mockCreateOrUpdateConfigMap.Reset()
			hccls := make([]string, 0)
			result := preDeleteCM(jobInfo, hccls)
			convey.So(result, convey.ShouldEqual, true)
		})
		convey.Convey("when framework is pytorch, totalCmNum is 1, CreateOrUpdateConfigMap failed."+
			" return false", func() {
			mockCreateOrUpdateConfigMap := gomonkey.ApplyFunc(kube.CreateOrUpdateConfigMap,
				func(cmName, nameSpace string, data, label map[string]string) error {
					return errors.New("test error")
				})
			defer mockCreateOrUpdateConfigMap.Reset()
			jobInfo.TotalCmNum = 1
			jobInfo.Framework = ptFramework
			hccls := make([]string, 1)
			hccls[0] = "{}"
			result := preDeleteCM(jobInfo, hccls)
			convey.So(result, convey.ShouldEqual, false)
		})
		convey.Convey("when framework is pytorch, totalCmNum is 1 but hccls is nil, "+
			"CreateOrUpdateConfigMap success. return true", func() {
			mockCreateOrUpdateConfigMap := gomonkey.ApplyFunc(kube.CreateOrUpdateConfigMap,
				func(cmName, nameSpace string, data, label map[string]string) error {
					return nil
				})
			defer mockCreateOrUpdateConfigMap.Reset()
			jobInfo.TotalCmNum = 1
			jobInfo.Framework = ptFramework
			hccls := make([]string, 0)
			result := preDeleteCM(jobInfo, hccls)
			convey.So(result, convey.ShouldEqual, true)
		})
	})
}

func TestDeleteCm(t *testing.T) {
	convey.Convey("test deleteCm", t, func() {
		jobInfo := getDemoJob(jobName1, jobNameSpace, jobUid1)
		convey.Convey("when totalCmNum is 0, DeleteConfigMap success. return true", func() {
			mockDeleteConfigMap := gomonkey.ApplyFunc(kube.DeleteConfigMap,
				func(cmName, nameSpace string) error {
					return nil
				})
			defer mockDeleteConfigMap.Reset()
			result := deleteCm(jobInfo)
			convey.So(result, convey.ShouldEqual, true)
		})
		convey.Convey("when totalCmNum is 1, DeleteConfigMap failed. return true", func() {
			mockDeleteConfigMap := gomonkey.ApplyFunc(kube.DeleteConfigMap,
				func(cmName, nameSpace string) error {
					return errors.New("test error")
				})
			defer mockDeleteConfigMap.Reset()
			jobInfo.TotalCmNum = 1
			result := deleteCm(jobInfo)
			convey.So(result, convey.ShouldEqual, false)
		})
		convey.Convey("when totalCmNum is 1, DeleteConfigMap is notFound failed. return true", func() {
			mockDeleteConfigMap := gomonkey.ApplyFunc(kube.DeleteConfigMap,
				func(cmName, nameSpace string) error {
					var err = &apierrors.StatusError{ErrStatus: metav1.Status{Reason: metav1.StatusReasonNotFound}}
					return err
				})
			defer mockDeleteConfigMap.Reset()
			jobInfo.TotalCmNum = 1
			result := deleteCm(jobInfo)
			convey.So(result, convey.ShouldEqual, true)
		})
	})
}

func TestGetDefaultLabel(t *testing.T) {
	convey.Convey("test getDefaultLabel. return map have lable", t, func() {
		labels := getDefaultLabel()
		convey.So(labels[configmapLabel], convey.ShouldEqual, "true")
	})
}

func TestRefreshFaultJobInfoCmData(t *testing.T) {
	convey.Convey("test refreshFaultJobInfoCmData", t, func() {
		// Initialize jobSummaryMap for testing
		originalJobSummaryMap := jobSummaryMap
		jobSummaryMap = sync.Map{}
		defer func() { jobSummaryMap = originalJobSummaryMap }()

		convey.Convey("when jobSummaryMap is empty, all data should be filtered out", func() {
			inputData := map[string]string{
				"job-uid-1": "job-data-1",
				"job-uid-2": "job-data-2",
			}
			result := refreshFaultJobInfoCmData(inputData)
			convey.So(len(result), convey.ShouldEqual, 0)
		})

		convey.Convey("when jobSummaryMap contains some jobUids, only matching data should be retained", func() {
			// Populate jobSummaryMap with some jobUids
			jobSummaryMap.Store("job-uid-1", constant.JobInfo{})
			jobSummaryMap.Store("job-uid-3", constant.JobInfo{})
			inputData := map[string]string{
				"job-uid-1": "job-data-1",
				"job-uid-2": "job-data-2",
				"job-uid-3": "job-data-3",
				"job-uid-4": "job-data-4",
			}
			expectedResult := map[string]string{
				"job-uid-1": "job-data-1",
				"job-uid-3": "job-data-3",
			}
			result := refreshFaultJobInfoCmData(inputData)
			convey.So(result, convey.ShouldResemble, expectedResult)
		})

		convey.Convey("when jobSummaryMap contains all jobUids, all data should be retained", func() {
			// Populate jobSummaryMap with all jobUids
			jobSummaryMap.Store("job-uid-1", constant.JobInfo{})
			jobSummaryMap.Store("job-uid-2", constant.JobInfo{})
			inputData := map[string]string{
				"job-uid-1": "job-data-1",
				"job-uid-2": "job-data-2",
			}
			result := refreshFaultJobInfoCmData(inputData)
			convey.So(result, convey.ShouldResemble, inputData)
		})
	})
}
