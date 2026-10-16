# AscendCommon

## 目录

- [简介](#简介)
- [软件架构](#软件架构)
    - [上下游依赖](#上下游依赖)
- [编译指南](#编译指南)
- [安装部署](#安装部署)
- [使用指南](#使用指南)
- [说明](#说明)

## 简介

- AscendCommon是MindCluster集群调度组件的公共代码库，为NPU-Exporter等组件提供可复用的公共能力。
- 主要功能：
    - 提供NPU设备管理能力（基于DCMI/HCCN接口的设备信息查询、拓扑获取、故障码管理等）。
    - 提供CDI（Container Device Interface）规格构建与容器设备挂载能力。
    - 提供Kubernetes API公共定义（标签、注解、常量、Ascend Operator客户端等）。
    - 提供通用工具集（日志、缓存、限流、健康检查、文件/字符串/时间等工具）。

## 软件架构

### 上下游依赖

- AscendCommon以源码形式被其他组件（如NPU Exporter、Ascend Operator等）引用，需与引用组件放置在同一目录下参与编译。
- 运行时依赖宿主节点的Ascend驱动（通过libdcmi等动态库交互）。

## 编译指南

1. 通过git拉取源码，获得ascend-common。

   示例：源码放在/home/mind-cluster/component/ascend-common目录下。

2. AscendCommon为公共库，不单独生成二进制。引用组件（如NPU-Exporter）编译时，需将ascend-common放在同一目录下，随引用组件一起编译。

## 安装部署

AscendCommon为源码公共库，不需要独立安装部署。它随引用组件（如NPU-Exporter等）一起编译、打包并部署，请参考对应组件的安装部署文档。

## 使用指南

1. 引用组件需与ascend-common放置在同一目录下，例如：

    ```text
    component/
    ├── ascend-common/
    └── npu-exporter/
    ```

2. 在引用组件的Go代码中通过模块名`ascend-common`导入所需包，例如：

    ```go
    import (
        "ascend-common/devmanager"
        "ascend-common/common-utils/hwlog"
    )
    ```

## 说明

1. 编译NPU-Exporter等组件时，AscendCommon要放在同一目录下。
2. 不要随意变更ascend-common的函数、常量等，可能导致其他组件编译失败。
