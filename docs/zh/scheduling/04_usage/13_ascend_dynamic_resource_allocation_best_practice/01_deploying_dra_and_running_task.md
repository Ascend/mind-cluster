# DRA机制下资源申请与释放<a name="ZH-CN_TOPIC_000000_dra_deploy_workflow"></a>

## 实现原理<a name="section_dra_principle"></a>

Ascend DRA基于Kubernetes动态资源分配（Dynamic Resource Allocation，DRA）机制管理NPU设备，其完整的生命周期请参见[使用前必读/生命周期](00_before_you_start.md#section_dra_life_cycle)。

申请整卡资源有两种方案，两者互斥选择其一：

- **方案一：通过ResourceClaimTemplate申请资源**。显式创建ResourceClaimTemplate，Pod通过`spec.resourceClaims`引用。适用于需要显式管理资源模板或需要与现有ResourceClaim流程保持一致的场景。该方案可申请未切分的整卡，也可通过替换`deviceClassName`申请静态vNPU或动态vNPU。
- **方案二：使用扩展资源名申请资源**。不创建ResourceClaimTemplate，Pod只在`resources.requests`与`resources.limits`中配置扩展资源名。适用于只需申请整卡、希望清单更简单的场景，要求K8s 1.36及以上版本。该方案只支持申请未切分的物理整卡，申请虚拟卡请使用方案一。

## 前提条件<a name="section_dra_prereq"></a>

- 已完成Volcano、Ascend DRA组件的安装部署，详细请参见[安装部署](../../03_installation_guide/02_installation/00_helm_installation.md)。Ascend DRA组件通过CDI机制向业务容器注入NPU设备，不安装也不依赖Ascend Docker Runtime。
- NPU节点已安装NPU驱动与固件，且设备运行正常。
- 使用[方案二](#section_scheme_extended_resource)时，K8s版本需为1.36及以上，`DynamicResourceAllocation`与`DRAExtendedResource`特性门控已开启（K8s 1.36及以上版本默认开启），且调度器为kube-scheduler。

## 方案一：通过ResourceClaimTemplate申请资源<a name="section_scheme_claim_template"></a>

用户通过ResourceClaimTemplate定义NPU资源请求，Pod通过引用该模板自动为每个副本创建独立的ResourceClaim。

### 配置YAML文件<a name="section_claim_template_config"></a>

1. 创建`ascend-npu-claim-template.yaml`文件：

    <pre codetype="yaml">
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
          - <strong>name: npu              # 请求名称，Pod中通过该名称引用</strong>
            <strong>exactly:</strong>
              <strong>deviceClassName: npu.huawei.com # 引用DRA组件部署时创建的DeviceClass</strong>
              <strong>count: 1                        # 请求的NPU数量，按需修改</strong>
    </pre>

    > [!NOTE]
    > `deviceClassName`取值为`npu.huawei.com`，该DeviceClass由`ascend-dra-driver.yaml`部署时自动创建，其选择器匹配driver为`npu.huawei.com`的未切分的物理整卡。申请虚拟卡时改为`static-vnpu.huawei.com`或`dynamic-vnpu.huawei.com`。如果只需申请整卡而不想创建模板，请参见[方案二](#section_scheme_extended_resource)。

2. 创建模板：

    ```bash
    kubectl apply -f ascend-npu-claim-template.yaml
    ```

    回显示例如下：

    ```text
    resourceclaimtemplate.resource.k8s.io/ascend-npu-template created
    ```

### 下发任务<a name="section_claim_template_apply"></a>

下发一个引用ResourceClaimTemplate的Deployment，每个Pod副本会自动创建独立的ResourceClaim。Pod调度到NPU节点后，kubelet会自动调用DRA插件完成Prepare。

1. 创建`dra-test-deployment.yaml`文件：

    <pre codetype="yaml">
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: dra-test
      namespace: default
      labels:
        app: dra-test
    spec:
      <strong>replicas: 1                               # 副本数，按需修改</strong>
      selector:
        matchLabels:
          app: dra-test
      template:
        metadata:
          labels:
            app: dra-test
        spec:
          <strong>resourceClaims:</strong>
          <strong>- name: npu                           # 与容器resources.claims中引用的name一致</strong>
            <strong>resourceClaimTemplateName: ascend-npu-template</strong>
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
              <strong>- name: npu                      # 引用上方resourceClaims中的name</strong>
    </pre>

2. 下发任务。

   ```bash
   kubectl apply -f dra-test-deployment.yaml
   ```

   回显示例如下：

   ```text
   deployment.apps/dra-test created
   ```

3. 查看Deployment状态，确认Pod进入Running状态。

   ```bash
   kubectl get pod -l app=dra-test -o wide
   ```

   回显示例如下：

   ```text
   NAME                       READY   STATUS    RESTARTS   AGE   IP            NODE          NOMINATED NODE   READINESS GATES
   dra-test-<random-suffix>   1/1     Running   0          30s   10.244.1.20   node-npu-01   <none>           <none>
   ```

4. 验证Prepare已完成。

   查看DRA组件日志，确认`PrepareResourceClaims`被调用且设备已分配（以下日志与代码`plugin.go`和`state.go`中的`Infof`输出一致）：

   ```bash
   kubectl logs -n kube-system <dra-pod-name> | grep "PrepareResourceClaims\|devices prepared\|Returning newly prepared"
   ```

   回显示例如下：

   ```text
   PrepareResourceClaims is called, number of claims: 1
   devices prepared for claim <claim-uid>, count=1, cdiIDs=[npu.huawei.com/ascend910=0]
   Returning newly prepared devices for claim '<claim-uid>', count=1
   ```

   其中，`<claim-uid>`为ResourceClaim的UID，cdiIDs为CDI设备ID列表，格式为`<driverName>/<class>=<index>`，driverName取值为`npu.huawei.com`。

5. 查看生成的CDI（Container Device Interface） spec文件。

   DRA插件在Prepare阶段会为每个Claim生成CDI spec文件，位于节点上的`/var/run/cdi`目录。

   1. 在Pod所在节点查看CDI spec文件：

      ```bash
      ls /var/run/cdi/
      ```

      回显示例如下，存在以Claim UID命名的CDI spec文件：

      ```text
      <claim-uid>.json
      ```

   2. 查看CDI spec文件内容：

      ```bash
      cat /var/run/cdi/<claim-uid>.json
      ```

      回显示例如下（以分配2颗NPU为例，实际内容随节点设备型号与驱动版本变化）：

      ```yaml
      cdiVersion: 0.8.0
      kind: ascend.com/npu
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

6. 进入业务Pod容器，验证NPU设备已注入。

   ```bash
   kubectl exec -it <pod-name> -- ls /dev/davinci*
   ```

   回显示例如下，容器内可见NPU设备节点：

   ```text
   /dev/davinci0
   ```

   > [!NOTE]
   > NPU设备节点路径为`/dev/davinci<id>`，其中`<id>`为设备物理ID，与ResourceSlice中设备的`physicId`属性一致。设备的具体注入由CDI规范完成，DRA插件负责生成CDI spec，容器运行时负责按CDI spec执行挂载。

### 删除任务<a name="section_claim_template_delete"></a>

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
   kubectl logs -n kube-system <dra-pod-name> | grep "UnprepareResourceClaims\|devices unprepared"
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

4. 删除ResourceClaimTemplate。

   任务不再需要时，删除模板：

   ```bash
   kubectl delete -f ascend-npu-claim-template.yaml
   ```

## 方案二：使用扩展资源名申请资源<a name="section_scheme_extended_resource"></a>

不创建ResourceClaim，直接在Pod的`resources.requests`与`resources.limits`中配置扩展资源名申请未切分的物理整卡。kube-scheduler按DeviceClass的`extendedResourceName`匹配设备，自动合成ResourceClaim并完成分配，分配结果记录在`pod.status.extendedResourceClaimStatus`中。

该方式只支持申请未切分的物理整卡。`resources.requests`与`resources.limits`必须同时配置且数值相等，取值`1`表示申请一张未切分的物理整卡。`static-vnpu.huawei.com`与`dynamic-vnpu.huawei.com`没有`extendedResourceName`，虚拟卡不能使用这种方式申请，请改用[方案一](#section_scheme_claim_template)。

### 配置YAML文件<a name="section_extended_config"></a>

1. 确认资源名与芯片代际的对应关系。资源名必须等于某个DeviceClass的`extendedResourceName`：

    **表 1**  扩展资源名与DeviceClass对应关系

    <a name="table_dra_extended_resource"></a>

    | 扩展资源名 | DeviceClass | 适用芯片 |
    |--|--|--|
    | `huawei.com/Ascend910` | ascend910.ext.huawei.com | Ascend910系列，`chipName`以`910`或`Ascend910`开头 |
    | `huawei.com/npu` | npu.ext.huawei.com | Ascend950系列，`chipName`以`Ascend950`开头 |

    > [!NOTE]
    > 同一节点不要同时部署Ascend Device Plugin。Device Plugin上报的扩展资源名与本表相同，两者同时存在会对同一张卡重复计数，并触发资源名冲突。

2. 创建`dra-extended-npu.yaml`文件，业务Pod只需配置`requests`与`limits`，不配置`spec.resourceClaims`：

    <pre codetype="yaml">
    apiVersion: v1
    kind: Pod
    metadata:
      name: dra-extended-npu
      namespace: default
    spec:
      containers:
      - name: test
        image: ubuntu:22.04
        imagePullPolicy: IfNotPresent
        command:
        - /bin/bash
        - -c
        - sleep inf
        resources:
          <strong>requests:</strong>
            <strong>huawei.com/Ascend910: 1     # 申请1张未切分的物理整卡，按需修改</strong>
          <strong>limits:</strong>
            <strong>huawei.com/Ascend910: 1     # 必须与requests相等</strong>
    </pre>

    芯片为Ascend950系列时，将两处`huawei.com/Ascend910`同时改为`huawei.com/npu`。

### 下发任务<a name="section_extended_apply"></a>

1. 下发任务。

   ```bash
   kubectl apply -f dra-extended-npu.yaml
   ```

2. 验证Pod进入Running状态，且调度器已完成资源分配。

   ```bash
   kubectl get pod dra-extended-npu
   kubectl get pod dra-extended-npu -o jsonpath='{.status.extendedResourceClaimStatus.resourceClaimName}'
   ```

   回显示例如下，第一条`STATUS`为`Running`表示容器已启动，第二条输出的`resourceClaimName`非空表示调度器已完成资源分配：

   ```text
   NAME                READY   STATUS    RESTARTS   AGE
   dra-extended-npu    1/1     Running   0          25s

   dra-extended-npu-abcde
   ```

3. 进入业务Pod容器，验证NPU设备已注入。

   ```bash
   kubectl exec -it dra-extended-npu -- ls /dev/davinci*
   ```

   回显示例如下：

   ```text
   /dev/davinci0
   ```

### 删除任务<a name="section_extended_delete"></a>

删除Pod后，kubelet会自动调用DRA插件的`UnprepareResourceClaims`回调，释放设备并清理CDI spec文件。与本方案合成的ResourceClaim由集群随Pod一并清理，无需手工删除。

1. 删除Pod。

   ```bash
   kubectl delete -f dra-extended-npu.yaml
   ```

   回显示例如下：

   ```text
   pod "dra-extended-npu" deleted
   ```

2. 验证设备已释放。

   ```bash
   kubectl logs -n ascend-dra-driver <dra-pod-name> | grep "UnprepareResourceClaims\|devices unprepared"
   ```

   回显示例如下：

   ```text
   UnprepareResourceClaims is called, number of claims: 1
   devices unprepared for claim <claim-uid>
   ```

3. 确认CDI spec文件已清理。

   在原Pod所在节点执行以下命令：

   ```bash
   ls /var/run/cdi/
   ```

   回显中不再包含已删除Claim对应的spec文件。

## 关键说明<a name="section_dra_notes"></a>

- **幂等性**：Prepare和Unprepare操作均支持幂等。对已Prepare的Claim再次Prepare会返回相同的分配结果；对已Unprepare的Claim再次Unprepare为空操作，不会报错。
- **Checkpoint机制**：DRA插件通过Checkpoint文件（`checkpoint.json`）持久化Claim与设备的分配关系。插件重启后从Checkpoint恢复分配状态，确保已Prepare的Claim不受影响。
- **清理部署**：删除业务侧资源后，如需卸载组件，执行`kubectl delete -f ascend-dra-driver.yaml`。
- **DeviceClass**：DRA组件部署时自动创建名为`npu.huawei.com`的DeviceClass，其选择器匹配driver为`npu.huawei.com`的所有设备。用户无需手动创建DeviceClass。
- **与Ascend Device Plugin的关系**：DRA和Device Plugin是两种不同的设备接入方式，两者功能在同一节点上互斥，无需同时部署。
