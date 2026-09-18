# 配置日志采集<a name="ZH-CN_TOPIC_00000026faultdiagnosis03"></a>

集群运维Agent通过**采集契约**（collect_manifest.yaml）定义采集哪些日志。采集契约内置4类日志实体，分别是 `process_log`（CANN plog）、`dl_log`（MindCluster组件日志）、`device_log`（Device侧日志）和 `host_log`（主机OS日志）。其中plog是最常用的故障定位日志，且配置最复杂，本章节重点说明。

## 采集契约示例<a name="sectionfaultdiagnosismanifestexample"></a>

采集契约的核心是 `entities` 列表，每个实体定义一类日志的采集方式。完整示例如下：

```yaml
entities:
  - name: process_log            # CANN plog, 容器挂载文件日志
    env: ASCEND_PROCESS_LOG_PATH  # pod spec 显式配置时 agent-core 记录并随采集指令下发; node-collector 读取反查宿主路径
    mount_keywords: [plog]        # 兜底: 匹配挂载对/宿主子目录/共享盘静态挂载, 用 pod IP+任务标识过滤

  - name: dl_log                 # MindCluster 组件日志 (宿主机路径, 直接读取, 保留目录结构)
    paths:
      - "/var/log/mindx-dl/devicePlugin"
      - "/var/log/mindx-dl/noded"
      - "/var/log/ascend-docker-runtime"
      - "/var/log/mindx-dl/volcano-scheduler"
      - "/var/log/mindx-dl/volcano-controller"
      - "/var/log/mindx-dl/npu-exporter"

  - name: device_log             # Device 侧日志 (现场命令生成)
    commands:
      - "/usr/bin/msnpureport --docker"

  - name: host_log               # 主机 OS 日志 (paths 为节点宿主路径, 直接读取)
    commands:
      - "dmesg -T | tail -n 100000 > dmesg"
      - "dmidecode > dmidecode.txt"
    paths:
      - "/var/log/messages*"
      - "/var/log/sysmonitor.log"
```

## 字段说明<a name="sectionfaultdiagnosismanifestfields"></a>

每个实体通过以下字段描述采集方式：

| 字段 | 字段含义 |
|---|---|
| `name` | 实体的名称，标识一类日志（如 `process_log`、`dl_log`） |
| `env` | 环境变量名，其值指向任务Pod内容器的日志目录（如 `ASCEND_PROCESS_LOG_PATH`） |
| `paths` | 日志路径列表，指定需要采集的日志文件或目录 |
| `commands` | 现场命令列表，指定在采集节点上执行、用于生成日志文件的命令 |
| `mount_keywords` | 兜底关键词列表，按关键词匹配任务Pod的挂载对来定位日志；无挂载对命中时，按关键词匹配宿主机子目录或共享盘（/mnt/shared-storage）上的日志目录 |

## 采集契约处理逻辑<a name="sectionfaultdiagnosismanifestlogic"></a>

同一个实体可以只配置其中一个字段，也可以同时配置多个字段。当配置了多个字段时，Node Collector按以下优先级决定实际执行的采集方式：

| 优先级 | 字段 | 处理逻辑 |
|---|---|---|
| 1 | `env` | 环境变量。Agent Core在Pod存活时将该env值和全部hostPath挂载对记入中心挂载关系表，并在采集指令（TriggerCollect）中随Pod列表下发，Node Collector读取env值（容器内路径）后，经挂载对反查宿主机路径进行采集。env采集成功时，不再执行 `paths`、`commands`、`mount_keywords`；env未记录或反查失败时，继续按低优先级字段处理 |
| 2 | `paths` / `commands` | 同级字段，env未配置或采集失败时均会执行。`paths` 优先按容器内路径匹配任务Pod的挂载对，反查宿主机路径后复制（保留目录结构）；若任务Pod无匹配挂载对，则直接将配置的路径作为宿主机路径读取。`dl_log`、`host_log` 实体是特例，其 `paths` 始终作为宿主机路径直接读取，不经过任务Pod挂载对匹配。`commands` 在临时目录执行命令生成文件，并复制到采集目录；整个命令串必须精确匹配白名单，否则拒绝执行 |
| 3 | `mount_keywords` | 关键词兜底。仅当未配置 `paths` 时执行；按关键词匹配任务Pod的全部挂载对（宿主机/容器路径含关键词即命中），无直接命中时扫描挂载宿主路径的子目录；仍未命中时，扫描静态挂载的共享盘（`/mnt/shared-storage`，协议不限：NFS/CephFS/云盘等，由部署方挂载），并用Pod IP与任务标识（`MINDX_TASK_ID`）过滤出属于该Pod的日志子目录 |

