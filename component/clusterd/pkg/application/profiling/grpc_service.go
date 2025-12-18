// Copyright (c) Huawei Technologies Co., Ltd. 2024-2025. All rights reserved.

// Package profiling a series of service function for profiling
package profiling

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"ascend-common/common-utils/hwlog"
	"clusterd/pkg/application/config"
	"clusterd/pkg/common/constant"
	"clusterd/pkg/common/logs"
	"clusterd/pkg/domain/common"
	"clusterd/pkg/domain/job"
	"clusterd/pkg/domain/podgroup"
	"clusterd/pkg/domain/profile"
	"clusterd/pkg/interface/grpc/profiling"
	"clusterd/pkg/interface/kube"
)

const (
	// PartsOfJobNs parts num of in param
	PartsOfJobNs = 2
)

// SwitchManager represents profiling switch manager
type SwitchManager struct {
	profiling.UnimplementedTrainingDataTraceServer
	publishers map[string]*config.ConfigPublisher[*profiling.DataStatusRes]
	lock       sync.RWMutex
	ctx        context.Context
}

// NewSwitchManager new SwitchManager
func NewSwitchManager(ctx context.Context) *SwitchManager {
	return &SwitchManager{
		publishers: make(map[string]*config.ConfigPublisher[*profiling.DataStatusRes]),
		lock:       sync.RWMutex{},
		ctx:        ctx,
	}
}

const (
	// ErrInvalidParam is returned when a parameter is invalid
	ErrInvalidParam = 300
	// ErrNotFound is returned when a parameter is not found
	ErrNotFound = 404
	// OK all good
	OK = 200
	// ErrServerFault is returned when server error occurs
	ErrServerFault = 500
)

func parseAndValidateJobNsName(in *profiling.DataTypeReq) (string, string, error) {
	jobNsName := in.GetJobNsName()
	jobNameInfo := strings.Split(jobNsName, "/")
	if len(jobNameInfo) != PartsOfJobNs {
		return "", "", fmt.Errorf("the format of jobNsName is not namespace/jobName")
	}
	return jobNameInfo[0], jobNameInfo[1], nil
}

func (ps *SwitchManager) sendProfilingSwitchByCm(
	dtc *profile.DataTraceController, in *profiling.DataTypeReq) (*profiling.DataTypeRes, error) {
	notifyErr := ps.notifySubscriber(dtc, in)
	if notifyErr != nil {
		hwlog.RunLog.Warnf("sendProfilingSwitchByCm notify subscribe failed: %v", notifyErr)
	}
	cmName := profile.DataTraceCmPrefix + dtc.JobName
	owner := getPGOwner(dtc.JobNamespace, dtc.JobName)
	if cm, err := kube.GetConfigMap(cmName, dtc.JobNamespace); cm == nil || err != nil {
		if !errors.IsNotFound(err) {
			hwlog.RunLog.Errorf("sendProfilingSwitchByCm get cm failed: %v", err)
			return &profiling.DataTypeRes{Message: fmt.Sprintf("failed to found configmap:[%s/%s]",
				dtc.JobNamespace, dtc.JobName), Code: ErrNotFound}, err
		}
		createErr := dtc.CreateDataTraceCm(in.ProfilingSwitch, owner)
		if createErr != nil {
			hwlog.RunLog.Errorf("sendProfilingSwitchByCm create cm failed: %v", createErr)
			return &profiling.DataTypeRes{Message: fmt.Sprintf("failed to create comfigmap:[%s/%s]",
					dtc.JobNamespace, dtc.JobName), Code: ErrServerFault},
				fmt.Errorf("create cm err:%v", createErr)
		}
	} else {
		updateCmErr := dtc.UpdateDataTraceCm(in.ProfilingSwitch, owner)
		if updateCmErr != nil {
			hwlog.RunLog.Errorf("sendProfilingSwitchByCm update cm failed: %v", updateCmErr)
			return &profiling.DataTypeRes{Message: fmt.Sprintf("failed to update comfigmap:[%s/%s]",
					dtc.JobNamespace, dtc.JobName), Code: ErrServerFault},
				fmt.Errorf("update cm err:%v", updateCmErr)
		}
	}
	return &profiling.DataTypeRes{Message: "successfully changed profiling marker enable status", Code: OK}, nil
}

func (ps *SwitchManager) sendProfilingSwitchByGrpc(
	dtc *profile.DataTraceController, in *profiling.DataTypeReq) (*profiling.DataTypeRes, error) {
	notifyErr := ps.notifySubscriber(dtc, in)
	if notifyErr != nil {
		hwlog.RunLog.Errorf("sendProfilingSwitchByGrpc notify subscribe failed: %v", notifyErr)
		return &profiling.DataTypeRes{Message: fmt.Sprintf("notify subscribe failed"), Code: ErrServerFault},
			fmt.Errorf("notify subscribe err %v", notifyErr)
	}
	return &profiling.DataTypeRes{Message: "successfully changed profiling marker enable status", Code: OK}, nil
}

