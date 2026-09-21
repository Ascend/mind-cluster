# Container Snapshot Deployment and Usage

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-09-01T08:02:30.042Z pushedAt=2026-09-01T08:03:28.198Z -->

This feature implements the container snapshot capability for inference services, supporting fast startup of large model inference services and rapid recovery in failure scenarios. Through the collaboration of MindCluster's Infer Operator, NodeD, and Ascend Docker Runtime components, host-side and device-side snapshots are generated after an inference job completes warm up. When a Pod is abnormally deleted, the service is quickly recovered through the snapshot, reducing the inference service startup time from over 30 minutes to the minute level.

## Before You Start

**Environment Requirements**

The environment requirements for using the container snapshot feature are as follows:

| Item   | Requirement                                                                                                                                        |
|------|-------------------------------------------------------------------------------------------------------------------------------------------|
| OS   | EulerOS R15C10 or HCE3.0, with CRIU 3.19 installed                                                                                                       |
| Container engine   | Containerd 1.6 or later; 1.6 recommended                                                                                                                            |

**Prerequisites**

- To use the container snapshot feature, ensure that the following components are installed.
    - Volcano (This feature supports only Volcano as the scheduler. Other schedulers are not supported.)
    - Ascend Device Plugin
    - Ascend Docker Runtime
    - ClusterD
    - NodeD
    - Infer Operator

- If they are not installed, refer to [Installation and Deployment](../../03_installation_guide/02_installation/00_helm_installation.md) for the procedure. Note that some installation steps need to be modified for NodeD and Infer Operator.

  - NodeD
     - The NodeD image must be built using the following Dockerfile, where `http_proxy` and `https_proxy` are configured as proxies that can access the public network.

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

     - The startup YAML of NodeD must use `noded-container-snapshot.yaml` corresponding to the container snapshot feature in the component package, where the snapshot path `/user/snapshot` is configured according to the actual situation and must be shared storage

        ```Yaml
           - name: image-path
             mountPath: /user/snapshot

           - name: image-path
             hostPath:
               path: /user/snapshot
               type: Directory
       ```

  - Infer Operator
     - Add the snapshot path mount item to the startup YAML of Infer Operator, where `mountPath` and `hostPath` are configured according to the actual situation and must be the same as the snapshot path of NodeD. In addition, `snapshotTimeout` (>= 1, in minutes) can be configured, with a default of 60 minutes.

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

**Usage Notes**

- Container snapshot supports only jobs whose workload is of the `StatefulSet` type. For such jobs, the label `infer.huawei.com/container-snapshot` that enables container snapshot must be added and set to `true`. In addition, the same snapshot path as that of NodeD must be configured in the container environment variables, as shown below:

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

For details about inference job deployment and usage, see [Usage Instructions](https://gitcode.com/Ascend/MindIE-PyMotor/blob/master/docs/zh/user_guide/features/container_snapshot.md). The following demonstration only shows the usage process.

**Supported Product Forms**

The following products support container snapshots.

- <term>Atlas A2 training products</term>

## Use Case

### Viewing Job Processes After a Job Is Successfully Delivered

Run the following command to view the Pod running status.

```shell
kubectl get pod --all-namespaces
```

Command output:

```ColdFusion
NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE
...
default          my-test-hb-0-hybrid-0-0                      1/1     Running   0          20s
default          my-test-hb-0-hybrid-0-1                      1/1     Running   0          20s
default          my-test-hb-0-hybrid-1-0                      1/1     Running   0          20s
default          my-test-hb-0-hybrid-1-1                      1/1     Running   0          20s
...
```

### Viewing Snapshot Generation

View the snapshot in the previously configured snapshot path:

```ColdFusion
[root@master]# cd /user/snapshot
[root@master]# ls default/my-test-hb-0-hybrid/
0 1 snapshot_status.json
[root@master]# ls default/my-test-hb-0-hybrid/0
container.id image rootfs-diff.digest rootfs-diff.tar rootfs-external-diff.tar
```

If no snapshot file is generated, check the Ascend Docker Runtime logs on the corresponding compute node for faults:

```ColdFusion
[root@worker]# cd /var/log/ascend-docker-runtime
[root@worker]# tail -f runtime-run.log
```

### Simulate a Pod Failure and Trigger Rescheduling

Enter the container and terminate the inference service process:

```shell
kill -9 xxx
```

The pod restarts successfully:

```ColdFusion
NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE
...
default          my-test-hb-0-hybrid-0-0                      1/1     Running   0          20s
default          my-test-hb-0-hybrid-0-1                      1/1     Running   0          20s
default          my-test-hb-0-hybrid-1-0                      1/1     Running   0          3m20s
default          my-test-hb-0-hybrid-1-1                      1/1     Running   0          3m20s
...
```

If the Pod fails to restart, check the fault in the restore log of Ascend Docker Runtime on the corresponding compute node:

```ColdFusion
[root@worker]# cd /var/log/ascend-docker-runtime/restore
[root@worker]# cd k8s.io_{container_id}/work
[root@worker]# tail -f restore.log
```

### Ascend Docker Runtime Log Shows Successful Container Restoration

View `runtime-run.log` in the Ascend Docker Runtime log directory:

```ColdFusion
...
[INFO] 1 runtime/runc.go:147 calling runc restore args: [--root /run/containerd/runc/k8s.io --log /run/containerd/io.containerd.runtime.v2.task/k8s.io/xxxxxx/log.json --log-format json --systemd-cgroup restore --bundle /run/con
[INFO] 1 runtime/runc.go:158 calling runc restore xxxxxx success
...
```
