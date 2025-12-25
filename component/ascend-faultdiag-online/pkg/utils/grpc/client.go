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

// Package grpc provides some utility functions for grpc
package grpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/core/model/enum"
	"ascend-faultdiag-online/pkg/model"
	"ascend-faultdiag-online/pkg/utils"
	"ascend-faultdiag-online/pkg/utils/constants"
	"ascend-faultdiag-online/pkg/utils/grpc/job"
	"ascend-faultdiag-online/pkg/utils/grpc/profiling"
	"ascend-faultdiag-online/pkg/utils/grpc/pubfault"
)

const (
	on  = "on"
	off = "off"
)

var storage = utils.NewStorage[*model.JobSummary]()

type callback struct {
	registerId string
	jobName    string
	namespace  string
	f          func(job *model.JobSummary)
}

type jobSummaryWatcher struct {
	mu          sync.Mutex
	isRegisterd bool
	callbacks   *utils.Storage[callback]
	// disconnectedSignal is a signal the grpc stream disconnected(by server) or not
	disconnectedSignal chan struct{}
	// closeSignal is a  signal the stream is close(by client) or not
	closeSignal chan struct{}
}

func (jw *jobSummaryWatcher) reset() {
	jw.mu.Lock()
	jw.isRegisterd = false
	storage.Clear()
	jw.callbacks.Clear()
	jw.mu.Unlock()
}

// Client is a grpc client struct
type Client struct {
	conn *grpc.ClientConn
	tc   profiling.TrainingDataTraceClient
	pf   pubfault.PubFaultClient
	jc   job.JobClient

	jobSummaryWatcher
}

func (c *Client) connect(host string) error {
	if c.conn != nil {
		return nil
	}
	// validate the host
	parsedIp := net.ParseIP(host)
	if parsedIp == nil {
		return fmt.Errorf("invalid host: %s, not the ip type", host)
	}
	var err error
	serverAddr := host + constants.GrpcPort
	c.conn, err = grpc.Dial(
		serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to grpc server: %v", err)
	}
	c.tc = profiling.NewTrainingDataTraceClient(c.conn)
	c.pf = pubfault.NewPubFaultClient(c.conn)
	c.jc = job.NewJobClient(c.conn)
	c.callbacks = utils.NewStorage[callback]()
	return nil
}

// Close is a function to close the grpc connection
func (c *Client) Close() {
	if c == nil || c.conn == nil {
		return
	}
	c.conn.Close()
}

// StartAllProfiling start all the profiling, including heavy&light profiling
func (c *Client) StartAllProfiling(name, namespace string) error {
	if c == nil {
		return errors.New("grpc client is nil")
	}
	data := &profiling.DataTypeReq{
		JobNsName: fmt.Sprintf("%s/%s", namespace, name),
		ProfilingSwitch: &profiling.ProfilingSwitch{
			CommunicationOperator: on,
			Step:                  on,
			SaveCheckpoint:        on,
			FP:                    on,
			DataLoader:            on,
		},
	}
	res, err := c.tc.ModifyTrainingDataTraceSwitch(context.Background(), data)
	hwlog.RunLog.Infof("[FD-OL]got grpc response of start all profiling switch: %v", res)
	return err
}

// StopAllProfiling stop all the profiling, only occurs the job closes
func (c *Client) StopAllProfiling(name, namespace string) error {
	if c == nil {
		return errors.New("grpc client is nil")
	}
	data := &profiling.DataTypeReq{
		JobNsName: fmt.Sprintf("%s/%s", namespace, name),
		ProfilingSwitch: &profiling.ProfilingSwitch{
			CommunicationOperator: off,
			Step:                  off,
			SaveCheckpoint:        off,
			FP:                    off,
			DataLoader:            off,
		},
	}
	res, err := c.tc.ModifyTrainingDataTraceSwitch(context.Background(), data)
	hwlog.RunLog.Infof("[FD-OL]got grpc response of stop all profiling switch: %v", res)
	return err
}

// StartHeavyProfiling start the heavy profiling, it has big impact on the performance
func (c *Client) StartHeavyProfiling(name, namespace string) error {
	if c == nil {
		return errors.New("grpc client is nil")
	}
	data := &profiling.DataTypeReq{
		JobNsName: fmt.Sprintf("%s/%s", namespace, name),
		ProfilingSwitch: &profiling.ProfilingSwitch{
			CommunicationOperator: on,
			Step:                  on,
			SaveCheckpoint:        on,
			FP:                    on,
			DataLoader:            on,
		},
	}
	_, err := c.profilingSwitch(data)
	return err
}

// StopHeavyProfiling stop the heavy profiling, only keep the light profiling
func (c *Client) StopHeavyProfiling(name, namespace string) error {
	if c == nil {
		return errors.New("grpc client is nil")
	}
	data := &profiling.DataTypeReq{
		JobNsName: fmt.Sprintf("%s/%s", namespace, name),
		ProfilingSwitch: &profiling.ProfilingSwitch{
			CommunicationOperator: off,
			Step:                  on,
			SaveCheckpoint:        on,
			FP:                    on,
			DataLoader:            on,
		},
	}
	_, err := c.profilingSwitch(data)
	return err
}

// profilingSwitch is a switch for profiling
func (c *Client) profilingSwitch(data *profiling.DataTypeReq) (*profiling.DataTypeRes, error) {
	return utils.Retry(func() (*profiling.DataTypeRes, error) {
		res, err := c.tc.ModifyTrainingDataTraceSwitch(context.Background(), data)
		hwlog.RunLog.Infof("[FD-OL]got grpc response of profiling switch: %v", res)
		return res, err
	}, nil)
}

