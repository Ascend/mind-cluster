# NodeD

# 组件介绍

    NodeD是一个节点心跳检测组件，当NodeD最近一次上报心跳之后一段时间内未再次上报心跳（大于两次心跳上报间隔的阈值）时，调度组件就会认为NodeD所在的节点故障，从而触发故障重调度。当后续NodeD两次心跳上报间隔小于等于阈值时，则认为NodeD所在节点恢复正常

# 编译NodeD

1.  通过git拉取源码，获得noded。

    示例：源码放在/home/mind-cluster/component/noded目录下

2.  执行以下命令，进入NodeD构建目录，执行构建脚本，在“output“目录下生成二进制noded、yaml文件和Dockerfile文件。

    **cd** _/home/mind-cluster/component/_**noded/build/**

    **chmod +x build.sh**

    **./build.sh**

3.  执行以下命令，查看**output**生成的软件列表。

    **ll** _/home/mind-cluster/component/_**noded/output**

    ```
    drwxr-xr-x  2 root root     4096 Nov 14 07:10 .
    drwxr-xr-x 10 root root     4096 Nov 14 07:10 ..
    -r--------  1 root root      623 Nov 14 07:10 Dockerfile
    -r-x------  1 root root 30522008 Nov 14 07:10 noded
    -r--------  1 root root     3438 Nov 14 07:10 noded-v5.0.0.yaml
    ```

# 说明

1. 当前容器方式部署本组件，本组件的认证鉴权方式为ServiceAccount， 该认证鉴权方式为ServiceAccount的token明文显示，建议用户自行进行安全加强。
2. 当前特权容器方式部署，该容器权限具有一定风险，建议用户自行进行安全加强。