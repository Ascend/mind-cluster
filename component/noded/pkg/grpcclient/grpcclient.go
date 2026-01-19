/* Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package grpcclient for grpc client
package grpcclient

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"nodeD/pkg/grpcclient/pubfault"
)

// Client is a grpc client struct
type Client struct {
	conn *grpc.ClientConn
	pf   pubfault.PubFaultClient
}

// New get a new grpc client
func New(serverAddr string) (*Client, error) {
	c := Client{}
	var err error
	c.conn, err = grpc.Dial(
		serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return &Client{}, fmt.Errorf("failed to connect to grpc server: %v", err)
	}
	c.pf = pubfault.NewPubFaultClient(c.conn)
	return &c, nil
}

// SendToPubFaultCenter send fault to public fault center
func (c *Client) SendToPubFaultCenter(data *pubfault.PublicFaultRequest) (*pubfault.RespStatus, error) {
	if c == nil || c.pf == nil {
		return nil, errors.New("not found public fault client")
	}
	ctx := context.Background()
	return c.pf.SendPublicFault(ctx, data)
}
