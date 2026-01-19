# TaskD

# 组件介绍
任务管理插件作为训练&&推理任务的管理组件，负责完成任务的状态管理，拥有以下功能：

-   进程管理：支持训练&&推理业务进程的进程生命周期管理。
-   状态采集：支持任务的状态数据采集，包括训练轻量级profiling采集等功能。

# 编译

1.  通过git拉取源码，并切换master分支，获得taskd。

    示例：源码放在/home/mind-cluster/component/taskd目录下

2.  执行以下命令，进入构建目录，执行构建脚本build.sh，并传入版本号version，在“output“目录下生成taskd-{version}-py3-none-linux_{arch}.whl的二进制文件。

    **cd** _/home/mind-cluster/component/_**taskd/build/**

    **chmod +x build.sh**

    **./build.sh 6.0.0**


3.  执行以下命令，查看**output**生成的软件列表。

    **ll** _/home/mind-cluster/component/_**taskd/output**

    ```
    drwxr-xr-x  2 root root     4096 Jan 18 17:04 ./
    drwxr-xr-x 12 root root     4096 Jan 18 17:04 ../
    -r-x------  1 root root  1270175 Jan 18 17:04 taskd-6.0.0-py3-none-linux_aarch64.whl
    ```
4.  执行以下命令，完成taskd组件的安装

    ```
    pip install taskd-6.0.0-py3-none-linux_aarch64.whl
    ```

# 说明

1. 本组件为python的软件库，用户需要手动安装到自己的python环境中