// ModifyTrainingDataTraceSwitch to modify the profiling marker status by updating the cm
func (ps *SwitchManager) ModifyTrainingDataTraceSwitch(ctx context.Context,
	in *profiling.DataTypeReq) (*profiling.DataTypeRes, error) {
	event := "modify profiling marker status"
	logs.RecordLog("", event, constant.Start)
	res := constant.Failed
	defer func() {
		logs.RecordLog("", event, res)
	}()
	jobNs, jobName, err := parseAndValidateJobNsName(in)
	if err != nil {
		return &profiling.DataTypeRes{Message: err.Error(), Code: ErrInvalidParam}, err
	}
	dtc := profile.NewDataTraceController(jobNs, jobName)
	var response *profiling.DataTypeRes
	isPodsMountProfilingCmPath, err := dtc.IsPodsMountProfilingCmPath()
	if !isPodsMountProfilingCmPath {
		hwlog.RunLog.Warnf("job not mount profiling cm path: %v", err)
		response, err = ps.sendProfilingSwitchByGrpc(dtc, in)
	} else {
		response, err = ps.sendProfilingSwitchByCm(dtc, in)
	}

	if err != nil {
		hwlog.RunLog.Errorf("send profiling switch by cm failed: %v", err)
	} else {
		res = constant.Success
		hwlog.RunLog.Infof("successfully changed profiling marker enable status: %#v", in.ProfilingSwitch)
	}
	return response, err
}

func getPGOwner(jobNs, jobName string) v1.OwnerReference {
	jobInfo := job.GetJobByNameSpaceAndName(jobName, jobNs)
	pgInfo := podgroup.GetPodGroup(jobInfo.Key)
	owner, err := podgroup.GetOwnerRefByPG(&pgInfo)
	if err != nil {
		hwlog.RunLog.Errorf("get owner from pg failed, error: %v", err)
		return v1.OwnerReference{}
	}
	return owner
}

func (ps *SwitchManager) notifySubscriber(dtc *profile.DataTraceController, in *profiling.DataTypeReq) error {
	jobInfo := job.GetRunningJobByNameSpaceAndName(dtc.JobName, dtc.JobNamespace)
	if len(jobInfo.Key) == 0 {
		return fmt.Errorf("no such job")
	}
	if err := ps.publish(jobInfo.Key, in.ProfilingSwitch); err != nil {
		return err
	}
	return nil
}

func (ps *SwitchManager) publish(jobId string, info *profiling.ProfilingSwitch) error {
	publisher, ok := ps.getPublisher(jobId)
	if ok {
		updateMsg := profiling.DataStatusRes{
			Message:         "update profiling switch",
			ProfilingSwitch: info,
			Code:            OK,
		}
		ok := publisher.SaveData(jobId, &updateMsg)
		if !ok {
			errSaveSwitchFailed := "send update profiling switch %v for job %v fail"
			hwlog.RunLog.Errorf(errSaveSwitchFailed, info, jobId)
			return fmt.Errorf(errSaveSwitchFailed, info, jobId)
		}
		return nil
	}
	errPublisherNotFound := "publisher for job %s not found"
	hwlog.RunLog.Errorf(errPublisherNotFound, jobId)
	return fmt.Errorf(errPublisherNotFound, jobId)
}

// GetTrainingDataTraceSwitch get  current profiling marker status
func (ps *SwitchManager) GetTrainingDataTraceSwitch(ctx context.Context,
	in *profiling.DataStatusReq) (*profiling.DataStatusRes, error) {
	event := "get profiling switch info"
	logs.RecordLog("", event, constant.Start)
	res := constant.Failed
	defer func() {
		logs.RecordLog("", event, res)
	}()

	jobNsName := in.GetJobNsName()
	jobNameInfo := strings.Split(jobNsName, "/")
	if len(jobNameInfo) != PartsOfJobNs {
		return &profiling.DataStatusRes{Message: "the format of jobNsName is not namespace/jobName",
			Code: ErrInvalidParam}, nil
	}
	jobNs, jobName := jobNameInfo[0], jobNameInfo[1]
	dtc := profile.NewDataTraceController(jobNs, jobName)
	cm, err := dtc.IsDataTraceCmExist()
	if cm == nil || err != nil {
		hwlog.RunLog.Errorf("can not find data trace configmap[%s/%s]", dtc.JobNamespace, dtc.JobName)
		return &profiling.DataStatusRes{
			Message: fmt.Sprintf("failed to found configmap:[%s/%s]", dtc.JobNamespace, dtc.JobName),
			Code:    ErrNotFound}, err
	}
	data, ok := cm.Data[profile.DataTraceCmProfilingSwitchKey]
	if !ok {
		hwlog.RunLog.Infof("data trace configmap[%s/%s] has no %s field",
			dtc.JobNamespace, dtc.JobName, profile.DataTraceCmProfilingSwitchKey)
		return &profiling.DataStatusRes{Message: "data trace configmap does not contain the 'profilingSwitch' field",
			Code: ErrNotFound}, nil
	}
	var resProfile profiling.ProfilingSwitch
	if err := json.Unmarshal([]byte(data), &resProfile); err != nil {
		hwlog.RunLog.Errorf("failed to unmarshal configmap[%s/%s], err:%v", dtc.JobNamespace, dtc.JobName, err)
		return &profiling.DataStatusRes{
			Message: fmt.Sprintf("failed to convert configmap:[%s/%s]", dtc.JobNamespace, dtc.JobName),
			Code:    ErrServerFault}, err
	}
	res = constant.Success
	return &profiling.DataStatusRes{
		Message:         fmt.Sprintf("successfully get the status of job[%s/%s]", dtc.JobNamespace, dtc.JobName),
		ProfilingSwitch: &resProfile,
		Code:            OK}, nil
}

