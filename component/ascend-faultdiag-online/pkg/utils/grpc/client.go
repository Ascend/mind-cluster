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
	"fmt"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/apimachinery/pkg/util/uuid"

	"ascend-common/common-utils/hwlog"
	"ascend-faultdiag-online/pkg/utils"
	"ascend-faultdiag-online/pkg/utils/constants"
	"ascend-faultdiag-online/pkg/utils/grpc/profiling"
	"ascend-faultdiag-online/pkg/utils/grpc/pubfault"
)

const (
	on        = "on"
	off       = "off"
	maxRetry  = 20
	sleepTime = 3 * time.Second
)

// Client is a grpc client struct
type Client struct {
	conn *grpc.ClientConn
	tc   profiling.TrainingDataTraceClient
	pf   pubfault.PubFaultClient
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
	return nil
}

// Close is a function to close the grpc connection
func (c *Client) Close() {
	if c.conn == nil {
		return
	}
	c.conn.Close()
}

// StartAllProfiling start all the profiling, including heavy&light profiling
func (c *Client) StartAllProfiling(name, namespace string) error {
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

// StopAllProfiling stop all the profiling, only occurs the job closes
func (c *Client) StopAllProfiling(name, namespace string) error {
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
	_, err := c.profilingSwitch(data)
	return err
}

// StartHeavyProfiling start the heavy profiling, it has big impact on the performance
func (c *Client) StartHeavyProfiling(name, namespace string) error {
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
	var res *profiling.DataTypeRes
	var err error
	for count := 0; count < maxRetry; count++ {
		res, err = c.tc.ModifyTrainingDataTraceSwitch(context.Background(), data)
		if err == nil {
			return res, nil
		}
		if count < maxRetry-1 {
			hwlog.RunLog.Warnf("[FD-OL]call grpc failed: %v, retry: %d/%d", err, count+1, maxRetry)
			time.Sleep(sleepTime)
		}
	}
	hwlog.RunLog.Errorf("[FD-OL]reached the max try: %d, and got err: %v", maxRetry, err)
	return nil, err
}

// ReportFault report fault to clusterd
func (c *Client) ReportFault(faults []*pubfault.Fault) error {
	req := pubfault.PublicFaultRequest{
		Id:        string(uuid.NewUUID()),
		Timestamp: time.Now().UnixMilli(),
		Version:   "1.0",
		Resource:  "fd-online",
		Faults:    faults,
	}

	_, err := c.SendToPubFaultCenter(&req)
	return err
}

// SendToPubFaultCenter send fault to public fault center
func (c *Client) SendToPubFaultCenter(data *pubfault.PublicFaultRequest) (*pubfault.RespStatus, error) {
	var res *pubfault.RespStatus
	var err error
	for count := 0; count < maxRetry; count++ {
		res, err = c.pf.SendPublicFault(context.Background(), data)
		if err == nil {
			return res, nil
		}
		if count < maxRetry-1 {
			hwlog.RunLog.Warnf("[FD-OL]call grpc failed: %v, retry: %d/%d", err, count+1, maxRetry)
			time.Sleep(sleepTime)
		}
	}
	hwlog.RunLog.Errorf("[FD-OL]reached the max try: %d, and got err: %v", maxRetry, err)
	return nil, err
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
