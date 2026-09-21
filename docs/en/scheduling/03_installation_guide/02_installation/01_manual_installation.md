# Manual Installation

Manual installation and deployment is intended for scenarios where Helm is not available or where components need to be manually deployed using kubectl. Manual installation consists of the following steps:

1. Obtain software packages: Download the installation packages for each component.
2. Pre-installation preparation: Create users, prepare images, create log directories, etc.
3. Install base components: Ascend Docker Runtime and Container Manager.
4. Install K8s components: NPU Exporter, Ascend Device Plugin, K8s RDMA Shared Dev Plugin, Volcano, ClusterD, Ascend Operator, Infer Operator, NodeD, etc.

For detailed installation steps, refer to [Developer Guide - Manual Installation and Deployment](../../05_developer_guide/00_installation_deployment/00_manual_installation/menu_manual_installation.md).
