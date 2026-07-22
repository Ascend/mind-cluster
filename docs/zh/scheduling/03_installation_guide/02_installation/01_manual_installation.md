# 手动安装

手动安装部署适用于没有 Helm 或需要使用 kubectl 手动部署组件的场景。手动安装包括如下操作：

1. 获取软件包：下载各组件的安装包。
2. 安装前准备：创建用户、准备镜像、创建日志目录等。
3. 安装基础组件：Ascend Docker Runtime、Container Manager。
4. 安装 K8s 组件：NPU Exporter、Ascend Device Plugin、K8s RDMA Shared Dev Plugin、Volcano、ClusterD、Ascend Operator、Infer Operator、NodeD等。

详细安装步骤，请参见[开发者指南-手动安装部署](../../05_developer_guide/00_installation_deployment/00_manual_installation/menu_manual_installation.md)。
