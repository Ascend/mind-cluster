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

// Package container for monitoring containers' npu allocation
package container

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"ascend-common/common-utils/limiter"
	"huawei.com/npu-exporter/v6/utils/logger"
)

const (
	httpUnixPre    = "http://unix/"
	reqTimeout     = 60 * time.Second
	headerTimeout  = 10 * time.Second
	maxHeaderBytes = 1024
)

type dockerCli struct {
	http.Client
	Host string
}

type dockerContainerRes struct {
	ID         string
	HostConfig *HostConfig
}

type dockerVersionRes struct {
	Version string
}

type HostConfig struct {
	CgroupParent string
}

func createDockerCli() *dockerCli {
	client := http.Client{
		Timeout: reqTimeout,
		Transport: &http.Transport{
			ResponseHeaderTimeout:  headerTimeout,
			MaxResponseHeaderBytes: maxHeaderBytes,
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, "unix", strings.TrimPrefix(defaultDockerdAddr, unixPre))
			},
		},
	}
	return &dockerCli{
		Client: client,
		Host:   httpUnixPre,
	}
}

// send http request to docker to get container info
func (d *dockerCli) inspectContainer(id string) (dockerContainerRes, error) {
	path := d.Host + fmt.Sprintf("containers/%s/json", id)
	var res dockerContainerRes
	err := d.get(path, &res)
	return res, err
}

// send http request to docker to get docker version
func (d *dockerCli) getDockerVersion() (string, error) {
	path := d.Host + "version"
	var res dockerVersionRes
	err := d.get(path, &res)
	return res.Version, err
}

func (d *dockerCli) get(path string, obj any) error {
	response, reader, err := d.doGet(path)
	if err != nil {
		return err
	}
	defer func() {
		if response.Body != nil {
			err := response.Body.Close()
			if err != nil {
				logger.Errorf("close response body failed, err: %v", err)
			}
		}
	}()
	err = json.NewDecoder(reader).Decode(obj)
	if err != nil {
		logger.Errorf("decode docker version info failed, err: %v", err)
		return err
	}
	return err
}

func (d *dockerCli) doGet(url string) (*http.Response, io.Reader, error) {
	response, err := d.Client.Get(url)
	if err != nil {
		return nil, nil, err
	}
	reader := io.LimitReader(response.Body, limiter.DefaultDataLimit)
	return response, reader, nil
}
