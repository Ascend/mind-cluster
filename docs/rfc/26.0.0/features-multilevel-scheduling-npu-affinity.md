# RFC: 多级网络亲和性调度特性说明

## 概述

多级网络亲和性调度（Multilevel Network Affinity Scheduling）是 Volcano 调度器中 `ascend-volcano-plugin` 插件提供的一个高级调度特性。该特性基于通用抽象网络拓扑模型，在标准 Volcano 网络拓扑感知调度的基础上，扩展支持了对 Ascend NPU 芯片的多级亲和性调度能力，能够根据实际物理网络拓扑结构智能分配分布式训练/推理任务，最大化网络带宽利用率，降低跨网络通信延迟。

## 背景与动机

在大规模分布式AI训练场景中，NPU设备之间的通信性能对整体训练效率至关重要。现代数据中心的网络拓扑通常呈现多层级结构，特别是AI集群通常采用叶脊网络架构，存在多个层级的交换机：

- **物理层级**：NPU芯片 → 服务器 → 机架 → leaf交换机 → spine交换机 → 数据中心
- **网络带宽**：片内互联 > 服务器内部 > 同机架 > 同leaf交换机 > 同spine交换机 > 跨交换机

不感知网络拓扑的调度可能导致任务分布不合理，造成：

- 跨网络通信增加，通信延迟升高
- 网络拥塞，整体训练吞吐量下降
- 资源碎片化，降低集群利用率

多级网络亲和性调度通过将任务按照网络拓扑层级进行分组分配，保证了亲和性任务尽可能部署在网络延迟更低、带宽更高的区域。

## 术语定义

| 术语                               | 说明                             |
| -------------------------------- | ------------------------------ |
| **资源树 (Resource Tree)**          | 根据实际物理网络拓扑构建的资源层级树，每个节点代表一个网络域 |
| **任务树 (Task Tree)**              | 根据用户作业多级亲和性要求构建的任务层级树          |
| **层级 (Level)**                   | 网络拓扑中的一个抽象层次，对应物理网络的一个组织单元     |
| **资源预留 (Resource Reservation)**  | 为特定层级预保留一定数量的资源，保障关键任务部署       |
| **碎片优化 (Fragment Optimization)** | 通过碎片得分评估资源利用率，选择碎片最少的分配方案      |

## 网络层级抽象定义

参考 Volcano 社区网络拓扑感知调度的抽象模型，本特性支持三级以上的网络层级抽象：

### 层级类型定义

```go
type LevelType string

const (
    // LevelTypeTree 根树节点，代表整个资源树
    LevelTypeTree LevelType = "tree"
    // LevelTypeMiddle 中间网络节点，代表机架、交换机等中间层级
    LevelTypeMiddle LevelType = "middle"
    // LevelTypeNode 物理节点，代表实际运行Pod的K8s节点
    LevelTypeNode LevelType = "node"
)
```

### 典型网络拓扑示例

按照本系统编号规则，`levelN` 的 N 越小越靠近物理节点：

| 层级编号   | 层级类型   | 示例物理含义      | 分组范围 | 通信开销 |
| ------ | ------ | ----------- | ---- | ---- |
| level1 | Middle | 机架 / 服务器组   | 最小   | 最低   |
| level2 | Middle | 交换机         | 中等   | 中等   |
| level3 | Middle | 汇聚交换机 / 叶脊网络上层 | 最大   | 最高   |

> **说明**：系统会自动添加根拓扑树层级和最终物理节点层级，用户只需要配置中间层级即可。调度时会根据节点标签自动匹配集群内已配置的拓扑树，用户作业无需指定拓扑树名称。

### 节点标签映射

每个层级通过Kubernetes节点标签来标识拓扑位置：

```yaml
# 示例节点标签配置
node.labels.huawei.com/topotree: "default"      # 拓扑树标识
node.labels.switch-id: "switch-001"            # 交换机层级标签
node.labels.rack-id: "rack-001"                # 机架层级标签
```

## 架构设计

### 整体架构

多级调度策略位于 `pkg/scheduler/plugins/ascend-volcano-plugin/internal/npu/policy/multilevelscheduling` 目录，核心模块划分如下：

