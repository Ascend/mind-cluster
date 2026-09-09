# ascend-clusterops-agent

把 `ascend-faultdiag`（离线日志清洗 + 诊断）包装成集群诊断 Agent。用户只需说"job X 有问题"，Agent 自动完成**采集 → 清洗 → 诊断 → 出报告**，可用 `kubectl` 插件直接调用，也可走 HTTP/CLI。

## 模块介绍

- **agent-core**（Deployment，`:9700` HTTP / `:9710` gRPC）：诊断主控。经 relcache 查任务 → pod → `{node, poduid, rank}`，并发向各节点下发采集（TriggerCollect），聚合各节点上传的清洗产物，调用 `ascend-fd diag` 生成诊断报告，并按需落缓存。
- **node-collector**（DaemonSet，每 NPU 节点一份，`:9720` gRPC）：节点采集端。按 `collect_manifest.yaml` 的 entity（`env` 反查 / `paths` / `mount_keywords`）经 pathmap CM 匹配宿主路径，现场执行采集命令，本地 `ascend-fd parse` 清洗后回传产物。
- **kubectl 插件**（`kubectl-ascend_diag` / `kubectl-clusterops`）：面向用户的入口。`ascend_diag` 经 port-forward 调用 agent-core；`clusterops` 管理 LLM 配置。纯标准库实现，无需 pip 安装。

## 三方依赖

| 组件 | 依赖 | 用途 |
|---|---|---|
| agent-core (whl) | `grpcio` + `protobuf` | agent↔collector gRPC（含流式 tar 回传 / 协议桩） |
| | `kubernetes` | in-cluster APIServer 访问（pod informer / CR 状态） |
| | `fastapi` + `uvicorn` | 对外 HTTP 入口 `/diag` |
| | `pyyaml` | 读任务 CR 配置 CM（task_crds） |
| | `langgraph` + `langchain-openai` | 可选 LLM agent 包装层 + 报告总结（OpenAI 兼容端点） |
| node-collector (whl) | `grpcio` + `protobuf` | TriggerCollect / UploadResult（协议桩） |
| | `kubernetes` | 读 pathmap / 采集契约 CM |
| | `pyyaml` | 解析采集契约 |
| 镜像内 | `ascend-fd` (ascend-faultdiag) | 节点端 `parse` + 集中端 `diag`（非 pip 依赖，镜像安装） |
| | `dmidecode` | SMBIOS 采集（机型/SN） |

`requirements.txt` 为开发/调试依赖，与两个组件 whl 的依赖声明一致。

## 构建与交付件

```bash
bash build/build.sh        # 构建 whl + 打 zip 交付包
```

产物：

```bash
Ascend-mindxdl-ascend-clusterops-agent_<ver>_linux.zip
├── agent_core-*.whl          # 组件 whl (内嵌 common/diagproto)
├── node_collector-*.whl
├── agent-core.yaml           # 部署清单
├── node-collector.yaml
├── Dockerfile                # 合并镜像 (SERVICE 参数选角色)
├── collect_manifest.yaml     # 采集契约
└── kubectl-plugin/           # kubectl 插件
    ├── install.sh            # 一键安装 (装到 /usr/local/bin)
    ├── kubectl-ascend_diag   # kubectl ascend_diag
    └── kubectl-clusterops    # kubectl clusterops
```

交付 zip 内 `agent_core-*.whl` / `node_collector-*.whl` 与 `Dockerfile` 同层（纯 Python，`py3-none-any`，不依赖架构）；镜像构建以解压目录为上下文，Dockerfile 从同层 `COPY agent_core-*.whl node_collector-*.whl` 安装组件，运行时依赖（grpcio、protobuf、fastapi 等）由镜像内 `pip install` 按目标架构自动拉取。组件 wheel 跨平台，配合 `docker buildx --platform` 可在单机交叉构建 x86_64 / aarch64 镜像（镜像构建需联网）。

## 部署

**1. 构建并推送镜像**

