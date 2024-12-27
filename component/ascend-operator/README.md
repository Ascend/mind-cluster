# Ascend Operator

# 组件介绍
- Ascend Operator 支持mindspore、pytorch、tensorflow三个AI框架在Kubernetes上进行分布式训练的插件。CRD（Custom Resource Definition）中定义了AscendJob任务，用户只需配置yaml文件，
  即可轻松实现分布式训练。
  
# 编译Ascend Operator
1.  通过git拉取源码，获得ascend-operator。

    示例：源码放在/home/mindxdl/component/ascend-operator目录下

2.  执行以下命令，进入构建目录，执行构建脚本，在“output“目录下生成二进制ascend-operator、yaml文件和Dockerfile。

    **cd** _/home/mindxdl/component/_**ascend-operator/build/**

    **chmod +x build.sh**

    **./build.sh**
3.  执行以下命令，查看**output**生成的软件列表。

    **ll** _/home/mindxdl/component/_**ascend-operator/output**

    ```
    drwxr-xr-x 2 root root     4096 Jan 29 19:12 ./
    drwxr-xr-x 9 root root     4096 Jan 29 19:09 ../
    -r-x------ 1 root root 43524664 Jan 29 19:09 ascend-operator
    -r-------- 1 root root   372080 Jan 29 19:09 ascend-operator-v5.0.RC1.yaml
    -r-------- 1 root root      482 Jan 29 19:12 Dockerfile
    ```

# 说明

1. 当前容器方式部署本组件，本组件的认证鉴权方式为ServiceAccount， 该认证鉴权方式为ServiceAccount的token明文显示，如果需要加密保存，请自行修改
