/*
Copyright(C) 2025. Huawei Technologies Co.,Ltd. All rights reserved.

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

// Package kubeclient a series of k8s function ut
package kubeclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/client-go/rest"
)

const (
	mockHostIP = "127.0.0.1"
)

func TestGetKltPodsURL(t *testing.T) {
	convey.Convey("test getKltPodsURL", t, func() {
		convey.Convey("should return empty string and error when env HOST_IP is empty", func() {
			err := os.Unsetenv(HostIPEnv)
			if err != nil {
				t.Errorf("unset env HOST_IP failed")
			}
			url, err := getKltPodsURL()
			convey.So(url, convey.ShouldBeEmpty)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("should return non-empty string and nil when env HOST_IP is not empty", func() {
			err := os.Setenv(HostIPEnv, mockHostIP)
			if err != nil {
				t.Errorf("set env HOST_IP failed")
			}
			url, err := getKltPodsURL()
			convey.So(url, convey.ShouldNotBeEmpty)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("should return non-empty string and nil when env HOST_IP is valid ipv6", func() {
			err := os.Setenv(HostIPEnv, "::1")
			if err != nil {
				t.Errorf("set env HOST_IP failed")
			}
			url, err := getKltPodsURL()
			convey.So(url, convey.ShouldNotBeEmpty)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestCreateKltPodsReqWithToken(t *testing.T) {
	convey.Convey("test createKltPodsReqWithToken", t, func() {
		convey.Convey("should return nil and error when get host ip failed", func() {
			patch := gomonkey.ApplyFuncReturn(getKltPodsURL, "", errors.New("get host ip failed"))
			defer patch.Reset()
			request, err := createKltPodsReqWithToken()
			convey.So(request, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		commonPatch1 := gomonkey.ApplyFuncReturn(getKltPodsURL, mockHostIP, nil)
		defer commonPatch1.Reset()
		convey.Convey("should return nil and error when new request failed", func() {
			patch := gomonkey.ApplyFuncReturn(http.NewRequest, nil, errors.New("new request failed"))
			defer patch.Reset()
			request, err := createKltPodsReqWithToken()
			convey.So(request, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		commonPatch2 := gomonkey.ApplyFuncReturn(http.NewRequest, &http.Request{Header: http.Header{}}, nil)
		defer commonPatch2.Reset()
		convey.Convey("should return nil and error when load token file failed", func() {
			patch := gomonkey.ApplyFuncReturn(rest.InClusterConfig, nil, fmt.Errorf("build config error"))
			defer patch.Reset()
			request, err := createKltPodsReqWithToken()
			convey.So(request, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("should return request and nil when create request with token success", func() {
			patch := gomonkey.ApplyFuncReturn(rest.InClusterConfig, &rest.Config{BearerToken: ""}, nil)
			defer patch.Reset()
			kubeConfig = &rest.Config{BearerToken: ""}
			request, err := createKltPodsReqWithToken()
			convey.So(request, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestGetPodsByKltPort(t *testing.T) {
	utKubeClient, err := initK8S()
	if err != nil {
		t.Fatal("TestAnnotationReset init kubernetes failed")
	}
	convey.Convey("should return nil and error when create request with token failed", t,
		createReqTokenFailedCase(utKubeClient))
	commonPatch := gomonkey.ApplyFuncReturn(createKltPodsReqWithToken, &http.Request{}, nil)
	defer commonPatch.Reset()
	convey.Convey("should return nil and error when send request failed", t, sendReqFailedCase(utKubeClient))
	convey.Convey("should return nil and error when response status code is not 200", t,
		reqStatusCase(utKubeClient))
	commonPatch = gomonkey.ApplyFuncReturn(createKltPodsReqWithToken, &http.Request{}, nil).
		ApplyMethodReturn(&http.Client{}, "Do",
			&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(nil)}, nil)
	convey.Convey("should return nil and error when read response body failed", t,
		readAllFailedCase(utKubeClient))
	convey.Convey("should return nil and error when read response body return EOF", t,
		readAllEOFCase(utKubeClient))
	convey.Convey("should return nil and error when unmarshal response body failed", t,
		unmarshalFailedCase(utKubeClient))
	convey.Convey("should return pod list and nil when get pods information success", t,
		getPodsByKltSuccessCase(utKubeClient))
}

func createReqTokenFailedCase(utKubeClient *ClientK8s) func() {
	return func() {
		patch := gomonkey.ApplyFuncReturn(createKltPodsReqWithToken, nil,
			errors.New("create request with token failed"))
		defer patch.Reset()
		pods, err := utKubeClient.getPodsByKltPort()
		convey.So(pods, convey.ShouldBeNil)
		convey.So(err, convey.ShouldNotBeNil)
	}
}

func sendReqFailedCase(utKubeClient *ClientK8s) func() {
	return func() {
		patch := gomonkey.ApplyMethodReturn(&http.Client{}, "Do",
			nil, errors.New("send request failed"))
		defer patch.Reset()
		pods, err := utKubeClient.getPodsByKltPort()
		convey.So(pods, convey.ShouldBeNil)
		convey.So(err, convey.ShouldNotBeNil)
	}
}

func reqStatusCase(utKubeClient *ClientK8s) func() {
	return func() {
		patch := gomonkey.ApplyMethodReturn(&http.Client{}, "Do",
			&http.Response{StatusCode: http.StatusInternalServerError, Body: io.NopCloser(nil)}, nil)
		defer patch.Reset()
		pods, err := utKubeClient.getPodsByKltPort()
		convey.So(pods, convey.ShouldBeNil)
		convey.So(err, convey.ShouldNotBeNil)
	}
}

func readAllFailedCase(utKubeClient *ClientK8s) func() {
	return func() {
		patch := gomonkey.ApplyFuncReturn(io.ReadAll, nil, errors.New("read response body failed"))
		defer patch.Reset()
		pods, err := utKubeClient.getPodsByKltPort()
		convey.So(pods, convey.ShouldBeNil)
		convey.So(err, convey.ShouldNotBeNil)
	}
}

func readAllEOFCase(utKubeClient *ClientK8s) func() {
	return func() {
		patch := gomonkey.ApplyFuncReturn(io.ReadAll, nil, io.EOF)
		defer patch.Reset()
		pods, err := utKubeClient.getPodsByKltPort()
		convey.So(pods, convey.ShouldBeNil)
		convey.So(err, convey.ShouldNotBeNil)
	}
}

func unmarshalFailedCase(utKubeClient *ClientK8s) func() {
	return func() {
		patch := gomonkey.ApplyFuncReturn(io.ReadAll, []byte("invalid json"), nil).
			ApplyFuncReturn(json.Unmarshal, errors.New("unmarshal response body failed"))
		defer patch.Reset()
		pods, err := utKubeClient.getPodsByKltPort()
		convey.So(pods, convey.ShouldBeNil)
		convey.So(err, convey.ShouldNotBeNil)
	}
}

func getPodsByKltSuccessCase(utKubeClient *ClientK8s) func() {
	return func() {
		patch := gomonkey.ApplyFuncReturn(io.ReadAll,
			[]byte(`{"items": [{"metadata": {"name": "pod1"}}, {"metadata": {"name": "pod2"}}]}`), nil)
		defer patch.Reset()
		pods, err := utKubeClient.getPodsByKltPort()
		convey.So(pods, convey.ShouldNotBeNil)
		convey.So(err, convey.ShouldBeNil)
	}
}
