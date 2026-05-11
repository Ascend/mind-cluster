/*
Copyright 2017 The Kubernetes Authors.
Copyright 2018 The Kubernetes Authors.
Copyright 2026. Huawei Technologies Co.,Ltd. All rights reserved.

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

package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// parseEndpointTestCase parseEndpoint test case
type parseEndpointTestCase struct {
	Name             string
	Endpoint         string
	ExpectError      bool
	ExpectedProtocol string
	ExpectedAddr     string
}

func buildParseEndpointTestCases() []parseEndpointTestCase {
	return []parseEndpointTestCase{
		{
			Name:             "01-unix socket endpoint, should parse successfully",
			Endpoint:         "unix:///tmp/s1.sock",
			ExpectedProtocol: "unix",
			ExpectedAddr:     "/tmp/s1.sock",
		},
		{
			Name:             "02-tcp endpoint, should parse successfully",
			Endpoint:         "tcp://localhost:15880",
			ExpectedProtocol: "tcp",
			ExpectedAddr:     "localhost:15880",
		},
		{
			Name:             "03-unsupported protocol, should return error",
			Endpoint:         "npipe://./pipe/mypipe",
			ExpectedProtocol: "npipe",
			ExpectError:      true,
		},
		{
			Name:             "04-invalid protocol prefix, should return error",
			Endpoint:         "tcp1://abc",
			ExpectedProtocol: "tcp1",
			ExpectError:      true,
		},
		{
			Name:        "05-malformed endpoint with spaces, should return error",
			Endpoint:    "a b c",
			ExpectError: true,
		},
		{
			Name:        "06-empty scheme endpoint, should return deprecated error",
			Endpoint:    "/tmp/s1.sock",
			ExpectError: true,
		},
	}
}

// TestParseEndpoint for test parseEndpoint
func TestParseEndpoint(t *testing.T) {
	testCases := buildParseEndpointTestCases()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			protocol, addr, err := parseEndpoint(tt.Endpoint)
			assert.Equal(t, tt.ExpectedProtocol, protocol)
			if tt.ExpectError {
				assert.NotNil(t, err, "expected error during parsing %q", tt.Endpoint)
				return
			}
			assert.Nil(t, err, "expected no error during parsing %q", tt.Endpoint)
			assert.Equal(t, tt.ExpectedAddr, addr)
		})
	}
}

// parseEndpointWithFallbackProtocolTestCase parseEndpointWithFallbackProtocol test case
type parseEndpointWithFallbackProtocolTestCase struct {
	Name             string
	Endpoint         string
	FallbackProtocol string
	ExpectError      bool
	ExpectedProtocol string
	ExpectedAddr     string
}

func buildParseEndpointWithFallbackProtocolTestCases() []parseEndpointWithFallbackProtocolTestCase {
	return []parseEndpointWithFallbackProtocolTestCase{
		{
			Name:             "01-unix endpoint, should parse without fallback",
			Endpoint:         "unix:///tmp/s1.sock",
			FallbackProtocol: "unix",
			ExpectedProtocol: "unix",
			ExpectedAddr:     "/tmp/s1.sock",
		},
		{
			Name:             "02-bare path, should fallback to unix protocol",
			Endpoint:         "/tmp/s1.sock",
			FallbackProtocol: "unix",
			ExpectedProtocol: "unix",
			ExpectedAddr:     "/tmp/s1.sock",
		},
		{
			Name:             "03-tcp endpoint, should parse without fallback",
			Endpoint:         "tcp://localhost:9090",
			FallbackProtocol: "unix",
			ExpectedProtocol: "tcp",
			ExpectedAddr:     "localhost:9090",
		},
		{
			Name:             "04-unsupported protocol with non-empty scheme, should not fallback",
			Endpoint:         "npipe://./pipe/mypipe",
			FallbackProtocol: "unix",
			ExpectedProtocol: "npipe",
			ExpectError:      true,
		},
	}
}

// TestParseEndpointWithFallbackProtocol for test parseEndpointWithFallbackProtocol
func TestParseEndpointWithFallbackProtocol(t *testing.T) {
	testCases := buildParseEndpointWithFallbackProtocolTestCases()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			protocol, addr, err := parseEndpointWithFallbackProtocol(tt.Endpoint, tt.FallbackProtocol)
			assert.Equal(t, tt.ExpectedProtocol, protocol)
			if tt.ExpectError {
				assert.NotNil(t, err, "expected error during parsing %q", tt.Endpoint)
				return
			}
			assert.Nil(t, err, "expected no error during parsing %q", tt.Endpoint)
			assert.Equal(t, tt.ExpectedAddr, addr)
		})
	}
}

// getAddressAndDialerTestCase getAddressAndDialer test case
type getAddressAndDialerTestCase struct {
	Name         string
	Endpoint     string
	ExpectError  bool
	ExpectedAddr string
}

func buildGetAddressAndDialerTestCases() []getAddressAndDialerTestCase {
	return []getAddressAndDialerTestCase{
		{
			Name:         "01-unix socket endpoint, should return address and dialer",
			Endpoint:     "unix:///tmp/s1.sock",
			ExpectedAddr: "/tmp/s1.sock",
		},
		{
			Name:         "02-unix socket with different path, should return address and dialer",
			Endpoint:     "unix:///tmp/f6.sock",
			ExpectedAddr: "/tmp/f6.sock",
		},
		{
			Name:         "03-bare path endpoint, should fallback to unix and return address and dialer",
			Endpoint:     "/var/lib/kubelet/pod-resources/kubelet.sock",
			ExpectedAddr: "/var/lib/kubelet/pod-resources/kubelet.sock",
		},
		{
			Name:        "04-tcp endpoint, should return only support unix error",
			Endpoint:    "tcp://localhost:9090",
			ExpectError: true,
		},
		{
			Name:        "05-http endpoint, should return unsupported protocol error",
			Endpoint:    "http://www.test-web.com/",
			ExpectError: true,
		},
		{
			Name:        "06-https endpoint, should return unsupported protocol error",
			Endpoint:    "https://www.test-web.com/",
			ExpectError: true,
		},
		{
			Name:        "07-invalid protocol, should return error",
			Endpoint:    "htta://test-web.com",
			ExpectError: true,
		},
	}
}

// TestGetAddressAndDialer for test getAddressAndDialer
func TestGetAddressAndDialer(t *testing.T) {
	testCases := buildGetAddressAndDialerTestCases()
	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			addr, dialer, err := getAddressAndDialer(tt.Endpoint)
			if tt.ExpectError {
				assert.NotNil(t, err, "expected error during parsing %s", tt.Endpoint)
				return
			}
			assert.Nil(t, err, "expected no error during parsing %s", tt.Endpoint)
			assert.Equal(t, tt.ExpectedAddr, addr)
			assert.NotNil(t, dialer, "expected non-nil dialer")
		})
	}
}
