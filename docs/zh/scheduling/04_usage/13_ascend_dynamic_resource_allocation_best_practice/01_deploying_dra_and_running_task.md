# DRA机制下资源申请与释放<a name="ZH-CN_TOPIC_000000_dra_deploy_workflow"></a>

## 实现原理<a name="section_dra_principle"></a>

Ascend DRA基于Kubernetes动态资源分配（DRA）机制管理NPU设备，完整的生命周期（部署与发现、用户提交、调度、Prepare、运行、Unprepare）请参见[使用前必读/生命周期](00_before_you_start.md#section_dra_life_cycle)。

## 前提条件<a name="section_dra_prereq"></a>

- 已完成Volcano、Ascend Docker Runtime的安装部署，详细请参见[安装部署](../../03_installation_guide/02_installation/00_helm_installation.md)。
- NPU节点已安装NPU驱动与固件，且设备运行正常。
- 已完成Ascend DRA组件的安装部署。

## 操作步骤<a name="section_dra_steps"></a>

### 创建ResourceClaimTemplate<a name="section_create_template"></a>

用户通过ResourceClaimTemplate定义NPU资源请求，Pod通过引用该模板自动为每个副本创建独立的ResourceClaim。

创建`ascend-npu-claim-template.yaml`文件：

```yaml
apiVersion: resource.k8s.io/v1
kind: ResourceClaimTemplate
metadata:
  name: ascend-npu-template
  namespace: default
spec:
  metadata: {}
  spec:
    devices:
      requests:
      - name: npu              # 请求名称，Pod中通过该名称引用
        deviceClassName: npu.huawei.com   # 引用DRA组件部署时创建的DeviceClass
        count: 1               # 请求的NPU数量，按需修改
```

执行以下命令创建模板：

```bash
kubectl apply -f ascend-npu-claim-template.yaml
```

回显示例如下：

```text
resourceclaimtemplate.resource.k8s.io/ascend-npu-template created
```

> [!NOTE]
> `deviceClassName`取值为`npu.huawei.com`，该DeviceClass由`ascend-dra-driver.yaml`部署时自动创建，其选择器匹配driver为`npu.huawei.com`的所有设备。

### 下发任务并触发Prepare<a name="section_prepare"></a>

下发一个引用ResourceClaimTemplate的Deployment，每个Pod副本会自动创建独立的ResourceClaim。Pod调度到NPU节点后，kubelet会自动调用DRA插件完成Prepare。

创建`dra-test-deployment.yaml`文件：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dra-test
  namespace: default
  labels:
    app: dra-test
spec:
  replicas: 1                               # 副本数，按需修改
  selector:
    matchLabels:
      app: dra-test
  template:
    metadata:
      labels:
        app: dra-test
    spec:
      resourceClaims:
      - name: npu                           # 与容器resources.claims中引用的name一致
        resourceClaimTemplateName: ascend-npu-template
      containers:
      - name: test
        image: ubuntu:22.04
        imagePullPolicy: IfNotPresent
        command:
        - /bin/bash
        - -c
        - sleep inf
        resources:
          claims:
          - name: npu                      # 引用上方resourceClaims中的name
```

1. 下发任务。

   ```bash
   kubectl apply -f dra-test-deployment.yaml
   ```

   回显示例如下：

   ```text
   deployment.apps/dra-test created
   ```

2. 查看Deployment状态，确认Pod进入Running状态。

   ```bash
   kubectl get pod -l app=dra-test -o wide
   ```

   回显示例如下：

   ```text
   NAME                       READY   STATUS    RESTARTS   AGE   IP            NODE          NOMINATED NODE   READINESS GATES
   dra-test-<random-suffix>   1/1     Running   0          30s   10.244.1.20   node-npu-01   <none>           <none>
   ```

3. 验证Prepare已完成。

   查看DRA组件日志，确认`PrepareResourceClaims`被调用且设备已分配（以下日志与代码`plugin.go`和`state.go`中的`Infof`输出一致）：

   ```bash
   kubectl logs -n ascend-dra-driver <dra-pod-name> | grep "PrepareResourceClaims\|devices prepared\|Returning newly prepared"
   ```

   回显示例如下（`<claim-uid>`为ResourceClaim的UID，cdiIDs为CDI设备ID列表，格式为`<driverName>/<class>=<index>`，driverName取值为`npu.huawei.com`）：

   ```text
   PrepareResourceClaims is called, number of claims: 1
   devices prepared for claim <claim-uid>, count=1, cdiIDs=[npu.huawei.com/ascend910=0]
   Returning newly prepared devices for claim '<claim-uid>', count=1
   ```

4. 查看生成的CDI spec文件。

   DRA插件在Prepare阶段会为每个Claim生成CDI spec文件，位于节点上的`/var/run/cdi`目录。在Pod所在节点执行以下命令查看：

   ```bash
   ls /var/run/cdi/
   ```

   回显示例如下，存在以Claim UID命名的CDI spec文件：

   ```text
   <claim-uid>.json
   ```

   查看CDI spec文件内容：

   ```bash
   cat /var/run/cdi/<claim-uid>.json
   ```

   回显示例如下（以分配2颗NPU为例，实际内容随节点设备型号与驱动版本变化）：

   ```yaml
   cdiVersion: 0.8.0
   kind: [ascend.com/npu](https://ascend.com/npu)
   devices:
   - name: "0"
     containerEdits:
       deviceNodes:
       - path: /dev/davinci0
         hostPath: /dev/davinci0
         type: c
         major: 503
         fileMode: 8624
   - name: "1"
     containerEdits:
       deviceNodes:
       - path: /dev/davinci1
         hostPath: /dev/davinci1
         type: c
         major: 503
         fileMode: 8624
   containerEdits:
     env:
     - LD_LIBRARY_PATH=/usr/local/Ascend/driver/lib64/common:/usr/local/Ascend/driver/lib64/driver:/usr/lib64
     deviceNodes:
     - path: /dev/davinci_manager
       hostPath: /dev/davinci_manager
       type: c
       major: 504
       fileMode: 8624
     - path: /dev/hisi_hdc
       hostPath: /dev/hisi_hdc
       ...
     mounts:
     - hostPath: /usr/local/dcmi
       containerPath: /usr/local/dcmi
       options:
       - rbind
       - rprivate
       - ro
       type: bind
     - hostPath: /usr/local/bin/npu-smi
       containerPath: /usr/local/bin/npu-smi
       ...
   ```

   字段说明：
   - `cdiVersion`：CDI spec版本，固定为`0.8.0`。
   - `kind`：CDI driver标识，固定为`https://ascend.com/npu`。
   - `devices`：按设备粒度划分的编辑项，每个`name`对应一颗NPU的`/dev/davinci<id>`设备节点。
   - `containerEdits`（spec级）：所有设备共享的编辑项，包括共享设备节点（`davinci_manager`、`hisi_hdc`等）、`LD_LIBRARY_PATH`环境变量与驱动库/工具挂载。

5. 进入业务Pod容器，验证NPU设备已注入。

   ```bash
   kubectl exec -it <pod-name> -- ls /dev/davinci*
   ```

   回显示例如下，容器内可见NPU设备节点：

   ```text
   /dev/davinci0
   ```

   > [!NOTE]
   > NPU设备节点路径为`/dev/davinci<id>`，其中`<id>`为设备物理ID，与ResourceSlice中设备的`physicId`属性一致。设备的具体注入由CDI（Container Device Interface）规范完成，DRA插件负责生成CDI spec，容器运行时负责按CDI spec执行挂载。

### 删除任务并触发Unprepare<a name="section_unprepare"></a>

删除Deployment后，每个Pod被删除时kubelet会自动调用DRA插件的`UnprepareResourceClaims`回调，释放设备并清理CDI spec文件。

1. 删除Deployment。

   ```bash
   kubectl delete -f dra-test-deployment.yaml
   ```

   回显示例如下：

   ```text
   deployment.apps "dra-test" deleted
   ```

2. 验证Unprepare已完成。

   查看DRA组件日志，确认`UnprepareResourceClaims`被调用且设备已释放（以下日志与代码`plugin.go`中的`Infof`输出一致）：

   ```bash
   kubectl logs -n ascend-dra-driver <dra-pod-name> | grep "UnprepareResourceClaims\|devices unprepared"
   ```

   回显示例如下：

   ```text
   UnprepareResourceClaims is called, number of claims: 1
   devices unprepared for claim <claim-uid>
   ```

3. 确认CDI spec文件已清理。

   在原Pod所在节点执行以下命令，确认CDI spec文件已被删除：

   ```bash
   ls /var/run/cdi/
   ```

   回显中不再包含已删除Claim对应的spec文件。

### 清理资源<a name="section_cleanup"></a>

如需清理DRA组件及相关资源，执行以下命令：

1. 删除ResourceClaimTemplate。

   ```bash
   kubectl delete -f ascend-npu-claim-template.yaml
   ```

2. （可选）删除DRA组件。

   ```bash
   kubectl delete -f ascend-dra-driver.yaml
   ```

## 关键说明<a name="section_dra_notes"></a>

- **幂等性**：Prepare和Unprepare操作均支持幂等。对已Prepare的Claim再次Prepare会返回相同的分配结果；对已Unprepare的Claim再次Unprepare为空操作，不会报错。
- **Checkpoint机制**：DRA插件通过Checkpoint文件（`checkpoint.json`）持久化Claim与设备的分配关系。插件重启后从checkpoint恢复分配状态，确保已Prepare的Claim不受影响。
- **DeviceClass**：DRA组件部署时自动创建名为`npu.huawei.com`的DeviceClass，其选择器匹配driver为`npu.huawei.com`的所有设备。用户无需手动创建DeviceClass。
- **与Ascend Device Plugin的关系**：DRA和Device Plugin是两种不同的设备接入方式，互斥使用。使用DRA方式的节点无需部署Ascend Device Plugin。
