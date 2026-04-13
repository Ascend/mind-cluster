# 准备安装环境<a name="ZH-CN_TOPIC_0000002479386402"></a>

**安装Kubernetes须知<a name="section1188585815256"></a>**

- Kubernetes使用Calico作为集群网络插件时，默认使用node-to-node mesh的网络配置；当集群规模较大时，该配置可能造成业务交换机网络负载过大，建议配置成reflector模式，具体操作请参考[Calico官方文档](https://docs.tigera.io/calico-enterprise/latest/networking/configuring/bgp#disable-the-default-bgp-node-to-node-mesh)。
- 在CentOS  7.6系统上安装Kubernetes，并且使用v3.24版本Calico作为集群网络插件，可能会安装失败，可参考[系统要求](https://docs.tigera.io/calico/3.24/getting-started/kubernetes/requirements)查看相关约束。
- Kubernetes  1.24及以上版本，Dockershim已从Kubernetes项目中移除。如果用户还想继续使用Docker作为Kubernetes的容器引擎，需要再安装cri-dockerd，可参考[使用1.24及以上版本的Kubernetes时，Docker使用失败](../faq.md#使用124及以上版本的kubernetes时docker使用失败)章节进行操作。
- Kubernetes  1.25.10及以上版本，不支持虚拟化的vNPU的恢复使能功能。

**安装开源系统<a name="section1849016313266"></a>**

在安装集群调度组件前，用户需确保完成以下基础环境的准备：

- 安装Docker，支持18.09.x\~28.5.1版本，具体操作请参见[安装Docker](https://docs.docker.com/engine/install/)。
- 安装Containerd，支持1.4.x\~2.1.4版本，具体操作请参见[安装Containerd](https://github.com/containerd/containerd/blob/main/docs/getting-started.md)。
- 安装Kubernetes，支持1.17.x\~1.34.x版本的Kubernetes（推荐使用1.19.x及以上版本），具体操作请参见[安装Kubernetes](https://kubernetes.io/zh/docs/setup/production-environment/tools/)推荐[使用Kubeadm创建集群](https://kubernetes.io/zh-cn/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/)，集群初始化过程中的部分问题可参考[初始化Kubernetes失败](../faq.md#初始化kubernetes失败)。且需解除管理节点隔离。如需解除管理节点隔离，命令示例如下。
    - 解除单节点隔离。

        ```shell
        kubectl taint nodes <hostname> node-role.kubernetes.io/master-
        ```

    - 解除所有节点隔离。

        ```shell
        kubectl taint nodes --all node-role.kubernetes.io/master-
        ```

        >[!NOTE] 
        >通过解除管理节点隔离可移除主节点的污点，以允许Pod被调度到主节点上。
