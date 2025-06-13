# MindCluster
-   [免责说明](#免责说明)
-   [组件介绍](#组件介绍)
-   [支持的产品形态](#支持的产品形态)
-   [编译](#编译)
-   [组件安装](#组件安装)
-   [说明](#说明)
-   [更新日志](#更新日志)
-   [版本配套说明](#版本配套说明)

# 免责说明

- 本仓库代码中包含多个开发分支，这些分支可能包含未完成、实验性或未测试的功能。在正式发布前，这些分支不应被应用于任何生产环境或者依赖关键业务的项目中。请务必使用我们的正式发行版本，以确保代码的稳定性和安全性。
  使用开发分支所导致的任何问题、损失或数据损坏，本项目及其贡献者概不负责。
- 正式版本请参考release版本 <https://gitee.com/ascend/mind-cluster/releases>


# 介绍

    MindCluster（AI集群系统软件）是支持NPU（昇腾AI处理器）训练和推理硬件的深度学习组件，使能构建集群全流程运行，提供NPU集群作业调度、运维监测、故障恢复等功能。深度学习平台开发厂商可以减少底层资源调度相关软件开发工作量，快速使能合作伙伴基于MindCluster开发深度学习平台。

# 支持的产品形态

- 支持以下产品使用资源监测
    - Atlas 训练系列产品
    - Atlas A2 训练系列产品
    - Atlas A3 训练系列产品
    - 推理服务器（插Atlas 300I 推理卡）
    - Atlas 推理系列产品
    - Atlas 800I A2 推理服务器

# 编译

1.  拉取mind-cluster整体源码，例如放在/home目录下。

2.  修改组件版本配置文件service_config.ini中mind-cluster-version字段值为所需编译版本，默认值如下：

        mind-cluster-version=6.0.0

3.  执行以下命令，进入/home/mind-cluster/build目录，选择构建脚本执行：

    **cd /home/mind-cluster/build**

        dos2unix *.sh && chmod +x *.sh
        
        ./build_all.sh $GOPATH

4.  执行完成后进入/home/mind-cluster，在各组件“output”目录下生成编译完成的文件。

5.  此处使用的go版本为1.21。


# 组件安装

1.  进入昇腾社区MindCluster产品界面，点击“查看文档”，再次点击页面上方横向导航栏“集群调度”，进入《MindCluster集群调度用户指南》。在安装和使用前，用户需要提前了解集群调度组件的特性，并根据具体特性选择安装相应的组件。
    
        入口地址：https://www.hiascend.com/software/mindx-dl
    

# 说明

1. 当前容器方式部署本组件，本组件的认证鉴权方式为ServiceAccount， 该认证鉴权方式为ServiceAccount的token明文显示，建议用户自行进行安全加强。
2. 当前特权容器方式部署，该容器权限具有一定风险，建议用户自行进行安全加强。

# 更新日志
该仓库融合不同组件内容。6.0.0之前的版本见各组件仓库：
| 组件         | 链接                                     |  说明   |
|------------|----------------------------------| -|
| NodeD       |  https://gitee.com/ascend/ascend-noded   | -|
| HCCL-Controller |  https://gitee.com/ascend/ascend-hccl-controller   |此组件功能已被Ascend-Operator收编，不建议使用|
| Ascend-Device-Plugin |  https://gitee.com/ascend/ascend-device-plugin   |-|
| NPU-Exporter |  https://gitee.com/ascend/ascend-npu-exporter   |-|
| Ascend-for-Volcano |  https://gitee.com/ascend/ascend-for-volcano   |-|
| Ascend-Docker-Runtime |  https://gitee.com/ascend/ascend-docker-runtime  |-|

6.0.0及之后版本发布如下：
| 版本         | 发布日期      | 修改说明         |
|------------|-----------|----------------------|
| v6.0.0     | 2024-12-15 | 配套MindX 6.0.0版本   |
| v7.0.RC1     | 2025-04-27 | 配套MindX 7.0.RC1版本   |

# 版本配套说明
版本配套详情请参考：[版本配套详情](https://www.hiascend.com/developer/download/community)