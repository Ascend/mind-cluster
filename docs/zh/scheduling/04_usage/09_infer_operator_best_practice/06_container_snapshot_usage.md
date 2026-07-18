# 容器快照部署及使用

本特性实现推理服务的容器快照能力，支持大模型推理服务快速启动和故障场景下的快速恢复。通过MindCluster的Infer Operator、NodeD和Ascend Docker Runtime组件协作，在推理任务完成warm up后生成Host和Device侧快照，在异常删除Pod后通过快照快速恢复服务，将推理服务启动时间从30分钟以上缩短至分钟级。

## 使用前必读

**环境要求**

- 使用容器快照特性的环境要求如下：

| 名称   | 要求                                                                                                                                        |
|------|-------------------------------------------------------------------------------------------------------------------------------------------|
| OS   | EulerOS R15C10或HCE3.0, 安装CRIU版本3.19                                                                                                       |
| 容器引擎   | containerd 1.6及以上，建议1.6                                                                                                                            |

**前提条件**

- 使用容器快照特性，需要确保已经安装如下组件。
    - Volcano（本特性只支持使用Volcano作为调度器，不支持使用其他调度器。）
    - Ascend Device Plugin
    - Ascend Docker Runtime
    - ClusterD
    - NodeD
    - Infer Operator

- 若没有安装，可以参考[安装部署](../../05_developer_guide/00_installation_deployment/00_manual_installation/00_obtaining_software_packages.md)章节进行操作，其中NodeD、Infer Operator需要修改部分安装步骤。

  - NodeD
     - 需要使用如下的Dockerfile制作NodeD镜像，其中http_proxy、https_proxy配置为能够访问公网的代理

        ```Dockerfile
        FROM openeuler-24.03-lts-sp2:latest

        RUN sed -i 's/root:x:0:0:root:\/root:.*$/root:x:0:0:root:\/root:\/sbin\/nologin/' /etc/passwd

        ENV http_proxy=xxx
        ENV https_proxy=xxx
        RUN echo "sslverify=0" >> /etc/yum.conf
        RUN yum makecache
        RUN dnf install -y protobuf-c protobuf libmnl libnftnl libseccomp libnet libnl3 iptables
        ENV http_proxy ""
        ENV https_proxy ""

        ENV LD_LIBRARY_PATH /usr/local/Ascend/driver/lib64:/usr/local/Ascend/driver/lib64/driver:/usr/local/Ascend/driver/lib64/common

        COPY ./noded /usr/local/bin
        COPY ./NodeDConfiguration.json /usr/local/
        COPY ./fdConfig.yaml /usr/local/fdConfig.yaml

        RUN chmod 550 /usr/local/bin/noded &&\
            chmod 550 /usr/local/bin &&\
            chmod 440 /usr/local/NodeDConfiguration.json &&\
            chmod 440 /usr/local/fdConfig.yaml &&\
            echo 'umask 027' >> /etc/profile &&\
            echo 'source /etc/profile' >> ~/.bashrc
        ```

     - NodeD的启动yaml需使用组件软件包中容器快照特性对应的noded-container-snapshot.yaml，其中快照路径/user/snapshot根据实际情况配置并为共享存储

        ```Yaml
           - name: image-path
             mountPath: /user/snapshot

           - name: image-path
             hostPath:
               path: /user/snapshot
               type: Directory
       ```

  - Infer Operator
     - Infer Operator的启动yaml中添加快照路径挂载项，其中mountPath与hostPath根据实际情况配置并与NodeD的快照路径相同，此外还可配置快照超时参数（>= 1，单位为分钟）snapshotTimeout，默认60分钟

        ```Yaml
           - name: image-path
             mountPath: /user/snapshot

           - name: image-path
             hostPath:
               path: /user/snapshot
               type: Directory

           containers:
             - command: [ "/bin/bash", "-c", "--"]
               args: [ "infer-operator
                       --logFile=/var/log/mindx-dl/infer-operator/infer-operator.log
                       --logLevel=0 --snapshotTimeout=120
                       --enable-healthz=true --healthz-address=11254" ]
        ```

**使用说明**

- 容器快照只支持workload为StatefulSet类型任务，且需在该类任务中增加容器快照开启的标签“infer.huawei.com/container-snapshot”，并将其设置为“true”，此外还需在容器环境变量中配置与NodeD相同的快照路径，如下所示：

   ```Yaml
      - name: prefill
        replicas: 1
        workload:
          apiVersion: apps/v1
          kind: StatefulSet
        metadata:
          labels:
            infer.huawei.com/container-snapshot: 'true'

      ...
        spec:
          containers:
          - env:
            - name: host_snapshot_dir_path
              value: "/user/snapshot"
   ```

推理任务部署使用说明详见[使用说明](https://gitcode.com/Ascend/MindIE-PyMotor/blob/master/docs/zh/user_guide/features/container_snapshot.md)，下面使用演示仅展示使用流程。

**支持的产品形态**

支持以下产品使用容器快照。

- Atlas A2 训练系列产品

## 使用演示

### 下发任务成功后查看任务进程

执行以下命令，查看Pod运行状况。

```shell
kubectl get pod --all-namespaces
```

回显示例如下：

```ColdFusion
NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE
...
default          my-test-hb-0-hybrid-0-0                      1/1     Running   0          20s
default          my-test-hb-0-hybrid-0-1                      1/1     Running   0          20s
default          my-test-hb-0-hybrid-1-0                      1/1     Running   0          20s
default          my-test-hb-0-hybrid-1-1                      1/1     Running   0          20s
...
```

### 查看快照生成

在之前配置的快照路径下查看快照：

```ColdFusion
[root@master]# cd /user/snapshot
[root@master]# ls default/my-test-hb-0-hybrid/
0 1 snapshot_status.json
[root@master]# ls default/my-test-hb-0-hybrid/0
container.id image rootfs-diff.digest rootfs-diff.tar rootfs-external-diff.tar
```

如果没有生成快照文件，在对应计算节点上Ascend Docker Runtime日志查看故障：

```ColdFusion
[root@worker]# cd /var/log/ascend-docker-runtime
[root@worker]# tail -f runtime-run.log
```

### 构造pod故障并发生重调度

进入容器杀死推理服务进程：

```shell
kill -9 xxx
```

pod 重启成功：

```ColdFusion
NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE
...
default          my-test-hb-0-hybrid-0-0                      1/1     Running   0          20s
default          my-test-hb-0-hybrid-0-1                      1/1     Running   0          20s
default          my-test-hb-0-hybrid-1-0                      1/1     Running   0          3m20s
default          my-test-hb-0-hybrid-1-1                      1/1     Running   0          3m20s
...
```

如果pod 重启失败，在对应计算节点上Ascend Docker Runtime的restore日志查看故障：

```ColdFusion
[root@worker]# cd /var/log/ascend-docker-runtime/restore
[root@worker]# cd k8s.io_{容器id}/work
[root@worker]# tail -f restore.log
```

### Ascend Docker Runtime日志显示恢复容器成功

查看Ascend Docker Runtime日志目录中runtime-run.log：

```ColdFusion
...
[INFO] 1 runtime/runc.go:147 calling runc restore args: [--root /run/containerd/runc/k8s.io --log /run/containerd/io.containerd.runtime.v2.task/k8s.io/xxxxxx/log.json --log-format json --systemd-cgroup restore --bundle /run/con
[INFO] 1 runtime/runc.go:158 calling runc restore xxxxxx success
...
```
