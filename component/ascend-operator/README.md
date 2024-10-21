# Ascend Operator
-   [免责申明](#免责申明)
-   [组件介绍](#组件介绍)
-   [支持的产品形态](#支持的产品形态)
-   [编译Ascend Operator](#编译Ascend Operator)
-   [组件安装](#组件安装)
-   [说明](#说明)
-   [更新日志](#更新日志)
-   [版本配套说明](#版本配套说明)


# 免责申明
- 本代码仓库中包含多个开发分支，这些分支可能包含未完成、实验性或未测试的功能。在正式发布之前，这些分支不应被用于任何生产环境或依赖关键业务的项目中。请务必仅使用我们的正式发行版本，以确保代码的稳定性和安全性。
  使用开发分支所导致的任何问题、损失或数据损坏，本项目及其贡献者概不负责。
- 正式版本请参考：[Ascend Operator正式release版本](https://gitee.com/ascend/ascend-operator/releases)

# 组件介绍
- Ascend Operator 是MindX DL支持mindspore、pytorch、tensorflow三个AI框架在Kubernetes上进行分布式训练的插件。CRD（Custom Resource Definition）中定义了AscendJob任务，用户只需配置yaml文件，
  即可轻松实现分布式训练。

# 支持的产品形态
- 支持以下产品使用：
    - Atlas 训练系列产品
    - Atlas A2 训练系列产品
    - Atlas A3 训练系列产品
    - Atlas 推理系列产品(Ascend 310P AI处理器)
    - Atlas 800I A2 推理服务器
  
# 编译Ascend Operator
1.  通过git拉取源码，获得ascend-operator。

    示例：源码放在/home/test/ascend-operator目录下

2.  执行以下命令，进入构建目录，执行构建脚本，在“output“目录下生成二进制ascend-operator、yaml文件和Dockerfile。

    **cd** _/home/test/_**ascend-operator/build/**

    **chmod +x build.sh**

    **./build.sh**
3.  执行以下命令，查看**output**生成的软件列表。

    **ll** _/home/test/_**ascend-operator/output**

    ```
    drwxr-xr-x 2 root root     4096 Jan 29 19:12 ./
    drwxr-xr-x 9 root root     4096 Jan 29 19:09 ../
    -r-x------ 1 root root 43524664 Jan 29 19:09 ascend-operator
    -r-------- 1 root root   372080 Jan 29 19:09 ascend-operator-v5.0.RC1.yaml
    -r-------- 1 root root      482 Jan 29 19:12 Dockerfile
    ```

# 组件安装
1.  请参考[《MindX DL用户指南》](https://www.hiascend.com/software/mindx-dl)
    中的“集群调度用户指南 > 安装部署指导 \> 安装集群调度组件 \> 典型安装场景 \> 集群调度场景”进行。

# 说明

1. 当前容器方式部署本组件，本组件的认证鉴权方式为ServiceAccount， 该认证鉴权方式为ServiceAccount的token明文显示，如果需要加密保存，请自行修改

# 更新日志

| 版本       | 发布日期      | 修改说明              |
|----------|-----------|-------------------|
| v6.0.RC2 | 2024-716  | 配套MindX 6.0.RC2版本 |
| v5.0.1   | 2024-518  | 配套MindX 5.0.1版本   |
| v6.0.RC1 | 2024-422  | 配套MindX 6.0.RC1版本 |
| v5.0.0   | 2023-1229 | 配套MindX 5.0.0版本   |
| v5.0.RC3 | 2023-1027 | 首次发布              |

# 版本配套说明
版本配套详情请参考：[版本配套详情](https://www.hiascend.com/developer/download/commercial)