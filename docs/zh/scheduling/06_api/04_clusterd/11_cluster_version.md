# 集群版本信息

## ConfigMap说明

ClusterD启动后，会创建或更新名为component-versions的ConfigMap，详细说明如下表：

| 参数                 | 说明                                                        |
|--------------------|-----------------------------------------------------------|
| ascend-for-volcano | 由Ascend-for-Volcano上报的版本详细信息                              |
| - version          | 版本号                                                       |
| - gitCommit        | git 提交记录ID                                                |
| - gitBranch        | git 分支名称                                                  |
| - buildOS          | 构建时操作系统                                                   |
| - buildArch        | 构建时系统架构                                                   |
| - goVersion        | Go 编译器版本                                                  |
| ascend-operator    | 由Ascend Operator上报的版本详细信息，详细字段说明同上                        |
| clusterd           | 由ClusterD上报的版本详细信息，详细字段说明同上                               |
| infer-operator     | 由Infer Operator上报的版本详细信息，详细字段说明同上                         |
| device-plugin      | 由ClusterD聚合的集群Ascend Device Plugin组件版本详细信息                |
| - type             | 组件类型，取值为DaemonSet                                         |
| - versions        | 各版本号数量统计                                                  |
| - totalNodes        | 集群节点数量                                                    |
| - queryCommand          | 查询节点annotation中对应组件版本详细信息命令                               |
| noded              | 由ClusterD聚合的集群NodeD组件版本详细信息，详细字段说明同上                      |
| k8s-rdma-shared-dp | 由ClusterD聚合的集群K8s RDMA Shared Dev Plugin组件版本详细信息，详细字段说明同上 |

>[!NOTE]
>
>- 只支持MindCluster 26.2.0及以上版本。
>- TaskD详细版本信息可通过其安装目录的version.py文件查看。
>- MindIO TFT与MindIO ACP详细版本信息可通过其安装目录的VERSION文件查看。
>- 若版本信息ConfigMap被误删，需要重启组件Volcano、Ascend Operator、ClusterD、Infer Operator。