> [!NOTE]
>
> - 采集字段的优先级为 `env` > `paths` = `commands` > `mount_keywords`。配置了高优先级字段且采集成功时，不再执行低优先级字段；`paths` 与 `commands` 同级，env未配置或采集失败时两者均会执行。
> - `mount_keywords` 的共享盘扫描不依赖固定目录结构。目录名**只要包含任一关键词即可**（`mount_keywords` 为子串匹配，默认 `plog` 可命中 `plog`/`plogs`/`plog_xxx` 等），不限于示例中的 `plogs`；同时用任务标识与Pod IP两级动态段过滤（支持 `alllogs/<task_id>/plogs/<ip>` 与 `alllogs/<ip>/plogs/<task_id>` 两种布局，动态段顺序不固定），只采集属于该任务、该Pod的日志，不会采到同一共享盘下其他任务的日志。任务Pod删除（kubelet卷挂载点随之消失）后，共享盘挂载点仍挂在Node Collector上，仍可完成采集。

## 各类日志实体配置<a name="sectionfaultdiagnosistentities"></a>

### process_log（CANN plog）<a name="sectionfaultdiagnosisplog"></a>

CANN plog是训练或推理进程的运行日志，包含昇腾算子、通信等关键错误信息，是故障定位的首选日志。采集契约中 `process_log` 实体的默认配置如下：

```yaml
entities:
  - name: process_log            # CANN plog, 容器挂载文件日志
    env: ASCEND_PROCESS_LOG_PATH  # pod spec 显式配置时 agent-core 记录并随采集指令下发; node-collector 读取反查宿主路径
    mount_keywords: [plog]        # 兜底: 匹配挂载对/宿主子目录/共享盘静态挂载, 用 MINDX_TASK_ID+pod IP 过滤
```

plog实体配置了 `env` 和 `mount_keywords` 两个字段，采集时按以下逻辑找到plog日志：

1. **env（首选）**：任务Pod在Pod spec中显式配置了 `ASCEND_PROCESS_LOG_PATH` 时使用。Agent Core将该Pod的字面env值和全部hostPath挂载对记录进中心挂载关系表（TTL内保留，Pod删除后仍可用），并在TriggerCollect采集指令中随Pod列表下发给对应节点的Node Collector。Node Collector读取该env值（容器内路径），经挂载对反查宿主机路径，再从宿主机读取plog日志。
2. **mount_keywords（兜底）**：env未记录或反查失败（例如 `ASCEND_PROCESS_LOG_PATH` 由启动脚本export、不在Pod spec中）时，Node Collector按 `mount_keywords`（`plog`）依次匹配：任务Pod的挂载对、挂载宿主路径的子目录、以及静态挂载的共享盘（`/mnt/shared-storage`）。共享盘扫描用Pod IP与任务标识（`MINDX_TASK_ID`）过滤出属于该Pod的plog子目录。目录名只要包含关键词即可（子串匹配，`plog` 可命中 `plog`/`plogs`/`plog_xxx` 等），不依赖固定目录结构，兼容 `alllogs/<task_id>/plogs/<ip>` 与 `alllogs/<ip>/plogs/<task_id>` 两种布局（动态段顺序不固定），也不会采到同一共享盘下其他任务的日志。

#### 配置plog输出<a name="sectionfaultdiagnosisplogtask"></a>

plog能否被采集取决于plog输出目录的挂载方式，对应两种场景：

