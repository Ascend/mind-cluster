# RFC: 支持重建的Pod优先调度回原节点

**状态 (Status):** Accepted
**作者 (Authors):** @shepherd-cheung
**创建日期 (Created):** 2026-05-01
**更新日期 (Updated):** 2026-06-17

# 1. 概述

## 1.1 简介

当Pod因驱逐（Preempt/Reclaim）、故障重调度或缩容后再扩容而被重新调度时，如果Pod被分配到**不同节点**，需重新拉取全套容器镜像，导致任务启动延迟5~20分钟。如果Pod能回到**原节点**，利用节点上已有的容器镜像缓存，启动时间可降至秒级。

本特性通过ascend-volcano-plugin在调度打分阶段为原节点增加偏好加分，引导重建Pod优先回到此前运行的节点。

**适用范围：**

| 任务类型              | 机制 | 说明                                    |
|-------------------|------|---------------------------------------|
| 普通 NPU 任务         | `addPreferPreviousNodeScore` 节点三分类+统治性得分 | 本特性主要作用对象                             |
| 超节点任务（super pod）/多级调度任务（multi-level） | 策略层内置 `VerifyCachedSuperPods` + `selectNodeFromCache` | 所有原节点存在直接回原节点；其他场景走原有逻辑               |
| VNPU 任务           | 与普通 NPU 任务逻辑一致 | 自动适用                                  |

## 1.2 动机

- **驱逐后恢复：** 被驱逐Pod回原节点复用镜像缓存（15~40 GB），避免重新拉取
- **缩容再扩容：** 推理/训练任务调整副本数后，新Pod回原节点秒级启动
- **故障重调度：** 健康Pod回原节点，故障Pod主动迁移到新节点脱离故障域

不做此提案的影响：Pod随机分配到新节点 → 重新拉取15~40 GB镜像 → 延迟5~20分钟。

## 1.3 目标

**目标：**

- Pod重新调度时优先回到原节点（在满足亲和性约束的前提下）
- 纯内存缓存，依托 `ScheduleHandler` 生命周期跨Session保持
- 对普通NPU任务通过 `maxScore + 100` 统治性得分实现最高优先级
- 超节点/多级调度任务由各自策略层内置逻辑处理

**非目标：**

- 不修改Volcano框架核心逻辑（全部改动在ascend-volcano-plugin内）
- 不作为硬性的node affinity绑定（原节点不满足条件时允许调度到其他节点）
- 不修改超节点/多级调度策略的强制约束逻辑
- 不支持有点调度回原芯片

---

# 2. 方案设计

## 2.1 核心流程

```text
NPUAllocateFunc (Pod分配到节点)
  │
  └─ 写入内存缓存: {ownerUID → rankIndex → RankNodeEntry{Node, Previous}}
       │   └─ Previous 保留旧值作为回滚锚点
       ▼
  缓存保存在 ScheduleHandler.AffinityCache 中，跨调度Session存活
       │
       ▼  下次调度Session
  InitNPUSession → initAffinityCache()
       │
       ├─ 冷启动（调度器重启后首次）：从当前运行中的Pod赋值缓存
       ├─ 热启动：RefreshOwner 刷新时间戳 + EvictExpired 清理过期条目
       └─ 为每个活跃Job预加载 PrefNodeMap（rankIndex→nodeName 快照）
       │
       ▼
  BatchNodeOrderFn (打分阶段)
  │
  ├─ 1. 策略层强制打分（ScoreBestNPUNodes）
  │    ├─ 超节点策略: SuperPod选择、Rack亲和性、HCCS环完整性
  │    ├─ 多级调度策略: 资源树匹配、拓扑约束
  │    └─ 普通策略: NPU数量、节点标签匹配
  │    打分基准: scoreForNode ≈ 100,000,000
  │
  ├─ 2. addPreferPreviousNodeScore ← 仅对普通Job生效
  │    ├─ 超节点/多级调度Job → 直接跳过（策略层自行处理）
  │    └─ 普通Job → 节点三分类 + maxScore + 100
  │
  ├─ 3. FaultHandle.ScoreBestNPUNodes
  │    ├─ 亚健康节点减分（-1）
  │    └─ 故障节点减分（-64）
  │
  └─ 4. scoreWeight乘权 (×100) → 最终排序
```

### 超节点/多级调度任务回原节点路径

超节点/多级调度策略在 `ScoreBestNPUNodes` 中内置了回原节点逻辑：

```text
ScoreBestNPUNodes (superpod/frame.go)
  │
  ├─ VerifyCachedSuperPods(nodes, PreferPreviousNode)
  │    ├─ 所有原节点健康可用 → JobReadyTag=true → 跳过完整调度
  │    │    └─ defer: selectNodeFromCache()
  │    │         ├─ scoreNodeBatchForReadyJob: sMap[node] = scoreForNode - rankId
  │    │         └─ scoreNodeForReadyJob:    sMap[node] += scoreForNode
  │    └─ 缓存验证失败 → 降级为完整调度 selectSuperPodForJob()
  │
  └─ addPreferPreviousNodeScore() 被跳过 (IsSuperPodJob=true)
```

