# NodeD
-   [免责说明](#免责说明)
-   [组件介绍](#组件介绍)
-   [编译NodeD](#编译NodeD)
-   [组件安装](#组件安装)
-   [说明](#说明)
-   [更新日志](#更新日志)

# 免责说明

- 本仓库代码中包含多个开发分支，这些分支可能包含未完成、实验性或未测试的功能。在正式发布前，这些分支不应被应用于任何生产环境或者依赖关键业务的项目中。请务必使用我们的正式发行版本，以确保代码的稳定性和安全性。
  使用开发分支所导致的任何问题、损失或数据损坏，本项目及其贡献值概不负责。
- 正式版本请参考ascend-noded正式release版本 <https://gitee.com/ascend/ascend-noded/releases>


# 组件介绍

    NodeD是一个节点心跳检测组件，当NodeD最近一次上报心跳之后一段时间内未再次上报心跳（大于两次心跳上报间隔的阈值）时，调度组件就会认为NodeD所在的节点故障，从而触发故障重调度。当后续NodeD两次心跳上报间隔小于等于阈值时，则认为NodeD所在节点恢复正常

# 编译NodeD

1.  通过git拉取源码，获得ascend-noded。

    示例：源码放在/home/test/ascend-noded目录下

2.  执行以下命令，进入构建目录，执行构建脚本，在“output“目录下生成二进制noded、yaml文件和Dockerfile文件。

    **cd** _/home/test/_**ascend-noded/build/**

    **chmod +x build.sh**

    **./build.sh**

3.  执行以下命令，查看**output**生成的软件列表。

    **ll** _/home/test/_**ascend-noded/output**

    ```
    drwxr-xr-x  2 root root     4096 Nov 14 07:10 .
    drwxr-xr-x 10 root root     4096 Nov 14 07:10 ..
    -r--------  1 root root      623 Nov 14 07:10 Dockerfile
    -r-x------  1 root root 30522008 Nov 14 07:10 noded
    -r--------  1 root root     3438 Nov 14 07:10 noded-v5.0.0.yaml
    ```

# 组件安装

1.  请参考《MindX DL用户指南》
    
        参考地址：https://www.hiascend.com/software/mindx-dl
    
        以手动安装为例：“社区版下载 > 用户手册 > 集群调度安装指南 > 安装部署 > 手动安装 > NodeD”进行。

# 说明

1. 当前容器方式部署本组件，本组件的认证鉴权方式为ServiceAccount， 该认证鉴权方式为ServiceAccount的token明文显示，建议用户自行进行安全加强。
2. 当前特权容器方式部署，该容器权限具有一定风险，建议用户自行进行安全加强。

# 更新日志

| 版本       | 发布日期   | 修改说明       |
| ---------- | ---------- | -------------- |
| v5.0.0 | 2023-1230 | 首次发布 |