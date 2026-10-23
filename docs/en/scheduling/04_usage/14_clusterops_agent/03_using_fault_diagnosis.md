# Using Fault Diagnosis<a name="ZH-CN_TOPIC_00000026faultdiagnosis04"></a>

When a training or inference job becomes abnormal, you can trigger ClusterOps Agent with one command by job dimension using the `kubectl ascend_diag` command.

## Diagnosis Process<a name="sectionfaultdiagnosisprocess"></a>

1. Agent Core queries the central relationship table by job name to obtain all Pods of the job and the nodes where they reside.
2. Agent Core delivers collection instructions to Node Collector on each node, and collects job logs (plog, system logs, device logs, and so on) from the host according to the collection contract.
3. Node Collector locally calls `ascend-fd parse` to parse the logs, packages the parsing output, and reports it to Agent Core.
4. Agent Core aggregates the parsing output of each node, assembles it into a diagnosis input directory, calls `ascend-fd diag` to perform centralized diagnosis, and generates a diagnosis report (an LLM can be optionally integrated for intelligent summarization).

## Executing Diagnosis<a name="sectionfaultdiagnosisexecute"></a>

After a job becomes abnormal, run the following command to initiate diagnosis (the default namespace is `default`, and you can specify another one through `-n`):

```shell
kubectl ascend_diag --job job-x -n training
```

Where:

| Parameter | Description |
|---|---|
| `--job` | Job name, that is, the job CR name (mandatory) |
| `-n` / `--namespace` | Namespace where the job resides. The default value is `default`. |
| `--refresh` | Ignores the cache and forcibly re-executes diagnosis and refreshes the cache |
| `--json` | Outputs the complete JSON response |
| `--collect-manifest` | Writes the local collection contract to the cluster ConfigMap and then exits |
| `--agent-core-service` | Specifies the Agent Core service in the format `svc.ns:9700` (default: `agent-core.mindx-dl:9700`) |

**Diagnosis output example**:

```text
$ kubectl ascend_diag --job job-x -n training
The diag job starts. Please wait. Job id: [20260918163512389528_b1b293d4-1c70-4e11-916e-a9023da3660a], run log file is [ascend_faultdiag_36.log].
+-----------------------------+------------------+--------------------------------------------------------------------------------------------------------------------------------------------+
|                                                                                 Ascend-fd Fault-Diag Report                                                                                 |
+-----------------------------+------------------+--------------------------------------------------------------------------------------------------------------------------------------------+
|     Version Information     |       Type       | Version                                                                                                                                    |
+-----------------------------+------------------+--------------------------------------------------------------------------------------------------------------------------------------------+
|                             |    Fault-Diag    | 26.1.0                                                                                                                                     |
+-----------------------------+------------------+--------------------------------------------------------------------------------------------------------------------------------------------+
|   Root Cause Node Analysis  |       Type       | Description                                                                                                                                |
+-----------------------------+------------------+--------------------------------------------------------------------------------------------------------------------------------------------+
|                             | Root Cause Node  | ['worker0 device-7']                                                                                                                       |
|                             |     Symptom      | No timeout error logs are recorded in the Plog of any node. Nodes with error logs are suspected root cause nodes. Please investigate.      |
|                             | First Error Node | worker0 device-7: 2026-09-18-12:19:30.141986                                                                                               |
|                             | Last Error Node  | worker0 device-7: 2026-09-18-12:19:30.141986                                                                                               |
+-----------------------------+------------------+--------------------------------------------------------------------------------------------------------------------------------------------+
|     Fault Event Analysis    |       Type       | Description                                                                                                                                |
+-----------------------------+------------------+--------------------------------------------------------------------------------------------------------------------------------------------+
|  Suspected Root Cause Fault |   Status Code    | Comp_OS_Service_Container_04                                                                                                               |
|                             |  Fault Category  | Category:Software Component:HostOS Module:OS                                                                                               |
|                             |   Fault Device   | ['worker0']                                                                                                                                |
|                             |    Fault Name    | Container storage exception                                                                                                                |
|                             |    Fault Desc    | Docker storage latency.                                                                                                                    |
|                             |    Suggestion    | 1. Replace the storage disk and restart to create the container;                                                                           |
|                             |     Key Log      | Sep 18 12:19:24 localhost /usr/sbin/irqbalance[3091]: Cannot change IRQ 1145 affinity: No space left on device                             |
+-----------------------------+------------------+--------------------------------------------------------------------------------------------------------------------------------------------+
The diag job is complete.

Tip: to output the full diagnosis JSON, add --json after the command
```

You can use the `--json` parameter to view the complete JSON diagnosis response (including the collected Pod information, the diagnosis report, and so on):

```shell
kubectl ascend_diag --job job-x -n training --json
```

## Repeated Execution and Cache Prompt<a name="sectionfaultdiagnosisrepeat"></a>

When diagnosis is repeatedly executed for the same job **within 10 seconds**, the latest diagnosis result (cached data) is directly returned, and you are prompted to use `--refresh` to obtain real-time data:

```text
$ kubectl ascend_diag --job job-x -n training
Result from cache; run --refresh for the latest diagnosis
The diag job starts. Please wait. Job id: [20260918163512389528_b1b293d4-1c70-4e11-916e-a9023da3660a], run log file is [ascend_faultdiag_36.log].
+-----------------------------+------------------+--------------------------------------------------------------------------------------------------------------------------------------------+
|                                                                                 Ascend-fd Fault-Diag Report                                                                                 |
+-----------------------------+------------------+--------------------------------------------------------------------------------------------------------------------------------------------+
|     Version Information     |       Type       | Version                                                                                                                                    |
+-----------------------------+------------------+--------------------------------------------------------------------------------------------------------------------------------------------+
|                             |    Fault-Diag    | 26.1.0                                                                                                                                     |
+-----------------------------+------------------+--------------------------------------------------------------------------------------------------------------------------------------------+
|   Root Cause Node Analysis  |       Type       | Description                                                                                                                                |
+-----------------------------+------------------+--------------------------------------------------------------------------------------------------------------------------------------------+
|                             | Root Cause Node  | ['worker0 device-7']                                                                                                                       |
|                             |     Symptom      | No timeout error logs are recorded in the Plog of any node. Nodes with error logs are suspected root cause nodes. Please investigate.      |
|                             | First Error Node | worker0 device-7: 2026-09-18-12:19:30.141986                                                                                               |
|                             | Last Error Node  | worker0 device-7: 2026-09-18-12:19:30.141986                                                                                               |
+-----------------------------+------------------+--------------------------------------------------------------------------------------------------------------------------------------------+
|     Fault Event Analysis    |       Type       | Description                                                                                                                                |
+-----------------------------+------------------+--------------------------------------------------------------------------------------------------------------------------------------------+
|  Suspected Root Cause Fault |   Status Code    | Comp_OS_Service_Container_04                                                                                                               |
|                             |  Fault Category  | Category:Software Component:HostOS Module:OS                                                                                               |
|                             |   Fault Device   | ['worker0']                                                                                                                                |
|                             |    Fault Name    | Container storage exception                                                                                                                |
|                             |    Fault Desc    | Docker storage latency.                                                                                                                    |
|                             |    Suggestion    | 1. Replace the storage disk and restart to create the container;                                                                           |
|                             |     Key Log      | Sep 18 12:19:24 localhost /usr/sbin/irqbalance[3091]: Cannot change IRQ 1145 affinity: No space left on device                             |
+-----------------------------+------------------+--------------------------------------------------------------------------------------------------------------------------------------------+
The diag job is complete.

Tip: to output the full diagnosis JSON, add --json after the command
```

> [!NOTE]
>
> - If diagnosis is executed again for the same job while it is being diagnosed, the message `job=job-x is being diagnosed (started at HH:MM:SS), please do not run it again` is displayed, and collection is not repeated.
> - If the job does not exist (the job CR has been deleted and there is no Pod record of the job in `relcache`), `Training/inference task not found: job=job-x in ns=training` is directly returned and diagnosis is terminated. If the job CR has been deleted but there is still a pod record of the job in `relcache` (within the deletion TTL, and the logs are still retained on the node or shared disk), diagnosis continues.
> - For details about the diagnosis result cache, see the [Diagnosis Result Cache](../../06_api/20_clusterops_agent.md#sectionascendclusteropscache) chapter.

## Data Persistence and Space Reclamation<a name="sectionfaultdiagnosisstorage"></a>

Agent Core and Node Collector persist collection and diagnosis outputs by job dimension in their respective working directories. The specific paths are as follows:

| Component | Working Directory | Description |
|---|---|---|
| Agent Core | `/user/clusterops/agent-core` | Diagnosis input, diagnosis output, and result cache |
| Node Collector | `/user/clusterops/node-collector` | Host log collection and parsing output |

### Agent Core Persistent Content<a name="sectionfaultdiagnosisagentstorage"></a>

```text
/user/clusterops/agent-core/
├── {YYYYMMDD}/
│   └── {namespace}_{job}/
│       ├── parse-result-{job}-{node}.tar.gz          # Collection archive reported by each node
│       ├── diag-input/
│       │   └── {host_ip}/                             # Parsing output of each node, named by server IP (including server-info.json and so on)
│       └── diag-output/
│           └── fault_diag_result/
│               └── diag_report.json                   # Centralized diagnosis report
└── cache/
    └── {namespace}_{job}.json                         # Diagnosis result cache
```

- `parse-result-{job}-{node}.tar.gz`: collection archive reported by Node Collector by node.
- `diag-input/{host_ip}/`: `ascend-fd parse` output of each node, named by machine IP (falls back to `worker{N}` when `host_ip` is missing), assembled as the centralized diagnosis input.
- `diag-output/`: directory of the diagnosis report generated by `ascend-fd diag`.
- `cache/{namespace}_{job}.json`: diagnosis result cached after each successful diagnosis. Repeated diagnosis directly returns it, and `--refresh` can be used to forcibly refresh it.

### Node Collector Persistent Content<a name="sectionfaultdiagnosiscollectorstorage"></a>

```text
/user/clusterops/node-collector/
└── {YYYYMMDD}/
    └── {namespace}_{job}/
        ├── collect/                                   # Original logs collected from the host
        └── parse-output/                              # ascend-fd parse output
```

- `collect/`: original logs collected from the host according to the collection contract (plog, system logs, device logs, and so on).
- `parse-output/`: output parsed by the local `ascend-fd parse`, which is uploaded as the diagnosis input.

### Retention Policy<a name="sectionfaultdiagnosisretention"></a>

The working directory of both components is limited to **10 GB** by default.

- Agent Core triggers space reclamation after each diagnosis ends, and Node Collector triggers space reclamation after each collection upload ends.
- When the total occupied space of the working directory exceeds the threshold, directories and cache files are deleted from the oldest modification time to the newest until the total occupied space falls below the threshold.
- The deletion granularity is the job level: Agent Core deletes the entire `{YYYYMMDD}/{namespace}_{job}` directory and the corresponding `{namespace}_{job}.json` cache file; Node Collector deletes the entire `{YYYYMMDD}/{namespace}_{job}` directory.

> [!NOTE]
>
> - The diagnosis result cache and diagnosis outputs are all included in the 10 GB quota management of Agent Core, and the oldest ones are also deleted first when the quota is exceeded.
> - The cache directory is located at `/user/clusterops/agent-core/cache/` by default.
