# RFC: 推理工作负载特性说明

## 1. 概述
推理工作负载特性是MindCluster针对大模型推理服务在昇腾硬件集群上提供的K8S工作负载能力，支持推理服务的单机、多机、PD分离等场景的任务管理能力。本特性新增组件Infer Operator,其是一个 Kubernetes Operator，用于部署和管理多角色合作的推理任务。它定义了三种自定义资源（CRD）：InferServiceSet、InferService 和 InstanceSet，并实现了相应的控制器来调谐这些资源的实例状态。通过 Infer Operator，用户可以轻松地在 Kubernetes 集群上部署、扩展和管理复杂的推理服务。

## 2. 提议方案
1、基于k8s controller manager机制新增组件Infer Operator管理推理任务工作负载。
2、新增三层CRD，一层CRD描述推理服务，二层CRD描述推理实例集，三层CRD描述推理实例。
3、三层CRD支持替换为K8S原生工作负载，如Deployment、StatefulSet等，同时支持用户扩展自定义的CRD类型。

## 3. 术语定义

| 术语 | 解释 |
|------|------|
| CRD | Kubernetes 自定义资源定义（Custom Resource Definition） |
| Operator | 基于 Kubernetes API 构建的软件扩展，用于管理自定义资源 |
| InferServiceSet | 管理一组 InferService 实例的顶层资源 |
| InferService | 管理多角色推理服务的中间层资源 |
| InstanceSet | 管理具体工作负载实例的底层资源 |
| PodGroup | Volcano 调度器的资源，用于实现 gang 调度 |
| Deployment | Kubernetes 无状态工作负载 |
| StatefulSet | Kubernetes 有状态工作负载 |

## 4. 介绍

### 4.1 背景

在小模型时代，一个模型通常不会超过单台虚拟机/物理机部署，因此通常将实例和Pod当作相似甚至对等的概念，此时使用K8S原生的无状态工作负载Deployment基本上能够解决大多数服务部署能力，Deployment能够提供Pod粒度的资源调度、滚动升级、故障恢复和扩缩容等必要的能力。但随着大模型时代的到来，模型参数规模逐渐庞大，单台虚拟机/物理机的资源上限无法承载模型对资源的诉求，再加上层出不穷的部署形态的优化创新，如PD分离部署、AF部署、大小模型混部的投机推理等，实例这个概念早已无法局限在单个Pod层级，单实例需要对多个Pod的分布式跨机通信进行管理，使得单实例内的多Pod以及实例间需要各种协同。因此，对于一个推理服务，实例下的实现内容或者载体范围出现了变更，它可能仍然是单个Pod（普通的单机服务），也可能是一组不相同的，各司其职的Pod（PD分离部署）。MindCluster此前通过ascend-operator提供的ascend job CRD部署推理任务，单acjob映射一个推理实例的方式只能解决部分场景下的协同问题，无法很好的支持越来越复杂的场景，因此，MindCluster需要针对推理场景提供一套完整的推理服务工作负载。

### 4.2 目标

- 简化多角色推理服务的部署和管理
- 支持水平扩展和缩容
- 提供统一的状态监控和管理界面
- 支持 gang 调度，确保推理服务的所有组件同时部署
- 兼容现有的 Kubernetes 生态系统

## 5. 架构设计

### 5.1 组件层次结构

Infer Operator 采用三层架构设计：

1. **InferServiceSet**：顶层资源，管理一组相同配置的 InferService 实例
2. **InferService**：中间层资源，管理多角色推理服务，每个角色对应一个 InstanceSet
3. **InstanceSet**：底层资源，管理具体的工作负载（Deployment 或 StatefulSet）

### 5.2 控制器结构

Infer Operator 包含三个主要控制器：

1. **InferServiceSetController**：管理 InferServiceSet 资源，负责创建、更新和删除 InferService 实例
2. **InferServiceController**：管理 InferService 资源，负责创建、更新和删除 InstanceSet 实例
3. **InstanceSetController**：管理 InstanceSet 资源，负责创建、更新和删除具体的工作负载和服务

### 5.3 工作负载处理

InstanceSetController 通过 WorkLoadReconciler 接口处理不同类型的工作负载：

- DeploymentHandler：处理无状态工作负载
- StatefulSetHandler：处理有状态工作负载

### 5.4 依赖关系

```
InferServiceSet
    └─── InferService (1..N)
            └─── InstanceSet (1..N)
                    ├─── Service (0..N)
                    └─── Workload (Deployment/StatefulSet)
                            └─── Pod (1..N)
```