- **hostPath方式（走env反查）**：任务Pod将plog目录通过hostPath挂载到宿主机，并在Pod spec中设置 `ASCEND_PROCESS_LOG_PATH`。Agent Core的pathmap记录hostPath挂载对，Node Collector经env反查宿主机路径采集。
- **共享盘方式（走mount_keywords的共享盘静态挂载）**：任务Pod将plog目录所在路径通过共享盘卷挂载（协议不限：NFS/CephFS/云盘等，如 `/job/code`），`ASCEND_PROCESS_LOG_PATH` 由启动脚本动态注入（如 `/job/code/alllogs/$MINDX_TASK_ID/plogs/$XDL_IP`）。部署时把共享盘静态挂载到Node Collector的 `/mnt/shared-storage`，Node Collector的 `mount_keywords` 兜底在挂载点下用 `MINDX_TASK_ID` 和Pod IP过滤出该Pod的plog目录。共享盘由kubelet/CSI挂载，Node Collector直接读取，无需mount命令，任务Pod删除后仍可采集。

以hostPath方式为例，AscendJob任务Pod模板关键片段如下：

```yaml
apiVersion: mindxdl.gitee.com/v1
kind: AscendJob
metadata:
  name: job-x
  namespace: training
spec:
  tasks:
  - template:
      spec:
        containers:
        - name: main
          image: <训练镜像>
          env:
          - name: ASCEND_PROCESS_LOG_PATH   # 指定容器内plog日志目录
            value: /home/ascend/log
          volumeMounts:
          - name: plog
            mountPath: /home/ascend/log     # 与ASCEND_PROCESS_LOG_PATH一致
        volumes:
        - name: plog
          hostPath:
            path: /data/plog               # plog宿主机目录 (需预先在节点上创建)
            type: Directory
```

该任务YAML与采集契约的对应关系如下：

1. `env` 中的 `ASCEND_PROCESS_LOG_PATH=/home/ascend/log` ，Node Collector据此反查宿主机路径 `/data/plog`，完成plog实体采集。
2. `volumeMounts.mountPath`（`/home/ascend/log`）与 `ASCEND_PROCESS_LOG_PATH` 必须一致，否则env反查无法命中挂载对。
3. `volumes.hostPath.path`（`/data/plog`）是plog在宿主机上的实际目录，Node Collector最终从该目录读取日志。

> [!CAUTION]
> 任务Pod的plog目录所在宿主机路径必须存在，否则Pod将调度失败。宿主机目录需要用户预先创建，部署清单不会自动创建目录。

以共享盘方式为例，任务Pod将plog所在路径通过共享盘卷挂载（协议不限：NFS/CephFS/云盘等；`ASCEND_PROCESS_LOG_PATH` 由启动脚本按 `MINDX_TASK_ID`/`XDL_IP` 动态拼接，共享盘上的plog目录路径不固定，两种布局均可匹配）：

```yaml
apiVersion: mindxdl.gitee.com/v1
kind: AscendJob
metadata:
  name: job-x
  namespace: training
spec:
  tasks:
  - template:
      spec:
        containers:
        - name: main
          image: <训练镜像>
          env:
          - name: MINDX_TASK_ID            # 任务标识 (动态 id, Agent Core记录, 共享盘扫描过滤用)
            value: task-42
          # ASCEND_PROCESS_LOG_PATH 由启动脚本动态注入, 不在 Pod spec 中:
          #   export ASCEND_PROCESS_LOG_PATH=/job/code/alllogs/$MINDX_TASK_ID/plogs/$XDL_IP
          volumeMounts:
          - name: code
            mountPath: /job/code          # 共享盘卷容器挂载路径 (mount_path)
        volumes:
        - name: code
          nfs:                             # 共享盘卷, 协议不限 (NFS/CephFS/云盘等)
            server: 10.0.0.9              # NFS 服务器 (server)
            path: /export/code            # NFS 导出路径 (path)
```

共享盘上的plog目录路径含两级**动态 id**，顺序不固定，两种布局均能命中（plog 目录名不限于 `plogs`，凡含关键词 `plog` 的目录均可，如 `plog`/`plog_xxx`）：

