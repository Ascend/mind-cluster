# ClusterD
-   **[免责声明](#免责声明)**
-   **[支持的产品形态](#支持的产品形态)**
-   **[组件介绍](#组件介绍)**
-   **[编译ClusterD](#编译ClusterD)**
-   **[组件安装](#组件安装)**
-   **[说明](#说明)**
-   **[版本更新记录](#版本更新记录)**
-   **[版本配套说明](#版本配套说明)**

<h2 id="免责声明">免责声明</h2>
- 本代码仓库中包含多个开发分支，这些分支可能包含未完成、实验性或未测试的功能。在正式发布之前，这些分支不应被用于任何生产环境或依赖关键业务的项目中。请务必仅使用我们的正式发行版本，以确保代码的稳定性和安全性。
  使用开发分支所导致的任何问题、损失或数据损坏，本项目及其贡献者概不负责。
- 正式版本请参考：ClusterD正式release版本 <https://gitee.com/ascend/mindxdl/clusted/releases>

<h2 id="支持的产品形态">支持的产品形态</h2>

- 支持以下产品使用资源监测
    - Atlas 训练系列产品
    - Atlas A2 训练系列产品
    - Atlas A3 训练系列产品
    - 推理服务器（插Atlas 300I 推理卡）
    - Atlas 推理系列产品（Ascend 310P AI处理器）
    - Atlas 800I A2 推理服务器


<h2 id="组件介绍">组件介绍</h2>
提供集群级别的可用资源信息。收集集群任务信息、资源信息和故障信息及影响范围，从任务、芯片和故障维度统计分析

-   xxx：
-   xxx：
-   xxx：

<h2 id="编译ClusterD">编译ClusterD</h2>

1.  通过git拉取源码，并切换master分支，获得ClusterD。

    示例：源码放在/home/test/clusterd目录下

2.  执行以下命令，进入构建目录，选择构建脚本执行，在“output“目录下生成二进制clusterd、yaml文件和Dockerfile等文件。

    **cd** _/home/test/_**clusterd/build/**

        chmod +x build.sh
        
        ./build.sh

3.  执行以下命令，查看**output**生成的软件列表。

    **ls** _/home/test/_**clusterd/output**

    ```
    Ascend-mindxdl-clusterd_xx_linux-xx.zip
    clusterd
    clusterd-v6.0.xx.yaml
    Dockerfile
    ```

    **说明：**
    “clusterd/build“目录下的**xx.zip**文件包含二进制，yaml及Dockerfile文件。


<h2 id="组件安装">组件安装</h2>
1.  请参考《MindX DL用户指南》(https://www.hiascend.com/software/mindx-dl)
    中的“集群调度用户指南 > 安装部署指导 > 安装集群调度组件 > 典型安装场景 > 集群调度场景”进行。

<h2 id="说明">说明</h2>

1. 当前容器方式部署本组件，本组件的认证鉴权方式为ServiceAccount， 该认证鉴权方式为ServiceAccount的token明文显示，建议用户自行进行安全加强。

<h2 id="版本更新记录">版本更新记录</h2>

| 版本         | 发布日期       | 修改说明                   |
|------------|------------|------------------------|
| v6.0.0-RC3 | 2024-xx-xx | 首次发布，配套MindX 6.0.RC3版本 |

<h2 id="版本配套说明">版本配套说明</h2>

版本配套详情请参考：[版本配套详情](https://www.hiascend.com/developer/download/commercial)


