# 手动卸载

手动卸载包括如下操作：

1. 卸载前确认：确保集群中无正在使用 MindCluster 组件的工作负载。
2. 关闭 pingmesh 灵衢网络检测（可选）。
3. 卸载 K8s 组件：NPU Exporter、Ascend Device Plugin、K8s RDMA Shared Dev Plugin、Volcano、ClusterD、Ascend Operator、Infer Operator、NodeD等。
4. 卸载基础组件：Ascend Docker Runtime、Container Manager。

卸载完成后建议检查相关配置文件和日志目录，确保清理完整。

详细卸载步骤，请查看[开发者指南-手动卸载](../../05_developer_guide/00_installation_deployment/02_uninstallation.md)。
