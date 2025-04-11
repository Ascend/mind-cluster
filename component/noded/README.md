# NodeD

# 组件介绍

    NodeD是一个检测节点状态异常的组件，负责从ipmi获取到计算节点的CPU、内存、硬盘的故障信息，上报给ClusterD。

# 编译NodeD

1.  通过git拉取源码，获得noded。

    示例：源码放在/home/mind-cluster/component/noded目录下

2.  执行以下命令，进入NodeD构建目录，执行构建脚本，在“output“目录下生成二进制noded、yaml文件和Dockerfile文件等。

    **cd** _/home/mind-cluster/component/_**noded/build/**

    **chmod +x build.sh**

    **./build.sh**

3.  执行以下命令，查看**output**生成的软件列表。

    **ll** _/home/mind-cluster/component/_**noded/output**

    ```
    -r--------  1 root root      480 Nov 14 07:10 Dockerfile
    -r-x------  1 root root 36550304 Nov 14 07:10 noded
    -r--------  1 root root      434 Nov 14 07:10 NodeDConfiguration.json
    -r--------  1 root root     2883 Nov 14 07:10 noded-v6.0.RC1.yaml
    -r--------  1 root root      273 Nov 14 07:10 pingmesh-config.yaml
    ```

# 说明

1. 当前容器方式部署本组件，本组件的认证鉴权方式为ServiceAccount， 该认证鉴权方式为ServiceAccount的token明文显示，建议用户自行进行安全加强。
2. 当前特权容器方式部署，该容器权限具有一定风险，建议用户自行进行安全加强。