镜像不内置 `ascend-fd`，需提前准备 `ascend_faultdiag-*.whl`：

- 从 release 下载 `Ascend-mindxdl-faultdiag_<ver>_linux-<arch>.zip`（`<arch>` 为 `x86_64` 或 `aarch64`）：
  `https://gitcode.com/Ascend/mind-cluster/releases/download/v<ver>/Ascend-mindxdl-faultdiag_<ver>_linux-<arch>.zip`
- 解压后仅提取其中 `ascend_faultdiag-*.whl`（zip 内其余 whl 不装），放到软件包解压目录（与 Dockerfile 同目录）。构建时 Dockerfile 按目标架构自动挑选匹配的 wheel（`linux_x86_64` / `linux_aarch64`，默认 x86_64），可同时放入两个架构的 wheel 以支持交叉构建；若匹配架构的 wheel 缺失，镜像构建会直接报错。

参考版本 v26.1.0 的直链：

- x86_64：`https://gitcode.com/Ascend/mind-cluster/releases/download/v26.1.0/Ascend-mindxdl-faultdiag_26.1.0_linux-x86_64.zip`
- aarch64：`https://gitcode.com/Ascend/mind-cluster/releases/download/v26.1.0/Ascend-mindxdl-faultdiag_26.1.0_linux-aarch64.zip`

```bash
# 在交付 zip 解压目录下执行 (上下文为解压目录, Dockerfile 与组件 whl 同层)
docker build -t <REGISTRY>/ascend-clusterops-agent:<TAG> .
# 交叉构建示例 (x86_64 主机构建 aarch64 镜像; 需联网拉取依赖; 解压目录需放 aarch64 的 ascend_faultdiag wheel)
docker buildx build --platform linux/arm64 -t <REGISTRY>/ascend-clusterops-agent:<TAG>-arm64 .
docker push <REGISTRY>/ascend-clusterops-agent:<TAG>
```

**2. 预创建宿主目录（每个部署节点）**

部署清单不再内置任何目录创建逻辑（日志/产物目录以 hostPath 挂载，需在宿主预先创建；否则 pod 会因 subPath 或属主权限无法启动/写盘）。

agent-core 运行用户为镜像内置 `hwMindX`（uid 9000），日志与产物目录需预先创建并归属 9000：

```bash
# 部署 agent-core 的节点:
mkdir -p /var/log/mindx-dl/agent-core && chown 9000:9000 /var/log/mindx-dl/agent-core
mkdir -p /user/clusterops/agent-core && chown 9000:9000 /user/clusterops/agent-core
```

node-collector 以 root（uid 0）运行，宿主目录只需存在、无需调整属主：

```bash
# 每个 NPU 节点:
mkdir -p /var/log/mindx-dl/node-collector
mkdir -p /user/clusterops/node-collector
```

**3. 部署 agent-core（Deployment）**

```bash
kubectl apply -f build/agent-core.yaml     # 含 SA/RBAC/Service/ConfigMap(task_crds)
```

**4. 部署 node-collector（DaemonSet，每 NPU 节点一份）**

采集契约已内置在镜像（`/home/hwMindX/collect_manifest.yaml`，随 node-collector 镜像装载），无需预先创建 ConfigMap：

```bash
kubectl apply -f build/node-collector.yaml
```

部署前按集群实际调整清单：

- **镜像 tag**：`spec.template.spec.containers[].image`（agent-core 与 node-collector 各一处）。
- **节点选择**：`nodeSelector.workerselector: dls-worker-node` 按实际 NPU 节点 label 调整。
- **任务 CR GVK**：agent-core 启动时经 K8s API 读取一次 ConfigMap `agent-core-task-crds`（cluster-system）的 `task_crds.yaml`（需对应 RBAC），默认含 AscendJob + InferServiceSet；其他任务类型（如 Volcano Job）需自行在该 ConfigMap 增补任务 GVK，并在 ClusterRole 中为该 CR 增补 `get` 权限（否则状态判定为 unknown，结果不缓存），然后重启 agent-core 生效。启动日志会打印当前支持检测的任务类型。
- **hostPath**：collector 需宿主 `/var/log`、`/var/log/ascend`、`/usr/local/Ascend`（全量，driver/toolkit 及 msnpureport 运行依赖）、`/usr/local/dcmi`、`/usr/bin/msnpureport`（单文件，device_log 采集命令）、`/bin/dmesg`（单文件，host_log 采集命令，挂在 `/usr/bin/dmesg`）；上一步已预创建的日志/产物目录不在此列。

