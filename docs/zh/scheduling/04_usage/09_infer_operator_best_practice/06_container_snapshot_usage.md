# 容器快照部署及使用

本特性实现推理服务的容器快照能力，支持大模型推理服务快速启动和故障场景下的快速恢复。通过MindCluster的Infer Operator、NodeD和Ascend Docker Runtime组件协作，在推理任务完成warm up后生成Host和Device侧快照，在异常删除Pod后通过快照快速恢复服务，将推理服务启动时间从30分钟以上缩短至分钟级。

此外，容器快照支持弹性扩缩容场景，扩容时可从快照拉起容器并启动实例，配置请参见[配置基于负载的弹性扩缩容](./05_configuring_elastic_scaling.md)。

## 使用前必读

**环境要求**

使用容器快照特性的环境要求如下：

| 名称   | 要求                                               |
|------|--------------------------------------------------|
| OS   | openEuler 24.03 LTS SP1及以上版本或HCE3.0，安装3.19版本CRIU |
| 容器引擎   | containerd 1.6及以上，建议1.6                          |

**前提条件**

- 使用容器快照特性，需要确保已经安装如下组件。
    - Volcano（本特性只支持使用Volcano作为调度器，不支持使用其他调度器。）
    - Ascend Device Plugin
    - Ascend Docker Runtime
    - ClusterD
    - NodeD
    - Infer Operator

- 若没有安装，可以参考[安装部署](../../03_installation_guide/02_installation/00_helm_installation.md)章节进行操作，其中NodeD、Infer Operator需要修改部分安装步骤。

  - NodeD
    - 需要组件软件包中容器快照特性对应的`Dockerfile-container-snapshot`Dockerfile制作NodeD镜像。
    - NodeD的启动yaml需使用组件软件包中容器快照特性对应的`noded-container-snapshot.yaml`，其中可根据实际快照大小修改CPU和内存资源请求与限制，此外快照路径`/user/snapshot`须根据实际情况配置，且为共享存储路径。

        ```Yaml
      ...
          volumeMounts:
      ...
            - name: image-path
              mountPath: /user/snapshot
      ...
      volumes:
      ...
        - name: image-path
          hostPath:
            path: /user/snapshot
            type: Directory
      ...
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

**系统配置**

使用如下命令检查系统iptables后端类型是否为legacy，

```shell
iptables-restore --version
```

如果为nf_tables需手动加载内核模块，初始化filter表映射：

```shell
modprobe ip6_tables
modprobe ip6table_filter
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

- <term>Atlas A2 训练系列产品</term>
- <term>Atlas A3 训练系列产品</term>

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
