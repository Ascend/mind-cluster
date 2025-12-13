// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package profiling a series of service function for profiling
package profiling

import (
	"context"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc/metadata"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/config"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/domain/profile"
	"clusterd/pkg/interface/grpc/profiling"
	"clusterd/pkg/interface/kube"
)

func init() {
	hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background())
}

const (
	on               = "on"
	jobName          = "jobName"
	JobNsName        = "nsName/jobName"
	invalidJobNsName = "nsName-jobName"
)

var defaultSwitch = &profiling.ProfilingSwitch{
	CommunicationOperator: on,
	Step:                  on,
	SaveCheckpoint:        on,
	FP:                    on,
	DataLoader:            on,
}

var publisher = config.NewConfigPublisher[*profiling.DataStatusRes](
	jobName, context.Background(), "", nil)

func TestModifyTrainingDataTraceSwitch(t *testing.T) {
	testInvalidJobNsNameFormat(t)
	testConfigMapNotExistAndCreateSuccess(t)
	testConfigMapNotExistAndCreateFail(t)
	testConfigMapExistAndUpdateSuccess(t)
	testConfigMapExistAndUpdateFail(t)
	testGetConfigMapUnexpectedError(t)
}

func testInvalidJobNsNameFormat(t *testing.T) {
	convey.Convey("when jobNsName format is invalid", t, func() {
		ps := NewSwitchManager(context.Background())
		req := &profiling.DataTypeReq{
			JobNsName: invalidJobNsName,
		}
		_, err := ps.ModifyTrainingDataTraceSwitch(context.Background(), req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func testConfigMapNotExistAndCreateSuccess(t *testing.T) {
	convey.Convey("when configmap does not exist and creation succeeds, but cm file not mount", t, func() {
		ps := NewSwitchManager(context.Background())
		req := &profiling.DataTypeReq{
			JobNsName:       "namespace/jobname",
			ProfilingSwitch: defaultSwitch,
		}

		patches := gomonkey.NewPatches()
		defer patches.Reset()

		setupConfigMapNotFoundMock(patches)
		setupCreateDataTraceCmSuccessMock(patches)
		_, err := ps.ModifyTrainingDataTraceSwitch(context.Background(), req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func testConfigMapNotExistAndCreateFail(t *testing.T) {
	convey.Convey("when configmap does not exist and creation fails", t, func() {
		ps := NewSwitchManager(context.Background())
		req := &profiling.DataTypeReq{
			JobNsName:       "namespace/jobname",
			ProfilingSwitch: defaultSwitch,
		}

		patches := gomonkey.NewPatches()
		defer patches.Reset()

		setupConfigMapNotFoundMock(patches)
		setupCreateDataTraceCmFailMock(patches)

		_, err := ps.ModifyTrainingDataTraceSwitch(context.Background(), req)

		convey.So(err, convey.ShouldNotBeNil)
	})
}

func testConfigMapExistAndUpdateSuccess(t *testing.T) {
	convey.Convey("when configmap exists and update succeeds, but cm file not mount", t, func() {
		ps := NewSwitchManager(context.Background())
		req := &profiling.DataTypeReq{
			JobNsName:       JobNsName,
			ProfilingSwitch: defaultSwitch,
		}

		patches := gomonkey.NewPatches()
		defer patches.Reset()

		setupConfigMapExistMock(patches)
		setupUpdateDataTraceCmSuccessMock(patches)
		setupGetJobByNameSpaceAndName(patches)
		setupPublishMock(patches, ps)

		_, err := ps.ModifyTrainingDataTraceSwitch(context.Background(), req)

		convey.So(err, convey.ShouldNotBeNil)
	})
}

func setupGetJobByNameSpaceAndName(patches *gomonkey.Patches) {
	patches.ApplyFunc(job.GetJobByNameSpaceAndName, func(name, nameSpace string) constant.JobInfo {
		return constant.JobInfo{Key: jobName}
	})
}

func testConfigMapExistAndUpdateFail(t *testing.T) {
	convey.Convey("when configmap exists and update fails", t, func() {
		ps := NewSwitchManager(context.Background())
		req := &profiling.DataTypeReq{
			JobNsName:       JobNsName,
			ProfilingSwitch: defaultSwitch,
		}

		patches := gomonkey.NewPatches()
		defer patches.Reset()

		setupConfigMapExistMock(patches)
		setupUpdateDataTraceCmFailMock(patches)

		_, err := ps.ModifyTrainingDataTraceSwitch(context.Background(), req)

		convey.So(err, convey.ShouldNotBeNil)
	})
}

func testGetConfigMapUnexpectedError(t *testing.T) {
	convey.Convey("when getting configmap returns unexpected error", t, func() {
		ps := NewSwitchManager(context.Background())
		req := &profiling.DataTypeReq{
			JobNsName:       JobNsName,
			ProfilingSwitch: defaultSwitch,
		}

		patches := gomonkey.NewPatches()
		defer patches.Reset()

		setupConfigMapUnexpectedErrorMock(patches)

		_, err := ps.ModifyTrainingDataTraceSwitch(context.Background(), req)

		convey.So(err, convey.ShouldNotBeNil)
	})
}

// Helper functions for setting up mocks
func setupConfigMapNotFoundMock(patches *gomonkey.Patches) {
	patches.ApplyFunc(kube.GetConfigMap, func(name, namespace string) (*corev1.ConfigMap, error) {
		return nil, apierrors.NewNotFound(corev1.Resource("configmaps"), name)
	})
}

func setupConfigMapExistMock(patches *gomonkey.Patches) {
	patches.ApplyFunc(kube.GetConfigMap, func(name, namespace string) (*corev1.ConfigMap, error) {
		return &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}, nil
	})
}

func setupConfigMapUnexpectedErrorMock(patches *gomonkey.Patches) {
	patches.ApplyFunc(kube.GetConfigMap, func(name, namespace string) (*corev1.ConfigMap, error) {
		return nil, fmt.Errorf("unexpected error")
	})
}

func setupCreateDataTraceCmSuccessMock(patches *gomonkey.Patches) {
	patches.ApplyMethodFunc(
		&profile.DataTraceController{},
		"CreateDataTraceCm",
		func(*profiling.ProfilingSwitch, metav1.OwnerReference) error { return nil },
	)
}

func setupCreateDataTraceCmFailMock(patches *gomonkey.Patches) {
	patches.ApplyMethodFunc(
		&profile.DataTraceController{},
		"CreateDataTraceCm",
		func(*profiling.ProfilingSwitch, metav1.OwnerReference) error { return fmt.Errorf("creation error") },
	)
}

func setupUpdateDataTraceCmSuccessMock(patches *gomonkey.Patches) {
	patches.ApplyMethodFunc(
		&profile.DataTraceController{},
		"UpdateDataTraceCm",
		func(*profiling.ProfilingSwitch, metav1.OwnerReference) error { return nil },
	)
}

func setupUpdateDataTraceCmFailMock(patches *gomonkey.Patches) {
	patches.ApplyMethodFunc(
		&profile.DataTraceController{},
		"UpdateDataTraceCm",
		func(*profiling.ProfilingSwitch, metav1.OwnerReference) error { return fmt.Errorf("update error") },
	)
}

func setupPublishMock(patches *gomonkey.Patches, ps *SwitchManager) {
	patches.ApplyPrivateMethod(
		ps,
		"publish",
		func(string, *profiling.ProfilingSwitch) {},
	)
}

func TestPublish(t *testing.T) {
	convey.Convey("Test publish", t, func() {
		ps := NewSwitchManager(context.Background())
		jobName := jobName
		info := defaultSwitch

		ps.publishers[jobName] = config.NewConfigPublisher[*profiling.DataStatusRes](jobName,
			ps.ctx, constant.ProfilingDataType, nil)
		err := ps.publish(jobName, info)
		convey.So(err, convey.ShouldBeNil)

		delete(ps.publishers, jobName)
		err = ps.publish(jobName, info)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestGetTrainingDataTraceSwitch(t *testing.T) {
	convey.Convey("Test GetTrainingDataTraceSwitch", t, func() {
		ps := NewSwitchManager(context.Background())
		ctx := context.Background()
		convey.Convey("when jobNsName format is invalid", func() {
			req := &profiling.DataStatusReq{JobNsName: "invalid"}
			_, err := ps.GetTrainingDataTraceSwitch(ctx, req)

			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("when configmap not found", func() {
			req := &profiling.DataStatusReq{JobNsName: "ns/job"}
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			patches.ApplyMethodFunc(
				&profile.DataTraceController{},
				"IsDataTraceCmExist",
				func() (*corev1.ConfigMap, error) {
					return nil, fmt.Errorf("not found")
				})

			_, err := ps.GetTrainingDataTraceSwitch(ctx, req)

			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("when success", func() {
			req := &profiling.DataStatusReq{JobNsName: "ns/job"}
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			patches.ApplyMethodFunc(
				&profile.DataTraceController{},
				"IsDataTraceCmExist",
				func() (*corev1.ConfigMap, error) {
					return &corev1.ConfigMap{Data: map[string]string{
						profile.DataTraceCmProfilingSwitchKey: `{"enabled":true}`,
					}}, nil
				})

			res, err := ps.GetTrainingDataTraceSwitch(ctx, req)

			convey.So(err, convey.ShouldBeNil)
			convey.So(res.ProfilingSwitch, convey.ShouldNotBeNil)
		})
	})
}

type MockStream struct{}

func (s *MockStream) Send(res *profiling.DataStatusRes) error { return nil }

func (s *MockStream) SetHeader(md metadata.MD) error { return nil }

func (s *MockStream) SendHeader(md metadata.MD) error { return nil }

func (s *MockStream) SetTrailer(md metadata.MD) {}

func (s *MockStream) Context() context.Context { return context.Background() }

func (s *MockStream) SendMsg(m interface{}) error { return nil }

func (s *MockStream) RecvMsg(m interface{}) error { return nil }

func TestSubscribeDataTraceSwitch(t *testing.T) {
	convey.Convey("Test SubscribeDataTraceSwitch", t, func() {
		ps := NewSwitchManager(context.Background())
		client := &profiling.ProfilingClientInfo{
			JobId: jobName,
			Role:  "worker",
		}
		var mockStream profiling.TrainingDataTrace_SubscribeDataTraceSwitchServer = &MockStream{}

		convey.Convey("when preRegistry fails", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			patches.ApplyPrivateMethod(ps, "getPublisher",
				func(string) (*config.ConfigPublisher[*profiling.DataStatusRes], bool) { return nil, false })
			patches.ApplyPrivateMethod(ps, "preRegistry",
				func(*profiling.ProfilingClientInfo) (common.RespCode, error) {
					return common.JobNotExist, fmt.Errorf("registry error")
				})

			err := ps.SubscribeDataTraceSwitch(client, mockStream)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("when subscription succeeds", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			patches.ApplyPrivateMethod(ps, "getPublisher",
				func(string) (*config.ConfigPublisher[*profiling.DataStatusRes], bool) { return nil, false })
			patches.ApplyPrivateMethod(ps, "preRegistry",
				func(*profiling.ProfilingClientInfo) (common.RespCode, error) {
					return common.OK, nil
				})
			patches.ApplyPrivateMethod(ps, "preemptPublisher",
				func(string) *config.ConfigPublisher[*profiling.DataStatusRes] { return publisher })

			go func() {
				publisher.Stop()
			}()

			err := ps.SubscribeDataTraceSwitch(client, mockStream)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestSwitchManagerMethods(t *testing.T) {
	convey.Convey("Test SwitchManager methods", t, func() {
		ps := NewSwitchManager(context.Background())
		jobId := "test-job"
		clientInfo := &profiling.ProfilingClientInfo{JobId: jobId, Role: "worker"}

		convey.Convey("getPublisher should return publisher correctly", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			// Mock publishers map
			mockPub := &config.ConfigPublisher[*profiling.DataStatusRes]{}
			ps.publishers = map[string]*config.ConfigPublisher[*profiling.DataStatusRes]{jobId: mockPub}

			pub, ok := ps.getPublisher(jobId)
			convey.So(ok, convey.ShouldBeTrue)
			convey.So(pub, convey.ShouldEqual, mockPub)
		})

		convey.Convey("preRegistry should validate job existence", func() {
			patches := gomonkey.NewPatches()
			defer patches.Reset()

			patches.ApplyFunc(job.GetJobCache, func(string) (constant.JobInfo, bool) {
				return constant.JobInfo{}, false
			})
			patches.ApplyFunc(job.GetNamespaceByJobIdAndAppType, func(string, string) (string, error) {
				return "", fmt.Errorf("not found")
			})
			patches.ApplyFunc(hwlog.RunLog.Errorf, func(string, ...interface{}) {})

			code, err := ps.preRegistry(clientInfo)
			convey.So(code, convey.ShouldEqual, common.JobNotExist)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}
