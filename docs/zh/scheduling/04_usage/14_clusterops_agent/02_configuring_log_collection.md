# 配置日志采集<a name="ZH-CN_TOPIC_00000026faultdiagnosis03"></a>

集群运维Agent通过**采集契约**（collect_manifest.yaml）定义采集哪些日志。采集契约内置4类日志实体，分别是 `process_log`（CANN plog）、`dl_log`（MindCluster组件日志）、`device_log`（Device侧日志）和 `host_log`（主机OS日志）。其中plog是最常用的故障定位日志，且配置最复杂，本章节重点说明。

## 采集契约示例<a name="sectionfaultdiagnosismanifestexample"></a>

采集契约的核心是 `entities` 列表，每个实体定义一类日志的采集方式。完整示例如下：

```yaml
entities:
  - name: process_log            # CANN plog, 容器挂载文件日志
    env: ASCEND_PROCESS_LOG_PATH  # agent-core 记录进 pathmap; node-collector 读取反查宿主路径; 未记录/反查失败回退 mount_keywords
    mount_keywords: [plog]        # 兜底: 按 plog 关键词匹配全部挂载对/宿主子目录

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
| `mount_keywords` | 兜底关键词列表，按关键词匹配任务Pod的挂载对来定位日志 |

## 采集契约处理逻辑<a name="sectionfaultdiagnosismanifestlogic"></a>

同一个实体可以只配置其中一个字段，也可以同时配置多个字段。当配置了多个字段时，Node Collector按以下优先级决定实际执行的采集方式：

| 优先级 | 字段 | 处理逻辑 |
|---|---|---|
| 1 | `env` | 环境变量。Agent Core在Pod存活时将该env值记录进pathmap ConfigMap，Node Collector读取env值（容器内路径）后，经挂载对反查宿主机路径进行采集。env采集成功时，不再执行 `paths`、`commands`、`mount_keywords`；env未记录或反查失败时，继续按低优先级字段处理 |
| 2 | `paths` / `commands` | 同级字段，env未配置或采集失败时均会执行。`paths` 优先按容器内路径匹配任务Pod的挂载对，反查宿主机路径后复制（保留目录结构）；若任务Pod无匹配挂载对，则直接将配置的路径作为宿主机路径读取。`dl_log`、`host_log` 实体是特例，其 `paths` 始终作为宿主机路径直接读取，不经过任务Pod挂载对匹配。`commands` 在临时目录执行命令生成文件，并复制到采集目录；整个命令串必须精确匹配白名单，否则拒绝执行 |
| 3 | `mount_keywords` | 关键词兜底。仅当未配置 `paths` 时执行；按关键词匹配任务Pod的全部挂载对（宿主机/容器路径含关键词即命中），无直接命中时扫描挂载宿主路径的子目录 |

> [!NOTE]
>
> - 采集字段的优先级为 `env` > `paths` = `commands` > `mount_keywords`。配置了高优先级字段且采集成功时，不再执行低优先级字段；`paths` 与 `commands` 同级，env未配置或采集失败时两者均会执行。

## 各类日志实体配置<a name="sectionfaultdiagnosistentities"></a>

### process_log（CANN plog）<a name="sectionfaultdiagnosisplog"></a>

CANN plog是训练或推理进程的运行日志，包含昇腾算子、通信等关键错误信息，是故障定位的首选日志。采集契约中 `process_log` 实体的默认配置如下：

```yaml
entities:
  - name: process_log            # CANN plog, 容器挂载文件日志
    env: ASCEND_PROCESS_LOG_PATH  # agent-core 记录进 pathmap; node-collector 读取反查宿主路径; 未记录/反查失败回退 mount_keywords
    mount_keywords: [plog]        # 兜底: 按 plog 关键词匹配全部挂载对/宿主子目录
```

plog实体配置了 `env` 和 `mount_keywords` 两个字段，采集时按以下逻辑找到plog日志：

1. **env（首选）**：任务Pod存活时，Agent Core将该Pod的字面env值（`ASCEND_PROCESS_LOG_PATH`）和全部hostPath挂载对记录到ConfigMap `clusterops-pathmap`（TTL内保留，Pod删除后仍可用）。Node Collector采集时，从pathmap读取该env值（容器内路径），经挂载对反查宿主机路径，再从宿主机读取plog日志。
2. **mount_keywords（兜底）**：若env值未记录（如Pod已删除且TTL过期）或反查失败，Node Collector按 `mount_keywords`（`plog`）匹配Pod的全部挂载对，或扫描挂载宿主路径的子目录进行兜底采集。

#### 配置plog输出<a name="sectionfaultdiagnosisplogtask"></a>

为保证plog可被采集，任务Pod需要同时满足以下两点：

1. 设置环境变量 `ASCEND_PROCESS_LOG_PATH`，指向容器内的plog日志目录。
2. 将plog日志目录所在路径通过 **hostPath** 挂载到宿主机。Agent Core的pathmap只记录hostPath挂载对，若plog目录未挂载到宿主机，Node Collector将无法从宿主机读取。

以AscendJob任务为例，任务Pod模板关键片段如下：

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

1. `env` 中的 `ASCEND_PROCESS_LOG_PATH=/home/ascend/log` 会被Agent Core记录进pathmap，Node Collector据此反查宿主机路径 `/data/plog`，完成plog实体采集。
2. `volumeMounts.mountPath`（`/home/ascend/log`）与 `ASCEND_PROCESS_LOG_PATH` 必须一致，否则env反查无法命中挂载对。
3. `volumes.hostPath.path`（`/data/plog`）是plog在宿主机上的实际目录，Node Collector最终从该目录读取日志。

> [!CAUTION]
> 任务Pod的plog目录所在宿主机路径必须存在，否则Pod将调度失败。宿主机目录需要用户预先创建，部署清单不会自动创建目录。

#### 将plog宿主机目录挂载到Node Collector<a name="sectionfaultdiagnosisplogmount"></a>

Node Collector是独立于任务Pod的DaemonSet，运行在每个NPU节点上，它通过hostPath挂载来读取宿主机日志。默认情况下，Node Collector已挂载以下宿主机目录：

- `/var/log`（系统日志）
- `/var/log/ascend`
- `/usr/local/Ascend`（Ascend全量目录）
- `/usr/local/dcmi`
- `/usr/bin/msnpureport`（单文件）

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

`dl_log` 实体配置了 `paths` 字段，Node Collector将配置的路径**直接作为宿主机路径读取**并复制文件（保留目录结构），不经过任务Pod的pathmap挂载对匹配。

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
> - 更新采集契约只影响Node Collector侧的采集匹配，不影响Agent Core的pathmap及已缓存的诊断结果。

## 注意事项<a name="sectionfaultdiagnosisnotes"></a>

- **不同任务输出到同一目录可能导致诊断不准**：多个任务若将日志输出到同一个宿主机目录（如多个任务共用同一个plog挂载目录），Node Collector会按任务Pod的挂载对采集该目录下全部日志，不同任务的日志会混杂在一起，可能导致诊断结果定位到错误的任务。建议每个任务使用独立的日志目录。
- **采集不到日志不影响诊断成功**：若任务没有产生日志输出（如进程未启动、日志路径未挂载）或Node Collector采集不到日志，诊断流程仍会正常执行并生成报告，只是缺少相应日志会导致诊断准确性下降。可通过诊断报告或Node Collector日志确认日志采集情况。