<div style="font-family: monospace;">
<table style="border-collapse: collapse; border: none; margin: 20px 0;">
  <tr>
    <td colspan="3" style="border: 1px solid #ccc; padding: 10px; text-align: center; background-color: #f8f8f8;">
      <strong>MultilevelHandler (frame.go)</strong>
      <br>• 对外接口实现 &nbsp;• 配置验证 &nbsp;• 节点打分
    </td>
  </tr>
  <tr>
    <td colspan="3" style="text-align: center; padding: 5px;">▼</td>
  </tr>
  <tr>
    <td style="border: 1px solid #ccc; padding: 15px; background-color: #f8f8f8; vertical-align: top;">
      <strong>Scheduling<br>(scheduling.go)</strong>
      <br><br>
      • 调度树创建<br>
      • 主调度逻辑<br>
      • 任务分配
    </td>
    <td style="width: 50px;"></td>
    <td rowspan="2" style="border: 1px solid #ccc; padding: 15px; background-color: #f8f8f8; vertical-align: top;">
      <strong>Common/Util</strong>
      <br><br>
      • 类型定义<br>
      • 树操作工具
    </td>
  </tr>
  <tr>
    <td style="border: 1px solid #ccc; padding: 15px; background-color: #f8f8f8; vertical-align: top;">
      <strong>Rescheduling<br>(rescheduling.go)</strong>
      <br><br>
      • 故障检测<br>
      • 重调度逻辑
    </td>
    <td></td>
  </tr>
</table>
</div>

### 核心数据结构

#### 1. 多级调度处理器

```go
type MultilevelHandler struct {
    base.NPUHandler
    taskLevels []util.TaskTreeLevel  // 任务层级配置
}
```

#### 2. 调度树结构

```go
// 完整调度树
type schedulingTree struct {
    root   *schedulingTreeNode
    levels []*schedulingTreeLevel
}

// 调度树层级
type schedulingTreeLevel struct {
    taskLevel     util.TaskTreeLevel
    resourceLevel util.ResourceTreeLevel
    nodes         []*schedulingTreeNode
}

// 调度树节点
type schedulingTreeNode struct {
    depth    int
    node     *util.ResourceNode
    parent   *schedulingTreeNode
    children map[string]*schedulingTreeNode

    // 资源碎片相关字段
    hasTraversed         bool
    allocatableTaskCount int
    fragmentScore        int
    freeSubTasks         []*util.TaskNode

    // 资源预留相关字段
    isReserved                    bool
    hasSufficientReservedResource bool
}
```

#### 3. 资源树与任务树

```go
// 资源树 - 描述集群物理网络拓扑
type ResourceTree struct {
    *ResourceNode
    Name   string
    Levels []ResourceTreeLevel
}

// 资源节点 - 代表一个网络域
type ResourceNode struct {
    Name     string
    Parent   *ResourceNode
    Children map[string]*ResourceNode
}

// 资源树层级配置
type ResourceTreeLevel struct {
    Label        string    `json:"label,omitempty"`  // 对应节点标签名称
    ReservedNode int       `json:"reservedNode,omitempty"` // 预留节点数
    Type         LevelType `json:"-"`
}

// 任务树 - 描述用户作业的多级亲和性要求
type TaskTree struct {
    *TaskNode
    FragmentScore  int
    ResourceLevels []ResourceTreeLevel
    Levels         []TaskTreeLevel
}

// 任务树层级配置
type TaskTreeLevel struct {
    Name    string  // 层级名称
    ReqNode int     // 该层级需要的节点数
}
```

## 配置方式

### 1. 集群层面插件启动配置

要启用多级调度特性，需要在 Volcano scheduler 的 ConfigMap 中配置 `ascend-volcano-plugin` 插件的启动参数。资源层级拓扑通过 `resource-level-config` 字段配置，格式为JSON。

完整配置示例：

```yaml
# 完整的 Volcano scheduler ConfigMap 配置示例
apiVersion: v1
kind: ConfigMap
metadata:
  name: volcano-scheduler-configmap
  namespace: volcano-system
data:
  # 调度器主配置文件
  scheduler.conf: |
    actions: "enqueue, allocate, backfill, reclaim, predicates, prioritize"
    tiers:
    - plugins:
      - name: priority
      - name: gang
      - name: conformance
      - name: ascend-volcano-plugin  # 启用 ascend-volcano-plugin 插件
    - plugins:
      - name: npu-resource-score
      - name: drf
      - name: predicates
      - name: sanity
      - name: nodeorder
      - name: bind

  # 插件初始化参数，必须配置此字段
  init-params: |
    # 资源层级拓扑配置，JSON格式
    # 格式: { "拓扑树名称": { "level1": { "label": "标签名", "reservedNode": 预留数 }, "level2": ... } }
    resource-level-config: |
      {
        "default": {
          "level1": {
            "label": "switch-id",
            "reservedNode": 0
          },
          "level2": {
            "label": "rack-id",
            "reservedNode": 0
          }
        }
      }
```

