# Preparing the Installation Environment

## Notes on Installing Kubernetes

- When Kubernetes uses Calico as the cluster network plugin, the default network configuration is node-to-node mesh. In large-scale clusters, this configuration may cause excessive network load on service switches. It is recommended to configure it to reflector mode. For details, see the [Calico official documentation](https://docs.tigera.io/calico-enterprise/latest/networking/configuring/bgp#disable-the-default-bgp-node-to-node-mesh).
- When installing Kubernetes on CentOS 7.6 and using Calico v3.24 as the cluster network plugin, the installation may fail. For related constraints, see [System Requirements](https://docs.tigera.io/calico/3.24/getting-started/kubernetes/requirements).
- Starting from Kubernetes 1.24, Dockershim has been removed from the Kubernetes project. If you still want to use Docker as the container engine for Kubernetes, you need to install cri-dockerd. For details, see the [Docker Usage Failure When Using Kubernetes 1.24 or Later](https://gitcode.com/Ascend/mind-cluster/issues/340) section.
- Kubernetes 1.25.10 and later versions do not support the recovery feature of virtualized NPUs.

## Installing the Open-Source System

Before installing the cluster scheduling components, ensure that the following basic environment preparations are complete:

- Install Docker. Versions 18.09.x to 28.5.1 are supported. For details, see [Install Docker](https://docs.docker.com/engine/install/).
- Install Containerd. Versions 1.4.x to 2.1.4 are supported. For details, see [Install Containerd](https://github.com/containerd/containerd/blob/main/docs/getting-started.md).
- Install Kubernetes. Kubernetes versions 1.17.x to 1.34.x are supported (version 1.19.x or later is recommended). For details, see [Install Kubernetes](https://kubernetes.io/docs/setup/production-environment/tools/). It is recommended to [create a cluster using Kubeadm](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/). For some issues during cluster initialization, see [Kubernetes Initialization Failure](https://gitcode.com/Ascend/mind-cluster/issues/338). Also, you need to de-isolate the management node. The command example is as follows.

    - Kubernetes < 1.24
        - De-isolate a single node.

          ```shell
          kubectl taint nodes <hostname> node-role.kubernetes.io/master-
          ```

        - De-isolate all nodes.

          ```shell
          kubectl taint nodes --all node-role.kubernetes.io/master-
          ```

    - Kubernetes ≥ 1.24
        - De-isolate a single node.

          ```shell
          kubectl taint nodes <hostname> node-role.kubernetes.io/control-plane:NoSchedule-
          ```

        - De-isolate all nodes.

          ```shell
          kubectl taint nodes --all node-role.kubernetes.io/control-plane:NoSchedule-
          ```

  >[!NOTE]
  >De-isolating the management node removes the taint from the master node, allowing pods to be scheduled onto the master node.