## 6. API 规范

### 6.1 核心 CRD

#### 6.1.1 InferServiceSet

```yaml
apiVersion: mindcluster.huawei.com/v1
kind: InferServiceSet
metadata:
  name: <iss-name>
  namespace: <namespace>
spec:
  replicas: <number>
  template:
    roles:
      - name: <role-name>
        replicas: <number>
        workload:
          kind: <Deployment/StatefulSet>
          apiVersion: apps/v1
        services:
          - name: <service-name>
            spec:
              <service-spec>
        spec:
          <workload-spec>
```

#### 6.1.2 InferService

```yaml
apiVersion: mindcluster.huawei.com/v1
kind: InferService
metadata:
  name: <is-name>
  namespace: <namespace>
spec:
  roles:
    - name: <role-name>
      replicas: <number>
      workload:
        kind: <Deployment/StatefulSet>
        apiVersion: apps/v1
      services:
        - name: <service-name>
          spec:
            <service-spec>
      spec:
        <workload-spec>
```

#### 6.1.3 InstanceSet

```yaml
apiVersion: mindcluster.huawei.com/v1
kind: InstanceSet
metadata:
  name: <inset-name>
  namespace: <namespace>
spec:
  name: <role-name>
  replicas: <number>
  services:
    - name: <service-name>
      spec:
        <service-spec>
  workload:
    kind: <Deployment/StatefulSet>
    apiVersion: apps/v1
  metadata:
    labels:
      <labels>
    annotations:
      <annotations>
  spec:
    <workload-spec>
```

### 6.2 标签和注解

| 标签/注解 | 用途 | 示例值 |
|-----------|------|--------|
| infer.huawei.com/ascend-infer-operator | 标识由 Infer Operator 管理的资源 | "infer-cm" |
| infer.huawei.com/inferserviceset-name | 关联 InferServiceSet | "my-iss" |
| infer.huawei.com/inferservice-name | 关联 InferService | "my-is" |
| infer.huawei.com/instanceset-name | 关联 InstanceSet | "my-inset" |
| infer.huawei.com/inferservice-index | InferService 索引 | "0" |
| infer.huawei.com/gang-schedule | 是否启用 gang 调度 | "true" |

## 7. 实现细节

### 7.1 控制器逻辑

#### 7.1.1 InferServiceSetController

1. **Reconcile 循环**：
    - 获取 InferServiceSet 资源
    - 验证资源配置
    - 列出所有关联的 InferService
    - 计算需要创建、更新或删除的 InferService
    - 执行相应的操作
    - 更新 InferServiceSet 状态

2. **关键功能**：
    - 水平扩展/缩容 InferService 实例
    - 保持 InferService 实例的配置一致性
    - 监控 InferService 实例状态

#### 7.1.2 InferServiceController

1. **Reconcile 循环**：
    - 获取 InferService 资源
    - 验证角色配置
    - 列出所有关联的 InstanceSet
    - 计算需要创建、更新或删除的 InstanceSet
    - 执行相应的操作
    - 更新 InferService 状态

2. **关键功能**：
    - 管理多角色推理服务
    - 确保所有角色的 InstanceSet 配置一致
    - 验证角色名称的唯一性和格式

#### 7.1.3 InstanceSetController

1. **Reconcile 循环**：
    - 获取 InstanceSet 资源
    - 验证资源配置
    - 处理服务创建/更新
    - 处理工作负载创建/更新
    - 更新 InstanceSet 状态

2. **关键功能**：
    - 管理具体的工作负载实例
    - 处理 NodePort 服务的端口冲突
    - 支持 gang 调度

### 7.2 工作负载管理

WorkLoadReconciler 支持两种类型的工作负载：

1. **Deployment**：用于无状态推理服务
2. **StatefulSet**：用于有状态推理服务，如需要持久化存储的场景

### 7.3 Gang 调度支持

当启用 gang 调度时，InstanceSetController 会：
1. 创建 Volcano PodGroup 资源
2. 配置最小成员数为工作负载的副本数
3. 确保所有 Pod 同时调度和运行

## 8. 部署和配置

### 8.1 部署方式

Infer Operator 可以通过以下方式部署：

1. **YAML 部署**：使用提供的 `infer-operator.yaml` 文件部署
2. **Helm 部署**：（计划中）

### 8.2 部署命令

```bash
kubectl apply -f infer-operator.yaml
```

### 8.3 配置参数