// SubscribeDataTraceSwitch subscribe profiling date trace switch
func (ps *SwitchManager) SubscribeDataTraceSwitch(
	clientInfo *profiling.ProfilingClientInfo, stream profiling.TrainingDataTrace_SubscribeDataTraceSwitchServer) error {
	event := "subscribe profiling message signal"
	logs.RecordLog(clientInfo.Role, event, constant.Start)
	res := constant.Failed
	defer func() {
		logs.RecordLog("", event, res)
	}()

	hwlog.RunLog.Infof("receive Subscribe profiling message signal request, %v", clientInfo)
	publisher, ok := ps.getPublisher(clientInfo.JobId)
	if !ok || publisher == nil {
		_, err := ps.preRegistry(clientInfo)
		if err != nil {
			hwlog.RunLog.Errorf("jobId=%s, preCheck err:%v", clientInfo.JobId, err)
			return fmt.Errorf("jobId=%s not registered, role=%s", clientInfo.JobId, clientInfo.Role)
		}
	}
	publisher = ps.preemptPublisher(clientInfo.JobId)
	publisher.ListenDataChange(stream)
	ps.deletePublisher(clientInfo.JobId, publisher.GetCreateTime())
	hwlog.RunLog.Infof("jobId=%s stop subscribe fault message signal, createTime=%v",
		clientInfo.JobId, publisher.GetCreateTime().UnixNano())
	res = constant.Success
	return nil
}

func (ps *SwitchManager) getPublisher(jobId string) (*config.ConfigPublisher[*profiling.DataStatusRes], bool) {
	ps.lock.RLock()
	defer ps.lock.RUnlock()
	publisher, ok := ps.publishers[jobId]
	return publisher, ok
}

func (ps *SwitchManager) preRegistry(req *profiling.ProfilingClientInfo) (common.RespCode, error) {
	_, ok := job.GetJobCache(req.JobId)
	_, err := job.GetNamespaceByJobIdAndAppType(req.JobId, req.Role)
	if !ok && err != nil {
		hwlog.RunLog.Errorf("jobId=%s not exist and is not multi-instance job", req.JobId)
		return common.JobNotExist, fmt.Errorf("jobId=%s not exist and is not multi-instance", req.JobId)
	}
	if ps.serveJobNum() >= constant.MaxServeJobs {
		return common.OutOfMaxServeJobs,
			fmt.Errorf("jobId=%s out of max serve jobs", req.JobId)
	}
	return common.OK, nil
}

func (ps *SwitchManager) serveJobNum() int {
	ps.lock.RLock()
	defer ps.lock.RUnlock()
	return len(ps.publishers)
}

func (ps *SwitchManager) preemptPublisher(jobId string) *config.ConfigPublisher[*profiling.DataStatusRes] {
	ps.lock.Lock()
	defer ps.lock.Unlock()
	publisher, ok := ps.publishers[jobId]
	if ok && publisher != nil {
		publisher.Stop()
	}
	newPublisher := config.NewConfigPublisher[*profiling.DataStatusRes](jobId,
		ps.ctx, constant.ProfilingDataType, nil)
	ps.publishers[jobId] = newPublisher
	return newPublisher
}

func (ps *SwitchManager) deletePublisher(jobId string, createTime time.Time) {
	ps.lock.Lock()
	defer ps.lock.Unlock()
	publisher, ok := ps.publishers[jobId]
	if !ok || publisher == nil || !createTime.Equal(publisher.GetCreateTime()) {
		return
	}
	delete(ps.publishers, jobId)
}