## 2.2 普通NPU任务：节点三分类与统治性得分

`addPreferPreviousNodeScore` 将打分节点划分为三个类别，按Pod的故障状态采用不同优先级：

### 节点分类

| 类别 | 定义 | 非故障Pod优先级 | 故障Pod优先级 |
|------|------|---------------|-------------|
| **selfNode** | 当前Pod自己的原节点（`PrefNodeMap[myRank]`） | **第1优先** | **第2优先**（兜底） |
| **otherNodes** | 同Job中其他Pod都未使用过的节点 | **第2优先** | **第1优先**（脱离故障域） |
| **peerNodes** | 同Job中其他Pod使用过的节点（非本Pod） | **第3优先** | —（不参与加分） |

### 得分规则

```text
非故障Pod:
  1. selfNode 在scoreMap中 → 设为 maxScore + 100 ✓ 回原节点
  2. selfNode 不在scoreMap中 → 选otherNodes中最高分节点 → 设为 maxScore + 100
  3. 无otherNodes → 选peerNodes中最高分节点 → 设为 maxScore + 100

故障Pod:
  1. 选otherNodes中最高分节点 → 设为 maxScore + 100 ✓ 脱离故障域
  2. 无otherNodes → selfNode 在scoreMap中 → 设为 maxScore + 100（兜底）
  3. peerNodes 不参与加分
```

统治性得分设计：原节点得分 = `maxScore + 100`，而非固定分值加成。这确保在策略层打分结果相同的候选节点中，原节点确定性胜出。故障减分（-64）在偏好加分之后执行，硬故障仍可使原节点被超越。

### 任务排序

`TaskOrderFn` 在打分前对同Job内任务排序：Pod在缓存中有原节点记录 → **优先调度**，先占住原节点；无记录 → 后调度。

## 2.3 缓存设计

### Key设计

```text
Owner UID: "abc-123-def" (PodGroup对应Controller的UID)
  ├─ rankIndex "0" → RankNodeEntry{Node: "node-gpu-05", Previous: "node-gpu-02"}
  ├─ rankIndex "1" → RankNodeEntry{Node: "node-gpu-06", Previous: ""}
  └─ rankIndex "2" → RankNodeEntry{Node: "node-gpu-07", Previous: "node-gpu-03"}
```

- **第一层Key：** `ownerUID` — PodGroup的Controller UID（如Deployment/StatefulSet），PodGroup重建后UID变化也不影响映射
- **第二层Key：** `rankIndex` — 来自Pod的 `hccl/rankIndex` Annotation，Fallback到 `task.Index`
- **`Previous`字段：** 回滚锚点。重新分配时旧Node变为Previous；`RollbackAssignment`时Node恢复为Previous

### 数据结构

```go
type RankNodeEntry struct {
    Node     string // 当前节点分配
    Previous string // 上一次分配（回滚锚点），首次分配为空
}

type PodNodeAffinityCache struct {
    // key: ownerUID, value: rankIndex → RankNodeEntry
    OwnerToRankNodes map[string]map[string]*RankNodeEntry
    // key: ownerUID, value: last update timestamp (Unix seconds)
    UpdateTime map[string]int64
}
```

无锁设计 — Volcano调度框架在单goroutine中顺序处理任务。

### 缓存生命周期

| 事件 | 行为 | 原因 |
|------|------|------|
| Pod分配到任意节点 | `RecordAssignment`: 旧Node→Previous，新节点→Node | 更新位置，保留旧值供回滚 |
| Pod分配被回滚（状态=Pending） | `RollbackAssignment`: Node←Previous，或删除条目 | 分配未实际生效 |
| Pod被驱逐/缩容删除（状态=Releasing） | **保留**缓存 | 回原节点的关键 |
| 调度器冷启动 | 从当前运行中Pod的分配信息重建缓存 | 恢复重启前的映射 |
| 每个Session，Owner的PG仍存在 | `RefreshOwner` 刷新时间戳 | 标记仍活跃 |
| 缓存超过TTL（72小时） | `EvictExpired` 清理 | 兜底清理 |

### 持久化

纯内存缓存，保存在 `ScheduleHandler.AffinityCache` 中，通过插件框架持有的 `ScheduleHandler` 指针跨Session存活。不使用ConfigMap或外部存储。调度器重启后通过冷启动从运行中Pod恢复。

## 2.4 配置

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: volcano-scheduler-configmap
  namespace: volcano-system
data:
  volcano-scheduler.conf: |
    actions: "enqueue, allocate, backfill"
    tiers:
    - plugins:
      - name: priority
      - name: gang
      - name: volcano-npu_v6.0.RC1_linux-x86_64    # ascend-volcano 插件
    - plugins:
      - name: drf
      - name: predicates
      - name: proportion
      - name: nodeorder
      - name: binpack
    configurations:
      - name: init-params
        arguments: {"grace-over-time":"900", "prefer-previous-node": "true"}
