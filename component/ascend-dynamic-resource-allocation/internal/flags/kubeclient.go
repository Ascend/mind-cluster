/*
 * Copyright(C) 2026. Huawei Technologies Co.,Ltd. All rights reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package flags

import (
	"flag"
	"fmt"

	coreclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeClientConfig struct {
	KubeConfig   string
	KubeAPIQPS   float64
	KubeAPIBurst int
}

type ClientSets struct {
	Core coreclientset.Interface
}

// RegisterFlags register kubernetes client config flags using standard flag package
func (k *KubeClientConfig) RegisterFlags() {
	flag.StringVar(&k.KubeConfig,
		"kubeconfig",
		"",
		"Absolute path to the `KUBECONFIG` file. Either this flag or the KUBECONFIG env variable need to be set if the driver is being run out of cluster.")

	flag.Float64Var(&k.KubeAPIQPS,
		"kube-api-qps",
		5,
		"`QPS` to use while communicating with the Kubernetes apiserver.")

	flag.IntVar(&k.KubeAPIBurst,
		"kube-api-burst",
		10,
		"`Burst` to use while communicating with the Kubernetes apiserver.")
}

// NewClientSetConfig builds the rest.Config (in-cluster or from kubeconfig).
func (k *KubeClientConfig) NewClientSetConfig() (*rest.Config, error) {
	var csconfig *rest.Config

	var err error
	if k.KubeConfig == "" {
		csconfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("create in-cluster client configuration: %v", err)
		}
	} else {
		csconfig, err = clientcmd.BuildConfigFromFlags("", k.KubeConfig)
		if err != nil {
			return nil, fmt.Errorf("create out-of-cluster client configuration: %v", err)
		}
	}

	csconfig.QPS = float32(k.KubeAPIQPS)
	csconfig.Burst = k.KubeAPIBurst

	return csconfig, nil
}

// NewClientSets builds the client sets (core K8s API) from the configured
// rest.Config, either in-cluster or from a kubeconfig file.
func (k *KubeClientConfig) NewClientSets() (ClientSets, error) {
	csconfig, err := k.NewClientSetConfig()
	if err != nil {
		return ClientSets{}, fmt.Errorf("create client configuration: %v", err)
	}

	coreclient, err := coreclientset.NewForConfig(csconfig)
	if err != nil {
		return ClientSets{}, fmt.Errorf("create core client: %v", err)
	}

	return ClientSets{
		Core: coreclient,
	}, nil
}