**启动参数字段说明**：

| 字段                                                          | 是否必需   | 说明                              |
| ----------------------------------------------------------- | ------ | ------------------------------- |
| `resource-level-config`                                     | **必需** | 资源层级拓扑配置，JSON格式。键是拓扑树名称，值是各层级配置 |
| `resource-level-config[<topo-name>][<levelN>].label`        | **必需** | 第N层级对应的节点标签名称，用于标识节点在该层级的ID     |
| `resource-level-config[<topo-name>][<levelN>].reservedNode` | 可选     | 该层级预留的节点数量，默认为0表示不预留            |

**配置解析规则**：

- 系统自动从最高层levelN向level1逆序构建拓扑树
- 自动在末尾添加物理节点层，不需要用户配置
- 层级编号必须从1开始连续，不能跳号
- **资源预留限制**：目前预留节点配置仅在 `level1` 层级生效，只有底层物理组可以配置预留节点

### 2. 作业层面多级亲和性配置

通过 PodGroup 或 Pod 的 Annotations 字段开启多级调度并配置多级亲和性：

```yaml
apiVersion: scheduling.volcano.sh/v1beta1
kind: PodGroup
metadata:
  name: distributed-training-job
  annotations:
    # 必需：指定使用多级调度策略
    huawei.com/schedule_policy: "multilevel"
    # 必需：多级亲和性配置，格式：level1=N,level2=N,...
    huawei.com/affinity-config: "level1=2,level2=4"
spec:
  minMembers: 8
  queue: default
```

**配置格式说明**：

- `huawei.com/schedule_policy`: **必须指定为** **`multilevel`** 才能启用多级调度策略
- `huawei.com/affinity-config`: 使用 `levelN=G` 格式，多个层级用逗号分隔
  - `levelN`: N必须从1开始连续编号，不能跳号，**数字越小层级越靠近物理节点，通信开销越小**；数字越大层级越顶层，网络范围越大
  - `G`: 该层级每个**分组**包含的节点数（pod数）

**亲和性保障语义**：

- 对于 `levelN=G`，表示**同一个levelN分组内的所有pod一定被分配到具有相同层级标签值的节点上**
- N越小 → 越靠近物理节点 → 分组范围越小 → 节点间通信开销越低
- N越大 → 越顶层网络 → 分组范围越大 → 包含更多节点

**层级顺序示例**：

| levelN | 层级位置 | 典型物理含义 | 分组范围 | 通信开销 |
| ------ | ---- | ------ | ---- | ---- |
| level1 | 最底层  | 机架     | 最小   | 最低   |
| level2 | 中层   | 交换机    | 中等   | 中等   |
| level3 | 最顶层  | 数据中心分区 | 最大   | 最高   |

**配置约束**：

- 多级调度要求每个任务占用整张NPU卡，每个节点的NPU数必须被任务数整除（满卡分配）
- 总任务数（minMembers）必须能被每个层级的分组大小G整除
- 层级编号必须连续从1开始，不能跳号

**示例配置说明（8 pod任务）**：

```
总任务数 minMembers = 8
affinity-config: "level1=2,level2=4"

含义解析：
- level1=2 → 顶层level1每个分组包含2个节点。总共分成 8/2 = 4 个level1分组
- level2=4 → 第二层level2每个分组包含4个节点。总共分成 8/4 = 2 个level2分组

亲和性保障：
- 同一个level1分组内的2个节点 → 保证具有相同的level1标签值（如同一个交换机）
- 同一个level2分组内的4个节点 → 保证具有相同的level2标签值（如同一个机架）
```

### 3. 节点拓扑标签配置

每个Kubernetes节点需要标注拓扑位置信息：

```yaml
kind: Node
metadata:
  labels:
    # 标识该节点属于哪个拓扑树
    huawei.com/topotree: "default"
    # 对应各层级的标签，必须与 resource-level-config 中配置的label一致
    switch-id: "switch-01"
    rack-id: "rack-01-01"
```

## 调度流程

### 1. 初始化与配置验证阶段