| 动态 id | 来源 | 用途 |
|---|---|---|
| `MINDX_TASK_ID`（如 `task-42`） | 任务Pod环境变量 | 过滤同盘上其他任务的日志目录 |
| pod IP（如 `10.0.0.5`） | 任务Pod状态 | 过滤同任务其他Pod的日志子目录 |

```text
# 布局一: 先 task_id 后 pod_ip
/job/code/alllogs/task-42/plogs/10.0.0.5/   # 命中: 路径含 task_id, plogs 下子目录名 == pod_ip
/job/code/alllogs/task-99/plogs/9.9.9.9/    # 跳过: 其他任务(task-99)的plog

# 布局二: 先 pod_ip 后 task_id
/job/code/alllogs/10.0.0.5/plogs/task-42/   # 命中: 路径含 pod_ip, 直接取整个plogs目录
/job/code/alllogs/9.9.9.9/plogs/task-99/    # 跳过: 其他Pod(9.9.9.9)的plog

# 目录名变体 (仅示例 plogs; 只要含关键词 plog 即可, 如 plog / plog_2026):
/job/code/alllogs/task-42/plog/10.0.0.5/    # 同样命中
```

该任务YAML与采集契约的对应关系如下：

1. Agent Core记录该Pod的字面env值（含 `MINDX_TASK_ID`）与pod IP，随采集指令下发给Node Collector；pathmap不记录共享盘卷的挂载信息（server/path/mount_path）。
2. `ASCEND_PROCESS_LOG_PATH` 由启动脚本注入（不在Pod spec中），env反查失败后走 `mount_keywords` 兜底。
3. Node Collector在静态挂载的共享盘（`/mnt/shared-storage`）下按关键词 `plog` 匹配plog目录，再用 `MINDX_TASK_ID`（`task-42`）与Pod IP（如 `10.0.0.5`）过滤出本Pod的plog目录后采集。

任务Pod删除后，kubelet卷挂载点随之消失，但共享盘挂载点仍挂在Node Collector上，仍可完成采集，不影响诊断。

#### 将plog宿主机目录挂载到Node Collector<a name="sectionfaultdiagnosisplogmount"></a>

Node Collector是独立于任务Pod的DaemonSet，运行在每个NPU节点上，它通过hostPath挂载来读取宿主机日志。默认情况下，Node Collector已挂载以下宿主机目录：

- `/var/log`（系统日志）
- `/usr/local/Ascend`（Ascend全量目录）
- `/usr/local/dcmi`
- `/usr/bin/msnpureport`（单文件）

> [!IMPORTANT]
> **共享盘必须静态挂载到Node Collector的 `/mnt/shared-storage`。** 若任务plog等共享日志落在共享盘上（NFS/CephFS/云盘等，协议不限），必须把该共享盘挂载到Node Collector的 `/mnt/shared-storage`，否则共享盘上的plog无法采集。共享盘由kubelet/CSI挂载（见node-collector.yaml中 `shared-storage` 注释示例，按集群实际协议填 nfs 或 csi 参数并取消注释启用），Node Collector直接读取挂载点（任务Pod删除后仍可采集）。
> [!CAUTION]
> **plog的宿主机地址必须挂载到node-collector.yaml中。** 若任务Pod的plog宿主机目录不在Node Collector已有挂载范围内（例如上例中的 `/data/plog`），Node Collector将无法从宿主机读取该plog目录，导致plog采集失败。此时必须在node-collector.yaml中为该宿主机目录增补hostPath挂载。

以挂载 `/data/plog` 为例，在 `node-collector.yaml` 的 `volumeMounts` 和 `volumes` 中分别增补以下配置：

```yaml
        volumeMounts:
        - name: plog
          mountPath: /data/plog            # 与宿主机路径一致
          readOnly: true
      volumes:
      - name: plog
        hostPath:
          path: /data/plog                # plog宿主机目录 (需预先在节点上创建)
          type: Directory
```

修改后重新应用清单并滚动更新DaemonSet：

