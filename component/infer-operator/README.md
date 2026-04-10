# Infer Operator

# 组件介绍
- Infer Operator 是一个Kubernetes Operator，用于部署和管理多角色合作的推理任务。Infer Operator 定义了InferServiceSet, InferService和InstanceSet三种CRD, 并实现了三种资源的控制器用于调谐三种资源实例状态。
  
# 编译Infer Operator
1.  通过git拉取源码，获得infer-operator。

    示例：源码放在/home/mind-cluster/component/infer-operator目录下

2.  执行以下命令，进入构建目录，执行构建脚本，在“output“目录下生成二进制infer-operator、yaml文件和Dockerfile。

    **cd** _/home/mind-cluster/component/_**infer-operator/build/**

    **chmod +x build.sh**

    **./build.sh**
3.  执行以下命令，查看**output**生成的软件列表。

    **ll** _/home/mind-cluster/component/_**infer-operator/output**

    ```
    drwxr-xr-x 2 root root     4096 Jan 29 19:12 ./
    drwxr-xr-x 9 root root     4096 Jan 29 19:09 ../
    -r-x------ 1 root root 43524664 Jan 29 19:09 infer-operator 
    -r-------- 1 root root   372080 Jan 29 19:09 infer-operator-v6.0.0.yaml
    -r-------- 1 root root      482 Jan 29 19:12 Dockerfile
    ```

# 说明

1. 当前容器方式部署本组件，本组件的认证鉴权方式为ServiceAccount， 该认证鉴权方式为ServiceAccount的token明文显示，如果需要加密保存，请自行修改
