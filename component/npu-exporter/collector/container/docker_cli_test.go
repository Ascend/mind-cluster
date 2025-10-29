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

package container

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"ascend-common/common-utils/hwlog"
	"ascend-common/common-utils/limiter"
	"huawei.com/npu-exporter/v6/utils/logger"
)

const (
	testContainerID2    = "test-container-id"
	testDockerVersion   = "20.10.0"
	testHostConfig      = "test-cgroup-parent"
	testURL             = "http://unix/containers/test-container-id/json"
	testErrorMsg        = "test error"
	testJSONData        = `{"ID":"test-container-id","HostConfig":{"CgroupParent":"test-cgroup-parent"}}`
	testVersionJSONData = `{"Version":"20.10.0"}`
	testInvalidJSON     = `{"invalid": json}`
)

func init() {
	logger.HwLogConfig = &hwlog.LogConfig{OnlyToStdout: true}
	logger.InitLogger("Prometheus")
}

type mockReadCloser struct{ io.Reader }

func (m *mockReadCloser) Close() error { return nil }

type mockHTTPClient struct {
	response *http.Response
	err      error
}

func (m *mockHTTPClient) Get(url string) (*http.Response, error) { return m.response, m.err }

type mockDockerCli struct {
	client mockHTTPClient
	host   string
}

func (m *mockDockerCli) doGet(url string) (*http.Response, io.Reader, error) {
	response, err := m.client.Get(url)
	if err != nil {
		return nil, nil, err
	}
	body := response.Body
	if body == nil {
		return response, nil, nil
	}
	limitedReader := io.LimitReader(body, limiter.DefaultDataLimit)
	return response, limitedReader, nil
}

func (m *mockDockerCli) inspectContainer(id string) (dockerContainerRes, error) {
	path := m.host + "containers/" + id + "/json"
	response, reader, err := m.doGet(path)
	if err != nil {
		return dockerContainerRes{}, err
	}
	defer func() {
		if response.Body != nil {
			err := response.Body.Close()
			if err != nil {
				return
			}
		}
	}()
	var res dockerContainerRes
	if reader != nil {
		err = json.NewDecoder(reader).Decode(&res)
		if err != nil {
			return dockerContainerRes{}, err
		}
	}
	return res, nil
}

func (m *mockDockerCli) getDockerVersion() (string, error) {
	path := m.host + "version"
	response, reader, err := m.doGet(path)
	if err != nil {
		return "", err
	}
	defer func() {
		if response.Body != nil {
			err := response.Body.Close()
			if err != nil {
				return
			}
		}
	}()
	var ver dockerVersionRes
	if reader != nil {
		err = json.NewDecoder(reader).Decode(&ver)
		if err != nil {
			return "", err
		}
	}
	return ver.Version, nil
}

func TestCreateDockerCli(t *testing.T) {
	convey.Convey("TestCreateDockerCli", t, func() {
		convey.Convey("should create docker client successfully when called", func() {
			cli := createDockerCli()
			convey.So(cli, convey.ShouldNotBeNil)
			convey.So(cli.Host, convey.ShouldEqual, httpUnixPre)
			convey.So(cli.Timeout, convey.ShouldEqual, reqTimeout)
		})
	})
}

func TestDockerCliDoGet(t *testing.T) {
	convey.Convey("should return response and reader when http get succeeds", t, func() {
		mockClient := mockHTTPClient{
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       &mockReadCloser{bytes.NewReader([]byte("test data"))},
			},
			err: nil,
		}
		mockCli := &mockDockerCli{client: mockClient, host: "http://unix/"}
		response, reader, err := mockCli.doGet(testURL)

		convey.So(err, convey.ShouldBeNil)
		convey.So(response, convey.ShouldNotBeNil)
		convey.So(reader, convey.ShouldNotBeNil)
		data, err := io.ReadAll(reader)
		if err != nil {
			t.Error(err)
		}
		convey.So(string(data), convey.ShouldEqual, "test data")
	})

	convey.Convey("should return error when http get fails", t, func() {
		mockClient := mockHTTPClient{response: nil, err: errors.New(testErrorMsg)}
		mockCli := &mockDockerCli{client: mockClient, host: "http://unix/"}
		response, reader, err := mockCli.doGet(testURL)

		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, testErrorMsg)
		convey.So(response, convey.ShouldBeNil)
		convey.So(reader, convey.ShouldBeNil)
	})
}

func TestDockerCliInspectContainer(t *testing.T) {
	convey.Convey("should return container info successfully when valid container id provided", t, func() {
		mockClient := mockHTTPClient{
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       &mockReadCloser{bytes.NewReader([]byte(testJSONData))},
			},
			err: nil,
		}
		mockCli := &mockDockerCli{client: mockClient, host: "http://unix/"}
		result, err := mockCli.inspectContainer(testContainerID2)

		convey.So(err, convey.ShouldBeNil)
		convey.So(result, convey.ShouldResemble, dockerContainerRes{
			ID:         testContainerID2,
			HostConfig: &HostConfig{CgroupParent: testHostConfig},
		})
	})

	convey.Convey("should return error when doGet fails", t, func() {
		mockClient := mockHTTPClient{response: nil, err: errors.New(testErrorMsg)}
		mockCli := &mockDockerCli{client: mockClient, host: "http://unix/"}
		result, err := mockCli.inspectContainer(testContainerID2)

		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, testErrorMsg)
		convey.So(result, convey.ShouldResemble, dockerContainerRes{})
	})

	convey.Convey("should return error when json decode fails", t, func() {
		mockClient := mockHTTPClient{
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       &mockReadCloser{bytes.NewReader([]byte(testInvalidJSON))},
			},
			err: nil,
		}
		mockCli := &mockDockerCli{client: mockClient, host: "http://unix/"}
		result, err := mockCli.inspectContainer(testContainerID2)

		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "invalid character")
		convey.So(result, convey.ShouldResemble, dockerContainerRes{})
	})
}

func TestDockerCliGetDockerVersion(t *testing.T) {
	convey.Convey("should return docker version successfully when valid response received", t, func() {
		mockClient := mockHTTPClient{
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       &mockReadCloser{bytes.NewReader([]byte(testVersionJSONData))},
			},
			err: nil,
		}
		mockCli := &mockDockerCli{client: mockClient, host: "http://unix/"}
		result, err := mockCli.getDockerVersion()

		convey.So(err, convey.ShouldBeNil)
		convey.So(result, convey.ShouldEqual, testDockerVersion)
	})

	convey.Convey("should return error when doGet fails", t, func() {
		mockClient := mockHTTPClient{response: nil, err: errors.New(testErrorMsg)}
		mockCli := &mockDockerCli{client: mockClient, host: "http://unix/"}
		result, err := mockCli.getDockerVersion()

		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, testErrorMsg)
		convey.So(result, convey.ShouldEqual, "")
	})

	convey.Convey("should return error when json decode fails", t, func() {
		mockClient := mockHTTPClient{
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       &mockReadCloser{bytes.NewReader([]byte(testInvalidJSON))},
			},
			err: nil,
		}
		mockCli := &mockDockerCli{client: mockClient, host: "http://unix/"}
		result, err := mockCli.getDockerVersion()

		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "invalid character")
		convey.So(result, convey.ShouldEqual, "")
	})
}
