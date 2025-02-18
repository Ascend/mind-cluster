package kubeclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	mockHostIP = "127.0.0.1"
	twoNum     = 2
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
		convey.Convey("should return nil and error when new request failed", func() {
			patch := gomonkey.ApplyFuncReturn(getKltPodsURL, mockHostIP, nil).
				ApplyFuncReturn(http.NewRequest, nil, errors.New("new request failed"))
			defer patch.Reset()
			request, err := createKltPodsReqWithToken()
			convey.So(request, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("should return nil and error when load token file failed", func() {
			patch := gomonkey.ApplyFuncReturn(getKltPodsURL, mockHostIP, nil).
				ApplyFuncReturn(http.NewRequest, &http.Request{Header: http.Header{}}, nil).
				ApplyFuncReturn(clientcmd.BuildConfigFromFlags, nil, fmt.Errorf("build config error"))
			defer patch.Reset()
			request, err := createKltPodsReqWithToken()
			convey.So(request, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("should return request and nil when create request with token success", func() {
			patch := gomonkey.ApplyFuncReturn(getKltPodsURL, mockHostIP, nil).
				ApplyFuncReturn(http.NewRequest, &http.Request{Header: http.Header{}}, nil).
				ApplyFuncReturn(clientcmd.BuildConfigFromFlags, &rest.Config{BearerToken: ""}, nil)
			defer patch.Reset()
			request, err := createKltPodsReqWithToken()
			convey.So(request, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestGetPodsByKltPortCase01(t *testing.T) {
	convey.Convey("test getPodsByKltPort case 01", t, func() {
		convey.Convey("should return nil and error when create request with token failed", func() {
			patch := gomonkey.ApplyFunc(createKltPodsReqWithToken, func() (*http.Request, error) {
				return nil, errors.New("create request with token failed")
			})
			defer patch.Reset()
			pods, err := getPodsByKltPort()
			convey.So(pods, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("should return nil and error when send request failed", func() {
			patch := gomonkey.ApplyFunc(createKltPodsReqWithToken, func() (*http.Request, error) {
				return &http.Request{}, nil
			}).ApplyMethod(reflect.TypeOf(new(http.Client)), "Do", func(_ *http.Client,
				_ *http.Request) (*http.Response, error) {
				return nil, errors.New("send request failed")
			})
			defer patch.Reset()
			pods, err := getPodsByKltPort()
			convey.So(pods, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("should return nil and error when response status code is not 200", func() {
			patch := gomonkey.ApplyFunc(createKltPodsReqWithToken, func() (*http.Request, error) {
				return &http.Request{}, nil
			}).ApplyMethod(reflect.TypeOf(new(http.Client)), "Do", func(_ *http.Client,
				_ *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusInternalServerError, Body: io.NopCloser(nil)}, nil
			})
			defer patch.Reset()
			pods, err := getPodsByKltPort()
			convey.So(pods, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestGetPodsByKltPortCase02(t *testing.T) {
	convey.Convey("test getPodsByKltPort case 02", t, func() {
		convey.Convey("should return nil and error when read response body failed", func() {
			patch := gomonkey.ApplyFunc(createKltPodsReqWithToken, func() (*http.Request, error) {
				return &http.Request{}, nil
			}).ApplyMethod(reflect.TypeOf(new(http.Client)), "Do", func(_ *http.Client,
				_ *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(nil)}, nil
			}).ApplyFunc(io.ReadAll, func(_ io.Reader) ([]byte, error) {
				return nil, errors.New("read response body failed")
			})
			defer patch.Reset()
			pods, err := getPodsByKltPort()
			convey.So(pods, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("should return nil and error when unmarshal response body failed", func() {
			patch := gomonkey.ApplyFunc(createKltPodsReqWithToken, func() (*http.Request, error) {
				return &http.Request{}, nil
			}).ApplyMethod(reflect.TypeOf(new(http.Client)), "Do", func(_ *http.Client,
				_ *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(nil)}, nil
			}).ApplyFunc(io.ReadAll, func(_ io.Reader) ([]byte, error) {
				return []byte("invalid json"), nil
			}).ApplyFunc(json.Unmarshal, func(data []byte, v any) error {
				return errors.New("unmarshal response body failed")
			})
			defer patch.Reset()
			pods, err := getPodsByKltPort()
			convey.So(pods, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("should return pod list and nil when get pods information success", func() {
			patch := gomonkey.ApplyFunc(createKltPodsReqWithToken, func() (*http.Request, error) {
				return &http.Request{}, nil
			}).ApplyMethod(reflect.TypeOf(new(http.Client)), "Do", func(_ *http.Client,
				_ *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(nil)}, nil
			}).ApplyFunc(io.ReadAll, func(_ io.Reader) ([]byte, error) {
				validJSON := []byte(`{"items": [{"metadata": {"name": "pod1"}}, {"metadata": {"name": "pod2"}}]}`)
				return validJSON, nil
			})
			defer patch.Reset()
			pods, err := getPodsByKltPort()
			convey.So(pods, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
