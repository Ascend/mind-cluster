# Manual Uninstallation

Manual uninstallation consists of the following steps:

1. Pre-uninstallation confirmation: Ensure that there are no active workloads using MindCluster components in the cluster.
2. Disable pingmesh UnifiedBus network detection (optional).
3. Uninstall K8s components: NPU Exporter, Ascend Device Plugin, K8s RDMA Shared Dev Plugin, Volcano, ClusterD, Ascend Operator, Infer Operator, NodeD, etc.
4. Uninstall base components: Ascend Docker Runtime and Container Manager.

After uninstallation is complete, it is recommended to check the relevant configuration files and log directories to ensure a thorough cleanup.

For detailed uninstallation steps, please refer to [Developer Guide - Manual Uninstallation](../../05_developer_guide/00_installation_deployment/02_uninstallation.md).
