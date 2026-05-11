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
	"context"
	"fmt"
	"net"
	"net/url"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"k8s.io/kubelet/pkg/apis/podresources/v1alpha1"
)

// This file is a local copy of k8s.io/kubernetes/pkg/kubelet/apis/podresources.GetV1alpha1Client
// and k8s.io/kubernetes/pkg/kubelet/util.GetAddressAndDialer (unix variant), following the
// k8s recommendation: "Consumers of the pod resources API should not be importing this package.
// They should copy paste the function in their project."

const unixProtocol = "unix"

// getV1alpha1Client returns a client for the PodResourcesLister grpc service
func getV1alpha1Client(socket string, connectionTimeout time.Duration,
	maxMsgSize int) (v1alpha1.PodResourcesListerClient, *grpc.ClientConn, error) {
	addr, dialer, err := getAddressAndDialer(socket)
	if err != nil {
		return nil, nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialer),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMsgSize)))
	if err != nil {
		return nil, nil, fmt.Errorf("error dialing socket %s: %v", socket, err)
	}
	return v1alpha1.NewPodResourcesListerClient(conn), conn, nil
}

// getAddressAndDialer returns the address parsed from the given endpoint and a context dialer.
func getAddressAndDialer(endpoint string) (string, func(ctx context.Context, addr string) (net.Conn, error), error) {
	protocol, addr, err := parseEndpointWithFallbackProtocol(endpoint, unixProtocol)
	if err != nil {
		return "", nil, err
	}
	if protocol != unixProtocol {
		return "", nil, fmt.Errorf("only support unix socket endpoint")
	}
	return addr, dial, nil
}

func parseEndpointWithFallbackProtocol(endpoint string, fallbackProtocol string) (protocol string, addr string, err error) {
	if protocol, addr, err = parseEndpoint(endpoint); err != nil && protocol == "" {
		fallbackEndpoint := fallbackProtocol + "://" + endpoint
		protocol, addr, err = parseEndpoint(fallbackEndpoint)
	}
	return
}

func dial(ctx context.Context, addr string) (net.Conn, error) {
	return (&net.Dialer{}).DialContext(ctx, unixProtocol, addr)
}

func parseEndpoint(endpoint string) (string, string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", "", err
	}
	switch u.Scheme {
	case "tcp":
		return "tcp", u.Host, nil
	case unixProtocol:
		return unixProtocol, u.Path, nil
	case "":
		return "", "", fmt.Errorf("using %q as endpoint is deprecated, please consider using full url format", endpoint)
	default:
		return u.Scheme, "", fmt.Errorf("protocol %q not supported", u.Scheme)
	}
}