<div style="font-family: monospace; background-color: #f8f8f8; border: 1px solid #ccc; padding: 15px; border-radius: 5px; margin: 10px 0;">
<table style="border-collapse: collapse; border: none; width: 100%;">
  <tr>
    <td style="padding: 8px;"><strong>1. ValidNPUJob() 验证作业配置</strong></td>
  </tr>
  <tr>
    <td style="padding-left: 25px;">├─ 检查每个任务NPU数量要求</td>
  </tr>
  <tr>
    <td style="padding-left: 25px;">└─ 调用 checkLevels() 验证多级亲和性配置</td>
  </tr>
  <tr>
    <td style="padding-left: 50px;">├─ 检查层级编号连续性</td>
  </tr>
  <tr>
    <td style="padding-left: 50px;">├─ 检查层级节点数整除关系</td>
  </tr>
  <tr>
    <td style="padding-left: 50px;">└─ 构建 taskLevels 配置</td>
  </tr>
</table>
</div>

### 2. 节点过滤阶段

<div style="font-family: monospace; background-color: #f8f8f8; border: 1px solid #ccc; padding: 15px; border-radius: 5px; margin: 10px 0;">
<table style="border-collapse: collapse; border: none; width: 100%;">
  <tr>
    <td colspan="2"><strong>CheckNodeNPUByTask() 检查节点是否满足要求</strong></td>
  </tr>
  <tr>
    <td style="padding-left: 25px;">1.</td>
    <td style="padding-left: 10px;">获取节点拓扑树标签</td>
  </tr>
  <tr>
    <td style="padding-left: 25px;">2.</td>
    <td style="padding-left: 10px;">检查集群是否配置了该拓扑树</td>
  </tr>
  <tr>
    <td style="padding-left: 25px;">3.</td>
    <td style="padding-left: 10px;">检查节点是否包含所有必需的层级标签</td>
  </tr>
  <tr>
    <td style="padding-left: 25px;">4.</td>
    <td style="padding-left: 10px;">验证节点上可用NPU数量满足任务需求</td>
  </tr>
</table>
</div>

### 3. 打分与节点选择阶段

<div style="font-family: monospace; background-color: #f8f8f8; border: 1px solid #ccc; padding: 15px; border-radius: 5px; margin: 10px 0;">
<table style="border-collapse: collapse; border: none; width: 100%;">
  <tr>
    <td colspan="3"><strong>ScoreBestNPUNodes() 为候选节点打分</strong></td>
  </tr>
  <tr>
    <td colspan="3"><br><strong>如果作业未完成节点选择：</strong></td>
  </tr>
  <tr>
    <td style="padding-left: 25px;">1.</td>
    <td colspan="2">收集所有健康NPU节点</td>
  </tr>
  <tr>
    <td style="padding-left: 25px;">2.</td>
    <td colspan="2">根据拓扑树标签构建资源树 ResourceTree(s)</td>
  </tr>
  <tr>
    <td style="padding-left: 25px;">3.</td>
    <td colspan="2">对每个资源树尝试调度：</td>
  </tr>
  <tr>
    <td></td>
    <td style="padding-left: 25px;">createSchedulingTree() 创建调度树</td>
  </tr>
  <tr>
    <td></td>
    <td style="padding-left: 50px;">├─ 根据层级结构逐层构建节点</td>
  </tr>
  <tr>
    <td></td>
    <td style="padding-left: 50px;">├─ initNode() 初始化节点</td>
  </tr>
  <tr>
    <td></td>
    <td style="padding-left: 50px;">├─ 计算可分配任务数 allocatableTaskCount</td>
  </tr>
  <tr>
    <td></td>
    <td style="padding-left: 50px;">└─ 计算碎片得分 fragmentScore</td>
  </tr>
  <tr>
    <td></td>
    <td style="padding-left: 25px;">Schedule() 执行调度</td>
  </tr>
  <tr>
    <td></td>
    <td style="padding-left: 50px;">├─ 第一次遍历：不使用预留资源</td>
  </tr>
  <tr>
    <td></td>
    <td style="padding-left: 50px;">├─ 遍历树选择最优节点 traverseTree()</td>
  </tr>
  <tr>
    <td></td>
    <td style="padding-left: 50px;">├─ 如果还有任务未分配，第二次遍历：使用预留资源</td>
  </tr>
  <tr>
    <td></td>
    <td style="padding-left: 50px;">└─ buildTaskTree() 构建任务树</td>
  </tr>
  <tr>
    <td style="padding-left: 25px;">4.</td>
    <td colspan="2">比较所有成功调度资源树的碎片得分，选择碎片最少（得分最低）的方案</td>
  </tr>
  <tr>
    <td style="padding-left: 25px;">5.</td>
    <td colspan="2">将选中节点缓存到 job.SuperPods</td>
  </tr>
  <tr>
    <td colspan="3"><br><strong>如果作业已完成节点选择：</strong></td>
  </tr>
  <tr>
    <td style="padding-left: 25px;"></td>
    <td colspan="2">为缓存中的每个节点打高分 (100000000 - rank)</td>
  </tr>
  <tr>
    <td style="padding-left: 25px;"></td>
    <td colspan="2">确保调度器绑定任务到预选的节点</td>
  </tr>
