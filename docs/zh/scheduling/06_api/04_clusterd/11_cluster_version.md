# 集群版本信息

## ConfigMap说明

ClusterD启动后，会创建或更新名为component-versions的ConfigMap，详细说明如下表：

| 参数                 | 说明                                       |
|--------------------|------------------------------------------|
| ascend-for-volcano | 由Ascend-for-Volcano上报的版本详细信息             |
| ascend-operator    | 由Ascend Operator上报的版本详细信息                |
| clusterd           | 由ClusterD上报的版本详细信息                       |
| infer-operator     | 由Infer Operator上报的版本详细信息                 |
| device-plugin      | 由ClusterD聚合的集群Ascend Device Plugin组件版本详细信息      |
| noded              | 由ClusterD聚合的集群NodeD组件版本详细信息              |
| k8s-rdma-shared-dp | 由ClusterD聚合的集群K8s RDMA Shared Dev Plugin组件版本详细信息 |

>[!NOTE]
>
>- 只支持MindCluster 26.2.0及以上版本。
>- TaskD详细版本信息可通过其安装目录的version.py文件查看。
>- MindIO TFT与MindIO ACP详细版本信息可通过其安装目录的VERSION文件查看。
>- 若版本信息ConfigMap被误删，需要重启组件Volcano、Ascend Operator、ClusterD、Infer Operator。