// ReportFault report fault to clusterd
func (c *Client) ReportFault(faults []*pubfault.Fault) error {
	if c == nil {
		return errors.New("grpc client is nil")
	}
	req := pubfault.PublicFaultRequest{
		Id:        uuid.New().String(),
		Timestamp: time.Now().UnixMilli(),
		Version:   "1.0",
		Resource:  "fd-online",
		Faults:    faults,
	}

	res, err := c.SendToPubFaultCenter(&req)
	hwlog.RunLog.Infof("[FD-OL]got grpc response of report fault: %v", res)
	return err
}

// SendToPubFaultCenter send fault to public fault center
func (c *Client) SendToPubFaultCenter(data *pubfault.PublicFaultRequest) (*pubfault.RespStatus, error) {
	if c == nil {
		return nil, errors.New("grpc client is nil")
	}
	return utils.Retry(func() (*pubfault.RespStatus, error) {
		return c.pf.SendPublicFault(context.Background(), data)
	}, nil)
}

func (c *Client) registerJobSummary() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.isRegisterd {
		return nil
	}
	// register
	var clientId = uuid.New().String()
	clientInfo := &job.ClientInfo{
		Role:     "FdAgent",
		ClientId: clientId,
	}
	jobStatus, err := utils.Retry(func() (*job.Status, error) {
		return c.jc.Register(context.Background(), clientInfo)
	}, nil)
	if err != nil {
		return err
	}
	clientInfo.ClientId = jobStatus.ClientId

	// sub the jobSummary
	stream, err := utils.Retry(func() (job.Job_SubscribeJobSummarySignalClient, error) {
		return c.jc.SubscribeJobSummarySignal(context.Background(), clientInfo)
	}, nil)
	if err != nil {
		return err
	}
	c.isRegisterd = true
	c.disconnectedSignal = make(chan struct{})
	c.closeSignal = make(chan struct{})
	go c.processJobSummary(stream)
	go c.supervisor()
	return nil
}

func (c *Client) supervisor() {
	select {
	case <-c.disconnectedSignal:
		c.reset()
		hwlog.RunLog.Info("[FD-OL]detected job summary stream disconnected, try to reconnect")
		if err := c.registerJobSummary(); err != nil {
			hwlog.RunLog.Errorf("[FD-OL]registered job summary failed: %v", err)
		}
	case <-c.closeSignal:
		c.reset()
		hwlog.RunLog.Info("[FD-OL]got job summary watcher stop signal, exit supervisor")
	}
}

func (c *Client) processJobSummary(stream job.Job_SubscribeJobSummarySignalClient) {
	for {
		if c.callbacks.Len() == 0 {
			hwlog.RunLog.Info("[FD-OL]detected callbacks are empty, close job summary register")
			if err := stream.CloseSend(); err != nil {
				hwlog.RunLog.Errorf("[FD-OL]close job summary register failed: %v", err)
				time.Sleep(time.Second)
				continue
			}
			c.closeSignal <- struct{}{}
			return
		}
		data, err := stream.Recv()
		if err != nil {
			hwlog.RunLog.Errorf("[FD-OL]job summary stream closed by server: %v", err)
			c.disconnectedSignal <- struct{}{}
			return
		}
		hwlog.RunLog.Infof("[FD-OL]got job summary data from grpc: %v", data)
		// convert JobSummarySignal to model.jobSummary
		job := &model.JobSummary{
			JobId:     data.JobId,
			JobName:   data.JobName,
			Namespace: data.Namespace,
			JobStatus: data.JobStatus,
			Operator:  data.Operator,
		}
		if err = json.Unmarshal([]byte(data.HcclJson), &job.HcclJson); err != nil {
			hwlog.RunLog.Errorf("[FD-OL]json unmarshal hcclJson data: %s failed: %v", data.HcclJson, err)
			continue
		}
		var key = fmt.Sprintf("%s/%s", job.Namespace, job.JobName)
		if data.JobStatus == enum.IsRunning || data.JobStatus == enum.IsPending {
			storage.Store(key, job)
		} else {
			storage.Delete(key)
		}
		c.callbacks.Range(func(key string, cb callback) bool {
			if cb.namespace == data.Namespace && cb.jobName == data.JobName && cb.f != nil {
				go cb.f(job)
			}
			return true
		})
	}
}

// SubscribeJobSummary will subscribe all the job summary
func (c *Client) SubscribeJobSummary(jobName, namespace string, f func(job *model.JobSummary)) (string, error) {
	if f == nil {
		return "", errors.New("callback function is nil")
	}
	// add f to process list
	registerId := uuid.New().String()
	cb := callback{
		registerId: registerId,
		jobName:    jobName,
		namespace:  namespace,
		f:          f,
	}
	var key = fmt.Sprintf("%s/%s", namespace, jobName)
	// send the storage data immediatelly
	if job, ok := storage.Load(key); ok {
		go cb.f(job)
	}
	c.callbacks.Store(registerId, cb)
	// register
	err := c.registerJobSummary()
	if err != nil {
		// unsub
		c.UnsubscribeJobSummary(registerId)
		return "", err
	}
	return registerId, err
}

// UnsubscribeJobSummary will subscribe all the job summary
func (c *Client) UnsubscribeJobSummary(registerId string) {
	c.callbacks.Delete(registerId)
}

var (
	connErr error
	once    sync.Once
	client  *Client
)

// GetClient returns a singleton instance of client
func GetClient() (*Client, error) {
	once.Do(func() {
		client = &Client{}
		connErr = client.connect(utils.GetClusterIp())
	})
	return client, connErr
}