</table>
</div>

### 4. 树遍历调度策略

<div style="font-family: monospace; background-color: #f8f8f8; border: 1px solid #ccc; padding: 15px; border-radius: 5px; margin: 10px 0;">
<strong>traverseTree() 递归遍历调度逻辑：</strong>
<br><br>
<table style="border-collapse: collapse; border: none; width: 100%;">
  <tr>
    <td colspan="2">对于当前层级每个节点：</td>
  </tr>
  <tr>
    <td style="padding-left: 25px;"></td>
    <td>按碎片得分排序（默认升序，优先选择碎片少的节点）</td>
  </tr>
  <tr>
    <td style="padding-left: 25px;"></td>
    <td>尝试在该节点分配任务：</td>
  </tr>
  <tr>
    <td></td>
    <td style="padding-left: 25px;">如果节点有足够空闲子任务：</td>
  </tr>
  <tr>
    <td></td>
    <td style="padding-left: 50px;">分配任务 → 递归到下一层级继续分配</td>
  </tr>
  <tr>
    <td></td>
    <td style="padding-left: 50px;">如果所有任务分配完成，成功返回</td>
  </tr>
  <tr>
    <td></td>
    <td style="padding-left: 50px;">否则回滚分配</td>
  </tr>
  <tr>
    <td style="padding-left: 25px;"></td>
    <td>如果分配失败：返回失败，尝试下一个节点</td>
  </tr>
  <tr>
    <td colspan="2"><br><strong>两次遍历策略：</strong></td>
  </tr>
  <tr>
    <td style="padding-left: 25px;">•</td>
    <td>第一次：只使用非预留资源</td>
  </tr>
  <tr>
    <td style="padding-left: 25px;">•</td>
    <td>第二次：允许使用预留资源分配剩余任务</td>
  </tr>
</table>
</div>

### 5. 碎片得分计算

碎片得分用于评估资源分配后产生的资源碎片化程度：

```
fragmentScore = 碎片大小加权和

得分越低表示：
 - 碎片越少
 - 大块连续资源保留越完整
 - 后续作业更容易分配

调度器优先选择碎片得分最低的分配方案。
```

## 资源预留机制

### 配置方式

在资源层级配置中，可以为 **level1 层级** 预留一定数量的节点：

```json
"resource-level-config": {
  "default": {
    "level1": {
      "label": "rack-id",
      "reservedNode": 1  // 每个rack预留1个节点作为备用
    }
  }
}
```

> **注意**：目前预留节点配置仅在 `level1` 层级生效。

### 工作原理

预留节点的主要目的是为故障重调度提供备用节点资源：

1. 正常调度阶段不使用预留节点
2. 当发生节点故障需要重调度时，优先使用预留节点快速恢复故障任务
3. 保证在节点故障后仍有可用的同亲和性域节点，提高重调度成功率
4. 当正常资源无法满足新作业需求时，也允许使用预留资源完成调度

## 故障重调度

### 特性支持

多级调度集成了故障重调度能力，当检测到节点故障时：

<div style="font-family: monospace; background-color: #f8f8f8; border: 1px solid #ccc; padding: 15px; border-radius: 5px; margin: 10px 0;">
<table style="border-collapse: collapse; border: none; width: 100%;">
  <tr>
    <td colspan="2"><strong>Reschedule() 故障重调度</strong></td>
  </tr>
  <tr>
    <td style="padding-left: 25px;">1.</td>
    <td>findLargestFaultSubTree() 定位故障子树</td>
  </tr>
  <tr>
    <td style="padding-left: 25px;">2.</td>
    <td>从故障子树中取出所有需要重调度的任务</td>
  </tr>
  <tr>
    <td style="padding-left: 25px;">3.</td>
    <td>在新的资源树上重新调度这些任务</td>
  </tr>
  <tr>
    <td style="padding-left: 25px;">4.</td>
    <td>维持原有层级关系不变</td>
  </tr>