```

| 参数 | 位置 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| `prefer-previous-node` | `configurations.init-params.arguments` | string | `"false"` | 是否开启回原节点功能 |

加分值固定为 `maxScore + 100`，TTL固定为72小时，均为代码硬编码。

## 2.5 故障场景安全保障

FaultHandle在加分之后执行减分（亚健康-1，故障-64）。配合节点三分类中故障Pod优先选otherNodes的策略：

```text
场景1：Pod健康，原节点非故障
  非故障Pod → 优先回selfNode → maxScore + 100 ✓

场景2：Pod为故障Pod，原节点恰好是故障节点
  故障Pod → 优先选otherNodes → 脱离故障域 ✓
  selfNode仅在无otherNodes时兜底（-64惩罚已生效）

场景3：Pod为故障Pod，原节点健康
  故障Pod → 仍优先选otherNodes → 主动迁移
  设计意图：故障Pod关联的故障可能在同节点上复现
```

## 2.6 与超节点策略的交互

`addPreferPreviousNodeScore()` 对 `IsSuperPodJob()` 为true的Job直接跳过。超节点Job的回原节点由 `ScoreBestNPUNodes` 内部的 `VerifyCachedSuperPods` → `selectNodeFromCache` 链路处理：

- **非故障场景：** `VerifyCachedSuperPods` 验证所有原节点健康 → 跳过完整调度 → `selectNodeFromCache` 直接指派原节点
- **故障重调度：** 缓存验证失败 → 降级到已有重调度逻辑（`selectNodeFromOriginVSuperPod` 等）

## 2.7 与多级调度策略的交互

`addPreferPreviousNodeScore()` 对 `IsMultiLevelJob()` 为true的Job直接跳过。多级调度Job的回原节点由 `ScoreBestNPUNodes` 内部链路处理：

- **非故障场景：** `tryUseCachedSuperPods` → `isCachedSuperPodsValid` 验证缓存 → `selectNodeFromCache` 直接指派原节点
- **重调度场景：** `resolveSuperPodsForReschedule` 从两个来源获取历史数据：
  - FaultJob缓存（始终检查）
  - Job历史分配（`PreferPreviousNode` 门控）→ `rescheduleWithSuperPods` 渐进式重调度：锁定健康原节点 → 逐级扩大故障范围 → 最终回退到完整Schedule

---

# 3. 文件结构

| 类别 | 文件 | 说明 |
|------|------|------|
| 缓存 | `common/cache/previous_node.go` | PodNodeAffinityCache 数据结构+所有方法 |
| 缓存测试 | `common/cache/previous_node_test.go` | 单元测试 |
| 公共 | `plugin/type.go` | ClusterCache.AffinityCache、ScheduleHandler.AffinityCache、SchedulerJob.PrefNodeMap |
| 公共 | `plugin/factory.go` | InitNPUSession→initAffinityCache; BatchNodeOrderFn→addPreferPreviousNodeScore; TaskOrderFn; 配置读取 |
| 公共 | `plugin/task.go` | NPUAllocateFunc写入; NPUDeallocateFunc回滚/保留 |
| 公共 | `plugin/const.go` | preferPreviousNodeKey、defaultPreferPreviousScore常量 |
| 超节点 | `internal/npu/ascend910/ascend910a3/superpod/frame.go` | VerifyCachedSuperPods→selectNodeFromCache（内置回原节点） |
| 多级调度 | `internal/npu/policy/multilevelscheduling/frame.go` | tryUseCachedSuperPods→selectNodeFromCache; resolveSuperPodsForReschedule |

---

# 4. 可靠性

- **纯内存缓存，无外部依赖：** 不依赖ConfigMap，避免读写失败的降级逻辑
- **重启自愈：** 调度器重启后首次Session从运行中Pod重建缓存
- **TTL兜底清理：** 72小时TTL，每次Session巡检清理过期Owner
- **回滚安全：** `RankNodeEntry.Previous` 保留旧分配，分配失败时自动恢复
- **锁无关设计：** 依赖Volcano调度框架的单goroutine顺序执行模型

---

# 5. 限制和注意事项

- 回原节点是软偏好，不保证100%命中
- 原节点不满足强制约束时（超节点亲和性、资源不足等），不会回到原节点
- 超节点和多级调度Job的策略层自行处理回原节点，不经过 `addPreferPreviousNodeScore`
- 缓存TTL固定72小时，不可配置
- `rankIndex` 依赖Pod的 `hccl/rankIndex` Annotation，缺失时Fallback到 `task.Index`
- 多调度器实例场景下，缓存为各实例独立的内存副本
- 故障Pod主动避免回原节点（优先选择otherNodes），这是设计行为而非缺陷
