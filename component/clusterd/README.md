# ClusterD

# 组件介绍
提供集群级别的可用资源信息。收集集群任务信息、资源信息和故障信息及影响范围，从任务、芯片和故障维度统计分析

# 编译

1.  通过git拉取源码，并切换master分支，获得ClusterD。

    示例：源码放在/home/mindx/component/clusterd目录下

2.  执行以下命令，进入构建目录，选择构建脚本执行，在“output“目录下生成二进制clusterd、yaml文件和Dockerfile等文件。

    **cd** _/home/mindx/component/_**clusterd/build/**

        chmod +x build.sh
        
        ./build.sh

3.  执行以下命令，查看**output**生成的软件列表。

    **ls** _/home/mindx/component/_**clusterd/output**

    ```
    Ascend-mindxdl-clusterd_xx_linux-xx.zip
    clusterd
    clusterd-v6.0.xx.yaml
    Dockerfile
    ```

    **说明：**
    “clusterd/build“目录下的**xx.zip**文件包含二进制，yaml及Dockerfile文件。

# 说明

- 当前容器方式部署本组件，本组件的认证鉴权方式为ServiceAccount， 该认证鉴权方式为ServiceAccount的token明文显示，建议用户自行进行安全加强。