</table>
</div>

### Pod级细粒度重调度

支持单个Pod故障的细粒度重调度：

- 仅重调度故障Pod
- 保持同L1组内其他Pod位置不变
- 最小化故障影响范围

## 使用示例

### 示例1：8卡分布式训练，一级亲和性配置

```yaml
# PodGroup 配置
apiVersion: scheduling.volcano.sh/v1beta1
kind: PodGroup
metadata:
  name: 8card-training
  annotations:
    huawei.com/schedule_policy: "multilevel"
    huawei.com/affinity-config: "level1=4"
spec:
  minMembers: 8
  queue: default
```

**配置说明：**

- 总任务数：8个任务（每个任务1卡）
- 配置 `level1=4` → 每个level1分组包含4个节点，总共分成 8/4 = 2 个level1分组
- level1是最底层 → 同一个level1分组内的4个节点保证具有相同的level1标签值（如同一个机架）
- 调度器会尽量将每一组4个任务分配到同一个网络域（如机架），最小化跨域通信开销

### 示例2：16卡训练，二级亲和性配置

```yaml
apiVersion: scheduling.volcano.sh/v1beta1
kind: PodGroup
metadata:
  name: 16card-training
  annotations:
    huawei.com/schedule_policy: "multilevel"
    huawei.com/affinity-config: "level1=2,level2=4"
spec:
  minMembers: 16
  queue: default
```

**计算验证：**

- 总任务数 = 16
- level1=2 → 总共分成 16/2 = 8 个level1分组，每个level1分组2个节点同level1标签
- level2=4 → 总共分成 16/4 = 4 个level2分组，每个level2分组4个节点同level2标签
- 所有除法都整除 → 配置合法

**预期调度结果：**

- 所有同一个level1分组内的2个节点 → 同一机架（相同rack-id）→ 通信延迟最低
- 所有同一个level2分组内的4个节点 → 同一交换机（相同switch-id）→ 同一个交换机范围内
- 网络分层亲和性保障，最小化跨层级通信延迟

## 与Volcano社区网络拓扑感知调度的关系

| 特性     | Volcano社区网络拓扑感知 | ascend-volcano-plugin多级调度 |
| ------ | --------------- | ------------------------- |
| 网络抽象层级 | 支持三级拓扑          | 支持任意多级拓扑                  |
| NPU亲和性 | 通用GPU支持         | 深度集成Ascend NPU感知          |
| 资源预留   | 不支持             | 支持层级级资源预留                 |
| 碎片优化   | 不支持             | 内置碎片优化算法                  |
| 故障重调度  | 不支持             | 支持细粒度故障重调度                |
| 多级亲和性  | 仅支持亲和性优先级       | 支持严格多级分组亲和性保障             |

后续ascend-volcano-plugin插件计划兼容高版本volcano定义的hypernode类型k8s自定义资源用于获取网络拓扑。

## 优势与收益

1. **网络性能优化**：根据实际网络拓扑分配任务，最小化跨域通信，降低通信延迟
2. **资源利用率提升**：碎片优化算法减少资源碎片化，提高集群利用率
3. **高可用性**：资源预留机制+故障重调度，提高调度成功率，快速故障恢复
4. **灵活可扩展**：支持任意多级拓扑配置，适配不同数据中心网络架构
5. **兼容生态**：基于Volcano标准调度框架扩展，与其他Volcano特性良好兼容

## 约束与限制

1. **NPU要求**：当前版本要求每个任务占用整卡，不支持部分NPU分配
2. **配置连续性**：层级编号必须连续，不能跳号
3. **整除约束**：总任务数必须能被每个层级的分组大小整除，保证均匀分组
4. **标签要求**：所有节点必须正确配置拓扑层级标签，否则会被过滤

## 参考文献

- Volcano 网络拓扑感知调度文档：<https://volcano.sh/zh/docs/network_topology_aware_scheduling/>
- 主要代码位置：`component/ascend-for-volcano/internal/npu/policy/multilevelscheduling/`

