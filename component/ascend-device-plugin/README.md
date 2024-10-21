# Ascend Device Plugin
-   **[免责声明](#免责声明)**
-   **[支持的产品形态](#支持的产品形态)**
-   **[组件介绍](#组件介绍)**
-   **[编译](#编译)**
-   **[组件安装](#组件安装)**
-   **[说明](#说明)**
-   **[版本更新记录](#版本更新记录)**
-   **[版本配套说明](#版本配套说明)**

# 免责声明
- 本代码仓库中包含多个开发分支，这些分支可能包含未完成、实验性或未测试的功能。在正式发布之前，这些分支不应被用于任何生产环境或依赖关键业务的项目中。请务必仅使用我们的正式发行版本，以确保代码的稳定性和安全性。
  使用开发分支所导致的任何问题、损失或数据损坏，本项目及其贡献者概不负责。
- 正式版本请参考：[Ascend Device Plugin正式release版本](https://gitee.com/ascend/ascend-device-plugin/releases)

# 支持的产品形态

- 支持以下产品使用资源监测
    - Atlas 训练系列产品
    - Atlas A2 训练系列产品
    - Atlas A3 训练系列产品
    - 推理服务器（插Atlas 300I 推理卡）
    - Atlas 推理系列产品（Ascend 310P AI处理器）
    - Atlas 800I A2 推理服务器


# 组件介绍
设备管理插件拥有以下功能：

-   设备发现：支持从昇腾设备驱动中发现设备个数，将其发现的设备个数上报到Kubernetes系统中。支持发现拆分物理设备得到的虚拟设备并上报kubernetes系统。
-   健康检查：支持检测昇腾设备的健康状态，当设备处于不健康状态时，上报到Kubernetes系统中，Kubernetes系统会自动将不健康设备从可用列表中剔除。虚拟设备健康状态由拆分这些虚拟设备的物理设备决定。
-   设备分配：支持在Kubernetes系统中分配昇腾设备；支持NPU设备重调度功能，设备故障后会自动拉起新容器，挂载健康设备，并重建训练任务。

# 编译

1.  通过git拉取源码，并切换master分支，获得ascend-device-plugin。

    示例：源码放在/home/test/ascend-device-plugin目录下

2.  执行以下命令，进入构建目录，根据设备插件应用场景，选择其中一个构建脚本执行，在“output“目录下生成二进制device-plugin、yaml文件和Dockerfile等文件。

    **cd** _/home/test/_**ascend-device-plugin/build/**

     2.1 中心侧场景编译device-plugin（构建镜像，容器启动设备插件场景）
        
        chmod +x build.sh
        
        ./build.sh
        
     2.2 边侧场景编译device-plugin（二进制启动设备插件场景）
        
        chmod +x build_edge.sh
            
        ./build_edge.sh

3.  执行以下命令，查看**output**生成的软件列表。

    **ll** _/home/test/_**ascend-device-plugin/output**

    ```
    drwxr-xr-x  2 root root     4096 Jan 18 17:04 ./
    drwxr-xr-x 12 root root     4096 Jan 18 17:04 ../
    -r-x------  1 root root 36058664 Jan 18 17:04 device-plugin
    -r--------  1 root root     2478 Jan 18 17:04 device-plugin-310P-1usoc-v5.0.RC3.yaml
    -r--------  1 root root     3756 Jan 18 17:04 device-plugin-310P-1usoc-volcano-v5.0.RC3.yaml
    -r--------  1 root root     2478 Jan 18 17:04 device-plugin-310P-v5.0.RC3.yaml
    -r--------  1 root root     3756 Jan 18 17:04 device-plugin-310P-volcano-v5.0.RC3.yaml
    -r--------  1 root root     2131 Jan 18 17:04 device-plugin-310-v5.0.RC3.yaml
    -r--------  1 root root     3431 Jan 18 17:04 device-plugin-310-volcano-v5.0.RC3.yaml
    -r--------  1 root root     2130 Jan 18 17:04 device-plugin-910-v5.0.RC3.yaml
    -r--------  1 root root     3447 Jan 18 17:04 device-plugin-volcano-v5.0.RC3.yaml
    -r--------  1 root root      654 Jan 18 17:04 Dockerfile
    -r--------  1 root root     1199 Jan 18 17:04 Dockerfile-310P-1usoc
    -r--------  1 root root     1537 Jan 18 17:04 run_for_310P_1usoc.sh
    ```

    >![](doc/figures/icon-note.gif) **说明：** 
    1、“ascend-device-plugin/build“目录下的**ascendplugin-910.yaml**文件在“ascend-device-plugin/output/“下生成的对应文件为**device-plugin-910-v5.0.RC3.yaml**，作用是更新版本号。
    2、边侧场景编译仅生成device-plugin二进制文件


# 组件安装
1.  请参考《MindX DL用户指南》(https://www.hiascend.com/software/mindx-dl)
    中的“开发文档 \> 基础调度 \> 集群调度 \> 安装 \> 组件安装”进行。

# 说明

1. 当前容器方式部署本组件，本组件的认证鉴权方式为ServiceAccount， 该认证鉴权方式为ServiceAccount的token明文显示，建议用户自行进行安全加强。

<h2 id="版本更新记录">版本更新记录</h2>

| 版本         | 发布日期       | 修改说明              |
|------------|------------|-------------------|
| v6.0.0-RC2 | 2024-07-16 | 配套MindX 6.0.RC2版本 |
| v5.0.1     | 2024-04-19 | 配套MindX 5.0.1版本   |
| v6.0.0-RC1 | 2024-04-12 | 配套MindX 6.0.RC1版本 |
| v5.0.0     | 2023-12-29 | 配套MindX 5.0.0版本   |
| v5.0.0-RC3 | 2023-10-14 | 配套MindX 5.0.RC3版本 |
| v5.0.0-RC2 | 2023-07-27 | 配套MindX 5.0.RC2版本 |
| v5.0.0-RC1 | 2023-04-10 | 配套MindX 5.0.RC1版本 |
| v3.0.0     | 2023-02-16 | 首次发布              |

# 版本配套说明

版本配套详情请参考：[版本配套详情](https://www.hiascend.com/developer/download/commercial)