**5. 配置 LLM（可选，不配则报告回退原始 JSON）**

```bash
kubectl clusterops --create-llm-config [--base-url U] [--model M]
# 清除:
kubectl clusterops --clear-llm-config
```

`llm-secret` 含 `api-key` / `base-url` / `model`，默认创建在 `mindx-dl`（可用环境变量 `ASCEND_CLUSTEROPS_NS` 覆盖，与 agent-core 部署命名空间一致即可）。agent-core 启动时经 k8s watch 实时监控 `llm-secret`（对应 RBAC `agent-core-llm` 已在 agent-core.yaml 中），**更新/清除后立即生效，无需重启 agent-core**；`get_llm()` 只读内存缓存、不再每次调 k8s API，secret 不存在时回退纯确定性诊断。

**6. 安装 kubectl 插件**

插件把诊断 Agent 暴露成 `kubectl ascend_diag --job <jobname>` 命令，纯标准库实现、无需 pip 安装。

解压交付包后，在 `kubectl-plugin/` 目录执行（需 sudo，装到 `/usr/local/bin`）：

```bash
bash kubectl-plugin/install.sh   # 安装并验证 kubectl ascend_diag / kubectl clusterops 可用
```

也可手动安装：

```bash
sudo cp kubectl-plugin/kubectl-ascend_diag /usr/local/bin/kubectl-ascend_diag
sudo cp kubectl-plugin/kubectl-clusterops /usr/local/bin/kubectl-clusterops
kubectl ascend_diag --help
```

## 使用

**kubectl 插件（默认经 port-forward 调部署的 agent，复用本机 kubeconfig 鉴权）**

```bash
kubectl ascend_diag --job job-x                  # 按任务 CR 名诊断 (默认 ns=default)
kubectl ascend_diag --job job-x -n training
kubectl ascend_diag --job job-x --json           # 完整 JSON 响应
kubectl ascend_diag --job job-x --refresh        # 忽略缓存强制重跑并刷新缓存
kubectl ascend_diag --collect-manifest ./collect_manifest.yaml   # 将本地采集契约写入集群 manifest ConfigMap
```

**更新采集契约（collect_manifest.yaml）**

采集契约默认来自镜像内置 `/home/hwMindX/collect_manifest.yaml`；如需现场调整，直接用插件把本地 collect_manifest.yaml 写入集群 ConfigMap（cluster-system/collect-manifest）覆盖（无需重打镜像/重启）：

```bash
kubectl ascend_diag --collect-manifest ./collect_manifest.yaml
# 写入 ConfigMap collect-manifest (cluster-system), 需当前 kubeconfig 有 create/apply 权限
```

node-collector 会 watch 该 CM（`get/list/watch` 权限，见 node-collector.yaml），一旦更新立即生效并把最新契约（`manifest_version` 与 `entities` 列表）打印到日志，便于确认覆盖内容；CM 删除后回退到镜像内置文件。

agent-core 不再读取采集契约（pathmap 只存任务 pod 的挂载对与字面 env 值），因此更新契约只影响 node-collector 侧的采集匹配，不影响 pathmap 与已缓存的诊断结果。`env` 实体仍由 node-collector 解析：env 值由 agent-core 在 pod 存活时记录进 pathmap CM（TTL 内保留，pod 删除后仍可用），collector 从 pathmap 读取 env 值（容器内路径）→ 经挂载对反查宿主路径 → 采集；env 未记录或反查失败时回退 `mount_keywords`。