```shell
kubectl apply -f node-collector.yaml
kubectl rollout restart daemonset node-collector -n mindx-dl
```

### dl_log（MindCluster组件日志）<a name="sectionfaultdiagnosisdllog"></a>

`dl_log` 实体采集MindCluster各组件的运行日志（如Ascend Device Plugin、NodeD、Ascend Docker Runtime、Volcano、NPU Exporter），默认配置如下：

```yaml
  - name: dl_log                 # MindCluster 组件日志 (宿主机路径, 直接读取, 保留目录结构)
    paths:
      - "/var/log/mindx-dl/devicePlugin"
      - "/var/log/mindx-dl/noded"
      - "/var/log/ascend-docker-runtime"
      - "/var/log/mindx-dl/volcano-scheduler"
      - "/var/log/mindx-dl/volcano-controller"
      - "/var/log/mindx-dl/npu-exporter"
```

`dl_log` 实体配置了 `paths` 字段，Node Collector将配置的路径**直接作为宿主机路径读取**并复制文件（保留目录结构），不经过任务Pod挂载对匹配。

### device_log（Device侧日志）<a name="sectionfaultdiagnosisdevicelog"></a>

`device_log` 实体采集昇腾Device侧日志，通过现场命令生成，默认配置如下：

```yaml
  - name: device_log             # Device 侧日志 (现场命令生成)
    commands:
      - "/usr/bin/msnpureport --docker"
```

`device_log` 实体配置了 `commands` 字段，Node Collector在临时目录执行白名单命令生成日志文件，并复制到采集目录。该命令串必须精确匹配白名单，篡改的命令串会被拒绝执行。

### host_log（主机OS日志）<a name="sectionfaultdiagnosishostlog"></a>

`host_log` 实体采集节点主机的OS日志（如内核日志、系统监控日志），默认配置如下：

```yaml
  - name: host_log               # 主机 OS 日志 (paths 为节点宿主路径, 直接读取)
    commands:
      - "dmesg -T | tail -n 100000 > dmesg"
      - "dmidecode > dmidecode.txt"
    paths:
      - "/var/log/messages*"
      - "/var/log/sysmonitor.log"
```

`host_log` 实体直接读取节点宿主路径（`paths`）并执行现场命令（`commands`）。

## 更新采集契约<a name="sectionfaultdiagnosismanifest"></a>

采集契约（collect_manifest.yaml）默认内置在镜像内，路径为 `/home/hwMindX/collect_manifest.yaml`。如需现场调整采集项（新增实体、修改env、调整关键词等），无需重新制作镜像或重启Node Collector，直接通过Kubectl Plugin将采集契约写入集群ConfigMap即可。

```shell
kubectl ascend_diag --collect-manifest ./collect_manifest.yaml
```

命令执行成功后输出如下：

```text
collect manifest written to ConfigMap collect-manifest/cluster-system: ./collect_manifest.yaml
subsequent diagnostics will use the updated collect manifest (node-collector re-reads this CM on every collection, no restart needed)
```

> [!NOTE]
>
> - 采集契约ConfigMap位于cluster-system命名空间，执行命令需要当前kubeconfig对该ConfigMap具备create/apply权限。
> - Node Collector每次采集时都会重新读取ConfigMap `collect-manifest`，因此更新后立即生效，无需重启。

## 注意事项<a name="sectionfaultdiagnosisnotes"></a>

- **不同任务输出到同一目录可能导致诊断不准**：多个任务若将日志输出到同一个宿主机目录（如多个任务共用同一个plog挂载目录），Node Collector会按任务Pod的挂载对采集该目录下全部日志，不同任务的日志会混杂在一起，可能导致诊断结果定位到错误的任务。建议每个任务使用独立的日志目录。
- **采集不到日志不影响诊断成功**：若任务没有产生日志输出（如进程未启动、日志路径未挂载）或Node Collector采集不到日志，诊断流程仍会正常执行并生成报告，只是缺少相应日志会导致诊断准确性下降。可通过诊断报告或Node Collector日志确认日志采集情况。