| 参数 | 描述 | 默认值 |
|------|------|--------|
| --version | 查询程序版本 | - |
| --logLevel | 日志级别 | 0（info） |
| --maxAge | 日志文件最大保存天数 | 7 |
| --isCompress | 是否压缩日志文件 | false |
| --logFile | 日志文件路径 | /var/log/mindx-dl/infer-operator/infer-operator.log |
| --maxBackups | 日志文件最大备份数 | 5 |

### 8.4 RBAC 配置

Infer Operator 需要以下权限：

- 管理自定义资源（InferServiceSet、InferService、InstanceSet）
- 管理工作负载（Deployment、StatefulSet）
- 管理服务（Service）
- 管理配置映射（ConfigMap）
- 管理 PodGroup（可选）

## 9. 使用示例

### 9.1 创建简单的推理服务

```yaml
apiVersion: mindcluster.huawei.com/v1
kind: InferService
metadata:
  name: simple-infer
  namespace: mindx-dl
spec:
  roles:
  - name: worker
    replicas: 3
    workload:
      kind: Deployment
      apiVersion: apps/v1
    services:
    - name: inference
      spec:
        type: NodePort
        ports:
        - port: 8080
          nodePort: 30080
    spec:
      template:
        spec:
          containers:
          - name: inference
            image: inference-image:latest
            ports:
            - containerPort: 8080
```

### 9.2 创建多角色推理服务

```yaml
apiVersion: mindcluster.huawei.com/v1
kind: InferService
metadata:
  name: multi-role-infer
  namespace: mindx-dl
spec:
  roles:
  - name: frontend
    replicas: 2
    workload:
      kind: Deployment
      apiVersion: apps/v1
    services:
    - name: frontend
      spec:
        type: NodePort
        ports:
        - port: 80
          nodePort: 30000
    spec:
      template:
        spec:
          containers:
          - name: frontend
            image: frontend-image:latest
            ports:
            - containerPort: 80
  - name: backend
    replicas: 4
    workload:
      kind: StatefulSet
      apiVersion: apps/v1
    services:
    - name: backend
      spec:
        clusterIP: None
        ports:
        - port: 8080
    spec:
      template:
        spec:
          containers:
          - name: backend
            image: backend-image:latest
            ports:
            - containerPort: 8080
```

### 9.3 创建 InferServiceSet

```yaml
apiVersion: mindcluster.huawei.com/v1
kind: InferServiceSet
metadata:
  name: infer-cluster
  namespace: mindx-dl
spec:
  replicas: 3
  template:
    roles:
    - name: worker
      replicas: 2
      workload:
        kind: Deployment
        apiVersion: apps/v1
      services:
      - name: inference
        spec:
          type: ClusterIP
          ports:
          - port: 8080
      spec:
        template:
          spec:
            containers:
            - name: inference
              image: inference-image:latest
              ports:
              - containerPort: 8080
```

## 10. 监控和维护

### 10.1 日志

Infer Operator 会将日志输出到配置的日志文件中。日志级别可以通过 `--logLevel` 参数调整：

- -1: debug
- 0: info
- 1: warning
- 2: error
- 3: critical

### 10.2 状态监控

每个资源都有状态字段，可以通过以下命令查看：

```bash
kubectl get inferserviceset <name> -n <namespace> -o yaml
kubectl get inferservice <name> -n <namespace> -o yaml
kubectl get instanceset <name> -n <namespace> -o yaml
```

### 10.3 常见问题排查

1. **资源创建失败**：
    - 检查资源配置是否正确
    - 查看 Infer Operator 日志
    - 检查 RBAC 权限

2. **Pod 调度失败**：
    - 检查集群资源是否充足
    - 如果启用了 gang 调度，检查 Volcano 是否正常运行

3. **服务无法访问**：
    - 检查 Service 配置是否正确
    - 检查 Pod 是否正常运行
    - 检查网络策略

## 11. 安全考虑

### 11.1 RBAC 权限

Infer Operator 遵循最小权限原则，只请求必要的权限。

### 11.2 容器安全

- 使用非 root 用户运行容器
- 限制容器权限
- 使用只读根文件系统
- 配置适当的 seccomp 策略

### 11.3 网络安全

- 建议使用网络策略限制访问
- 对于 NodePort 服务，建议配置防火墙规则

## 12. 未来计划

1. 支持更多类型的工作负载
2. 添加 Helm Chart 支持
3. 增强监控和告警功能
4. 支持自动扩缩容
5. 优化性能和稳定性