**HTTP API（agent-core :9700）**

```bash
curl -X POST localhost:9700/diag -H 'Content-Type: application/json' \
     -d '{"job":"job-x","namespace":"default","refresh":false}'
```

**CLI（本机直接诊断）**

```bash
python -m agent_core.main --job job-x -n default [--refresh]
```

**kubectl 插件工作原理与访问模型**

默认模式：用本机 kubeconfig 起一个临时 `kubectl port-forward` 到 agent-core Service（`mindx-dl.agent-core`），POST `/diag`，拿报告后关闭 port-forward。

- 复用用户已有 kubeconfig 鉴权，**无需单独给 agent 暴露入口或发 key**。
- agent Service 名可通过 `--agent-core-service svc.ns:9700` 或环境变量 `ASCEND_CLUSTEROPS_AGENT_SVC` 覆盖。
- k8s 鉴权走用户 kubeconfig（port-forward 用 kubeconfig 的权限）；agent 内部访问 k8s 走 pod 的 ServiceAccount + RBAC（见 `build/agent-core.yaml` 内 RBAC 段），无手工 key。
- LLM key：agent pod 的 `llm-secret`（用户经 `kubectl clusterops --create-llm-config` 配置，标准 OpenAI 兼容接口），用户侧不接触。

当前 CLI 是单次诊断。多轮 follow-up（"根因节点 plog 展开""是不是 ECC"）后续在 Web UI 形态做；CLI 保持单次脚本化。

## 关键行为说明

- **重复诊断防护**：同一 job 正在诊断中再次执行会提示"正在诊断中, 请勿重复执行"。
- **产物目录**：agent-core 的采集/诊断产物按天与任务组织在 `{AGENT_WORK_ROOT}/{YYYYMMDD}/{ns}_{job}/`（默认 `/user/clusterops/agent-core/`）：tar 留存 `parse-result-{job}-{node}.tar.gz`、解压目录 `diag-input/worker-{node}/`（assemble 重命名为 `worker0..N`）、诊断输出 `diag-output/fault_diag_result/diag_report.json`；node-collector 的本地采集产物同样按 `{COLLECTOR_WORK_ROOT}/{YYYYMMDD}/{ns}_{job}/collect|parse-output` 组织，不同任务/命名空间互不覆盖。
- **结果缓存**：仅任务处于停止状态时缓存到 `{AGENT_WORK_ROOT}/cache/<ns>_<job>.json`（默认 `/user/clusterops/agent-core/cache`，与 `{YYYYMMDD}` 产物目录同级、跨天可命中，可用 `AGENT_CACHE_DIR` 覆盖为固定目录）。停止状态判定（大小写不敏感）：任务 CR 终态 `JobSucceeded`/`JobFailed`/`Succeeded`/`Failed`/`Stopped`/`Complete(d)`/`Terminated`/`Aborted`，或 CR 不存在（404，任务已删除）；CR 状态查询失败（RBAC/网络等）判定 unknown，此时不缓存（但不阻塞诊断）；CR 状态读不到或滞后时，若任务 pod 全部到达终态（`Succeeded`/`Failed`）同样可缓存。running 状态不缓存。命中缓存会提示"(结果来自缓存)"。
- **任务不存在**：job 不存在时立即返回"训练/推理任务不存在"并终止，不再执行诊断。
- **日志落盘**：agent-core → `/var/log/mindx-dl/agent-core/agent-core.log`；node-collector → `/var/log/mindx-dl/node-collector/node-collector.log`（hostPath 挂载，宿主持久；10MB × 10 轮转）。

## 本地调试 / 验证

```bash
pip install -r requirements.txt
# 主流程不依赖 LLM
export KUBECONFIG=~/.kube/config
python -m agent_core.main --job job-x -n default
# 或
uvicorn agent_core.main:app --port 9700
```

单元测试（无需真集群）：

```bash
python -m pytest tests/ -q
```
